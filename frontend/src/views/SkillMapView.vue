<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { skillApi } from '@/api/client'
import type { SkillMap, SkillNode } from '@/types'

const route = useRoute()
const router = useRouter()
const skill = ref<SkillMap | null>(null)
const loading = ref(true)
const error = ref('')

const courseId = Number(route.params.id)

onMounted(async () => {
  try {
    skill.value = (await skillApi.map(courseId)).data
  } catch (e: any) {
    error.value = e.response?.data?.error || '加载能力图谱失败'
  } finally {
    loading.value = false
  }
})

// 按单元分组
const units = computed(() => {
  const map = new Map<string, SkillNode[]>()
  for (const n of skill.value?.nodes || []) {
    const key = `${n.unit_id}`
    if (!map.has(key)) map.set(key, [])
    map.get(key)!.push(n)
  }
  return Array.from(map.entries()).map(([key, nodes]) => ({
    unitId: Number(key),
    unitTitle: nodes[0]?.unit_title || '',
    nodes,
  }))
})

function statusLabel(status: string) {
  const map: Record<string, string> = { 未学: '未学', 学习中: '学习中', 已掌握: '已掌握', 薄弱: '薄弱' }
  return map[status] || status
}

function statusClass(status: string) {
  const map: Record<string, string> = { 未学: 'not-started', 学习中: 'learning', 已掌握: 'mastered', 薄弱: 'weak' }
  return map[status] || ''
}

function masteryClass(v: number) {
  if (v >= 70) return 'good'
  if (v >= 40) return 'mid'
  return 'low'
}

function goLesson(node: SkillNode) {
  if (node.completed) {
    router.push(`/lesson/${node.lesson_id}`)
  }
}
</script>

<template>
  <div class="container skill-page">
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else-if="skill">
      <div class="skill-header">
        <h1 class="skill-title">🧠 能力图谱</h1>
        <p class="skill-sub">{{ skill.course.emoji }} {{ skill.course.title }} · 已开始知识点平均掌握度</p>
        <div class="overall-ring">
          <div class="overall-circle" :class="masteryClass(skill.overall)">
            <span class="overall-value">{{ skill.overall }}<small>%</small></span>
          </div>
        </div>
      </div>

      <div class="legend">
        <span class="legend-item"><i class="dot mastered"></i>已掌握</span>
        <span class="legend-item"><i class="dot learning"></i>学习中</span>
        <span class="legend-item"><i class="dot weak"></i>薄弱</span>
        <span class="legend-item"><i class="dot not-started"></i>未学</span>
      </div>

      <div v-for="u in units" :key="u.unitId" class="unit-section">
        <h2 class="unit-title">📚 {{ u.unitTitle }}</h2>
        <div class="node-list">
          <div
            v-for="n in u.nodes"
            :key="n.lesson_id"
            class="skill-node"
            :class="statusClass(n.status)"
            @click="goLesson(n)"
          >
            <div class="node-main">
              <span class="node-icon">{{ n.icon }}</span>
              <div class="node-info">
                <div class="node-title">{{ n.lesson_title }}</div>
                <div class="node-track">
                  <div class="node-fill" :class="masteryClass(n.mastery)" :style="{ width: n.mastery + '%' }"></div>
                </div>
              </div>
              <span class="node-status" :class="statusClass(n.status)">{{ statusLabel(n.status) }}</span>
              <span class="node-mastery">{{ n.mastery }}%</span>
            </div>
          </div>
        </div>
      </div>

      <button class="btn-ghost back-btn" @click="router.push(`/course/${courseId}`)">← 返回学习路径</button>
    </div>
  </div>
</template>

<style scoped>
.skill-page { padding-top: 24px; padding-bottom: 40px; }

.skill-header { text-align: center; margin-bottom: 20px; }
.skill-title { font-size: 26px; font-weight: 900; margin-bottom: 6px; }
.skill-sub { color: var(--text-light); font-size: 14px; margin-bottom: 16px; }

.overall-ring { display: flex; justify-content: center; }
.overall-circle {
  width: 110px;
  height: 110px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 34px;
  font-weight: 900;
  color: white;
}
.overall-circle.good { background: conic-gradient(#22c55e 0 100%); box-shadow: 0 0 0 8px rgba(34,197,94,0.15); }
.overall-circle.mid { background: conic-gradient(#f59e0b 0 100%); box-shadow: 0 0 0 8px rgba(245,158,11,0.15); }
.overall-circle.low { background: conic-gradient(#ef4444 0 100%); box-shadow: 0 0 0 8px rgba(239,68,68,0.15); }
.overall-value small { font-size: 16px; }

.legend { display: flex; justify-content: center; gap: 16px; margin-bottom: 24px; flex-wrap: wrap; }
.legend-item { display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--text-light); }
.dot { width: 12px; height: 12px; border-radius: 50%; display: inline-block; }
.dot.mastered { background: #22c55e; }
.dot.learning { background: #3b82f6; }
.dot.weak { background: #ef4444; }
.dot.not-started { background: #e5e7eb; }

.unit-section { margin-bottom: 24px; }
.unit-title { font-size: 17px; font-weight: 800; margin-bottom: 12px; }
.node-list { display: flex; flex-direction: column; gap: 8px; }

.skill-node {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.15s;
}
.skill-node:hover { transform: translateX(4px); }
.skill-node.mastered { border-color: #86efac; }
.skill-node.learning { border-color: #93c5fd; }
.skill-node.weak { border-color: #fca5a5; }

.node-main { display: flex; align-items: center; gap: 12px; }
.node-icon { font-size: 22px; }
.node-info { flex: 1; min-width: 0; }
.node-title { font-size: 14px; font-weight: 700; margin-bottom: 6px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.node-track { height: 10px; background: var(--bg-gray); border-radius: 5px; overflow: hidden; }
.node-fill { height: 100%; border-radius: 5px; transition: width 0.5s; }
.node-fill.good { background: #22c55e; }
.node-fill.mid { background: #3b82f6; }
.node-fill.low { background: #ef4444; }

.node-status { font-size: 12px; font-weight: 800; padding: 3px 10px; border-radius: 12px; flex-shrink: 0; }
.node-status.mastered { background: #dcfce7; color: #15803d; }
.node-status.learning { background: #dbeafe; color: #1d4ed8; }
.node-status.weak { background: #fee2e2; color: #b91c1c; }
.node-status.not-started { background: var(--bg-gray); color: var(--text-light); }
.node-mastery { width: 52px; text-align: right; font-weight: 900; color: var(--text); flex-shrink: 0; }

.loading, .error { text-align: center; padding: 40px; color: var(--text-light); }
.back-btn { margin-top: 24px; width: 100%; }
</style>
