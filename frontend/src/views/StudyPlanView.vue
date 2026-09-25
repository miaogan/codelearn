<script setup lang="ts">
import { ref } from 'vue'
import { studyPlanApi, courseApi } from '@/api/client'
import type { StudyPlan, Course } from '@/types'

const courses = ref<Course[]>([])
const goal = ref('')
const language = ref('python')
const weeks = ref(4)
const plan = ref<StudyPlan | null>(null)
const loading = ref(false)
const error = ref('')

async function loadCourses() {
  try {
    const res = await courseApi.list()
    courses.value = res.data.courses
    if (courses.value.length > 0) {
      language.value = courses.value[0].language === 'py' ? 'python' : courses.value[0].language
    }
  } catch (e) {
    // ignore
  }
}
loadCourses()

async function generate() {
  loading.value = true
  error.value = ''
  plan.value = null
  try {
    const res = await studyPlanApi.generate({
      goal: goal.value,
      language: language.value,
      weeks: weeks.value,
    })
    plan.value = res.data
  } catch (e) {
    error.value = '生成失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="plan-page">
    <div class="header">
      <router-link to="/" class="back-btn">← 返回</router-link>
      <h1>🤖 AI 学习计划</h1>
    </div>

    <div class="section form-section">
      <h2 class="section-title">定制你的学习路线</h2>
      <label class="field">
        <span class="field-label">学习目标</span>
        <input v-model="goal" class="input" placeholder="例如：掌握 Python 基础并完成一个小项目" />
      </label>
      <label class="field">
        <span class="field-label">编程语言</span>
        <select v-model="language" class="input">
          <option value="python">Python</option>
          <option value="go">Go</option>
        </select>
      </label>
      <label class="field">
        <span class="field-label">计划周期（周）</span>
        <div class="week-options">
          <button
            v-for="w in [2, 4, 6, 8, 12]"
            :key="w"
            class="week-option"
            :class="{ active: weeks === w }"
            @click="weeks = w"
          >{{ w }} 周</button>
        </div>
      </label>
      <button class="btn-primary" :disabled="loading" @click="generate">
        {{ loading ? '生成中...' : '🚀 生成学习计划' }}
      </button>
      <p v-if="error" class="error">{{ error }}</p>
      <p v-if="!loading && !plan && !error" class="tip">
        AI 将根据目标拆解每周主题与任务；未配置 AI 时自动使用系统模板。
      </p>
    </div>

    <template v-if="plan">
      <div class="plan-hero">
        <div class="plan-goal">{{ plan.goal }}</div>
        <div class="plan-meta">{{ plan.language }} · {{ plan.total_weeks }} 周</div>
      </div>

      <div class="section summary">
        <h2 class="section-title">📌 整体规划</h2>
        <p class="summary-text">{{ plan.summary }}</p>
      </div>

      <div class="week-list">
        <div v-for="w in plan.weeks" :key="w.week" class="week-card">
          <div class="week-head">
            <span class="week-badge">第 {{ w.week }} 周</span>
            <span class="week-focus">{{ w.focus }}</span>
          </div>
          <ul class="task-list">
            <li v-for="(t, i) in w.tasks" :key="i" class="task-item">
              <span class="task-check">☐</span>{{ t }}
            </li>
          </ul>
        </div>
      </div>

      <div class="actions">
        <button class="btn-primary" @click="generate">重新生成</button>
        <router-link to="/" class="btn-secondary">开始学习</router-link>
      </div>
    </template>
  </div>
</template>

<style scoped>
.plan-page { max-width: 700px; margin: 0 auto; padding: 20px; }
.header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.back-btn { color: var(--text-light); text-decoration: none; font-size: 16px; }
.header h1 { font-size: 28px; font-weight: 900; }

.section {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 20px;
  margin-bottom: 16px;
}
.section-title { font-size: 16px; font-weight: 800; margin-bottom: 14px; }

.field { display: block; margin-bottom: 14px; }
.field-label { display: block; font-size: 13px; font-weight: 700; color: var(--text-light); margin-bottom: 6px; }
.input {
  width: 100%;
  padding: 10px 12px;
  border: 2px solid var(--border);
  border-radius: 10px;
  font-size: 14px;
  font-family: inherit;
  box-sizing: border-box;
}
.input:focus { outline: none; border-color: var(--primary); }

.week-options { display: flex; gap: 8px; flex-wrap: wrap; }
.week-option {
  border: 2px solid var(--border);
  background: white;
  border-radius: 16px;
  padding: 5px 14px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}
.week-option.active { border-color: var(--primary); color: var(--primary); background: rgba(74,137,255,0.08); }

.btn-primary {
  width: 100%;
  padding: 12px;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 800;
  background: var(--primary);
  color: white;
  box-shadow: 0 4px 0 var(--primary-dark);
  cursor: pointer;
}
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.error { color: var(--danger); font-size: 13px; margin-top: 10px; text-align: center; }
.tip { color: var(--text-light); font-size: 12px; margin-top: 10px; text-align: center; }

.plan-hero {
  background: linear-gradient(135deg, #8b5cf6, #6366f1);
  border-radius: var(--radius);
  color: white;
  padding: 24px 20px;
  margin-bottom: 16px;
  text-align: center;
  box-shadow: 0 4px 0 #4f46e5;
}
.plan-goal { font-size: 18px; font-weight: 800; margin-bottom: 6px; }
.plan-meta { font-size: 13px; opacity: 0.9; }

.summary { border-color: #c4b5fd; background: #f5f3ff; }
.summary-text { font-size: 14px; line-height: 1.7; }

.week-list { display: flex; flex-direction: column; gap: 14px; margin-bottom: 20px; }
.week-card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 16px;
}
.week-head { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; }
.week-badge {
  background: var(--primary);
  color: white;
  font-size: 12px;
  font-weight: 800;
  padding: 3px 10px;
  border-radius: 12px;
  flex-shrink: 0;
}
.week-focus { font-size: 15px; font-weight: 800; }
.task-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 8px; }
.task-item { display: flex; gap: 8px; font-size: 13px; color: var(--text); line-height: 1.5; }
.task-check { color: var(--text-light); }

.actions { display: flex; gap: 12px; }
.actions .btn-primary, .actions .btn-secondary { flex: 1; text-align: center; }
.btn-secondary {
  padding: 12px;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 800;
  background: white;
  color: var(--text);
  text-decoration: none;
  box-shadow: 0 4px 0 #e5e5e5;
}
</style>
