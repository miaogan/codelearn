<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { predictionApi } from '@/api/client'
import type { ExamPrediction } from '@/types'

const route = useRoute()
const router = useRouter()
const courseId = Number(route.params.id)
const loading = ref(true)
const prediction = ref<ExamPrediction | null>(null)

onMounted(async () => {
  try {
    const res = await predictionApi.get(courseId)
    prediction.value = res.data
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
})

const levelMeta = computed(() => {
  const p = prediction.value
  if (!p) return { label: '', color: '', bg: '' }
  if (p.level === 'high') return { label: '状态良好', color: '#10b981', bg: '#ecfdf5' }
  if (p.level === 'low') return { label: '风险较高', color: '#ef4444', bg: '#fef2f2' }
  return { label: '状态一般', color: '#f59e0b', bg: '#fffbeb' }
})

const R = 54
const CIRC = 2 * Math.PI * R

function goPrepTask(t: { type: string; target_id?: number }) {
  if (t.type === 'exam' && t.target_id) {
    router.push(`/unit-exam/${t.target_id}`)
  } else if (t.type === 'lesson' && t.target_id) {
    router.push(`/lesson/${t.target_id}`)
  }
}
</script>

<template>
  <div class="prediction-page">
    <div class="header">
      <router-link :to="`/course/${courseId}`" class="back-btn">← 返回课程</router-link>
      <h1>🎯 考试预测</h1>
    </div>

    <div v-if="loading" class="loading">计算中...</div>

    <template v-else-if="prediction">
      <div class="hero" :style="{ background: levelMeta.bg, borderColor: levelMeta.color }">
        <div class="gauge-wrap">
          <svg class="gauge" viewBox="0 0 120 120">
            <circle class="gauge-bg" cx="60" cy="60" :r="R" />
            <circle
              class="gauge-fill"
              cx="60"
              cy="60"
              :r="R"
              :stroke="levelMeta.color"
              :stroke-dasharray="CIRC"
              :stroke-dashoffset="CIRC * (1 - prediction.probability / 100)"
            />
          </svg>
          <div class="gauge-center">
            <span class="gauge-num">{{ prediction.probability }}</span>
            <span class="gauge-unit">%</span>
          </div>
        </div>
        <div class="hero-info">
          <div class="hero-title">{{ prediction.course_title }} · 认证考试</div>
          <div class="hero-level" :style="{ color: levelMeta.color }">{{ levelMeta.label }}</div>
          <div class="hero-eligible" :class="{ ok: prediction.eligible }">
            {{ prediction.eligible ? '✅ 已具备考试资格' : `⏳ 已通过 ${prediction.units_passed}/${prediction.units_total} 个单元` }}
          </div>
        </div>
      </div>

      <div class="section">
        <h2 class="section-title">📊 预测因子</h2>
        <div class="factor-list">
          <div v-for="f in prediction.factors" :key="f.label" class="factor-item">
            <div class="factor-head">
              <span class="factor-label">{{ f.label }}</span>
              <span class="factor-weight">{{ f.weight }}%</span>
              <span class="factor-score">{{ f.score }}</span>
            </div>
            <div class="factor-bar">
              <div class="factor-fill" :style="{ width: f.score + '%' }"></div>
            </div>
            <div class="factor-detail">{{ f.detail }}</div>
          </div>
        </div>
      </div>

      <div class="section advice" :style="{ borderColor: levelMeta.color, background: levelMeta.bg }">
        <h2 class="section-title">💡 备考建议</h2>
        <p class="advice-text">{{ prediction.advice }}</p>
      </div>

      <div class="section" v-if="prediction.prep_tasks.length > 0">
        <h2 class="section-title">📋 备考任务</h2>
        <div class="task-list">
          <div
            v-for="(t, i) in prediction.prep_tasks"
            :key="i"
            class="task-item"
            @click="goPrepTask(t)"
          >
            <span class="task-icon">{{ t.type === 'exam' ? '📝' : t.type === 'lesson' ? '📚' : t.type === 'review' ? '🧠' : '🛠️' }}</span>
            <div class="task-body">
              <div class="task-title">{{ t.title }}</div>
              <div class="task-reason">{{ t.reason }}</div>
            </div>
            <span class="task-arrow">→</span>
          </div>
        </div>
      </div>

      <div class="actions">
        <router-link :to="`/course/${courseId}`" class="btn-primary">继续学习</router-link>
        <router-link :to="`/course/${courseId}/cert-exam`" v-if="prediction.eligible" class="btn-secondary">去考试</router-link>
      </div>
    </template>

    <div v-else class="loading">暂无预测数据</div>
  </div>
</template>

<style scoped>
.prediction-page { max-width: 700px; margin: 0 auto; padding: 20px; }
.header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.back-btn { color: var(--text-light); text-decoration: none; font-size: 16px; }
.header h1 { font-size: 28px; font-weight: 900; }
.loading { text-align: center; padding: 60px 20px; color: var(--text-light); }

.hero {
  display: flex;
  align-items: center;
  gap: 24px;
  border: 2px solid;
  border-radius: var(--radius);
  padding: 24px;
  margin-bottom: 16px;
}
.gauge-wrap { position: relative; flex-shrink: 0; }
.gauge { width: 120px; height: 120px; transform: rotate(-90deg); }
.gauge-bg { fill: none; stroke: rgba(0,0,0,0.08); stroke-width: 10; }
.gauge-fill { fill: none; stroke-width: 10; stroke-linecap: round; transition: stroke-dashoffset 0.8s; }
.gauge-center {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: baseline;
  justify-content: center;
  transform: translateY(35px);
}
.gauge-num { font-size: 30px; font-weight: 900; }
.gauge-unit { font-size: 14px; font-weight: 700; color: var(--text-light); }
.hero-info { flex: 1; min-width: 0; }
.hero-title { font-size: 17px; font-weight: 800; margin-bottom: 4px; }
.hero-level { font-size: 15px; font-weight: 800; margin-bottom: 4px; }
.hero-eligible { font-size: 12px; color: var(--text-light); }
.hero-eligible.ok { color: #10b981; font-weight: 700; }

.section {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 20px;
  margin-bottom: 16px;
}
.section-title { font-size: 16px; font-weight: 800; margin-bottom: 14px; }

.factor-list { display: flex; flex-direction: column; gap: 16px; }
.factor-head { display: flex; align-items: center; gap: 10px; margin-bottom: 6px; }
.factor-label { flex: 1; font-size: 14px; font-weight: 700; }
.factor-weight { font-size: 11px; color: var(--text-light); }
.factor-score { font-size: 15px; font-weight: 900; color: var(--primary); min-width: 32px; text-align: right; }
.factor-bar { height: 10px; background: var(--bg-gray); border-radius: 5px; overflow: hidden; margin-bottom: 6px; }
.factor-fill { height: 100%; background: linear-gradient(90deg, var(--primary), #6366f1); border-radius: 5px; transition: width 0.6s; }
.factor-detail { font-size: 12px; color: var(--text-light); }

.advice-text { font-size: 14px; line-height: 1.7; }

.task-list { display: flex; flex-direction: column; gap: 10px; }
.task-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: var(--bg-gray);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s;
}
.task-item:hover { border: 2px solid var(--primary); padding: 10px 12px; }
.task-icon { font-size: 20px; }
.task-body { flex: 1; min-width: 0; }
.task-title { font-size: 14px; font-weight: 700; }
.task-reason { font-size: 12px; color: var(--text-light); }
.task-arrow { color: var(--text-light); }

.actions { display: flex; gap: 12px; margin-top: 8px; }
.actions a { flex: 1; text-align: center; text-decoration: none; }
.btn-primary, .btn-secondary {
  padding: 12px 20px;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 800;
}
.btn-primary { background: var(--primary); color: white; box-shadow: 0 4px 0 var(--primary-dark); }
.btn-secondary { background: white; color: var(--text); border: 2px solid var(--border); box-shadow: 0 4px 0 #e5e5e5; }

@media (max-width: 480px) {
  .hero { flex-direction: column; text-align: center; }
}
</style>
