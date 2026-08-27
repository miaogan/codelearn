<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { courseApi, adaptiveApi } from '@/api/client'
import StatsBar from '@/components/StatsBar.vue'

const router = useRouter()
const courses = ref<any[]>([])
const selectedCourse = ref<any>(null)
const loading = ref(false)
const result = ref<any>(null)
const error = ref('')

async function loadCourses() {
  try {
    const res = await courseApi.list()
    courses.value = res.data.courses
  } catch {
    // ignore
  }
}

async function getRecommendation(courseId: number, language: string) {
  selectedCourse.value = courses.value.find(c => c.id === courseId)
  loading.value = true
  error.value = ''
  result.value = null
  try {
    const res = await adaptiveApi.recommend(courseId, language)
    result.value = res.data
  } catch (e: any) {
    error.value = e.response?.data?.error || '获取推荐失败'
  } finally {
    loading.value = false
  }
}

function selectAnswer(exIndex: number, option: string) {
  if (!result.value?.exercises) return
  const ex = result.value.exercises[exIndex]
  if (!ex._userAnswer) ex._userAnswer = ''
  ex._userAnswer = option
}

function checkAnswer(exIndex: number) {
  const ex = result.value.exercises[exIndex]
  if (!ex?._userAnswer) return
  ex._correct = ex._userAnswer === ex.answer
  ex._checked = true
}

onMounted(loadCourses)
</script>

<template>
  <div class="page">
    <StatsBar />
    <div class="container">
      <h1 class="title">自适应学习路径</h1>
      <p class="subtitle">基于错题分析，AI 自动识别薄弱知识点并生成针对性练习</p>

      <div class="step-section">
        <h2 class="step-title">1. 选择课程</h2>
        <div class="course-list">
          <div
            v-for="c in courses"
            :key="c.id"
            class="course-card"
            :class="{ active: selectedCourse?.id === c.id }"
            @click="getRecommendation(c.id, c.language)"
          >
            <span class="course-emoji">{{ c.emoji }}</span>
            <div class="course-info">
              <div class="course-name">{{ c.title }}</div>
              <div class="course-lang">{{ c.language }}</div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="loading" class="loading-section">
        <div class="loading-spinner"></div>
        <p>AI 正在分析你的错题并生成针对性练习...</p>
      </div>

      <div v-if="error" class="error-msg">{{ error }}</div>

      <div v-if="result && !loading" class="result-section">
        <div v-if="result.weak_points?.length > 0" class="weak-points">
          <h2 class="step-title">2. 薄弱知识点分析</h2>
          <div class="weak-point-list">
            <div v-for="(wp, i) in result.weak_points" :key="i" class="weak-point-card">
              <div class="wp-header">
                <span class="wp-icon">⚠️</span>
                <span class="wp-topic">{{ wp.topic }}</span>
                <span class="wp-count">错 {{ wp.count }} 次</span>
              </div>
              <p class="wp-desc">{{ wp.description }}</p>
            </div>
          </div>
          <div class="summary-box">
            <span class="summary-icon">💡</span>
            {{ result.summary }}
          </div>
        </div>

        <div v-if="result.exercises?.length > 0" class="exercises-section">
          <h2 class="step-title">3. 针对性练习</h2>
          <div v-for="(ex, i) in result.exercises" :key="i" class="exercise-card">
            <div class="ex-question">{{ i + 1 }}. {{ ex.question }}</div>
            <div v-if="ex.type === 'choice'" class="ex-options">
              <div
                v-for="opt in (typeof ex.options === 'string' ? JSON.parse(ex.options) : ex.options)"
                :key="opt"
                class="ex-option"
                :class="{
                  selected: ex._userAnswer === opt,
                  correct: ex._checked && opt === ex.answer,
                  wrong: ex._checked && ex._userAnswer === opt && opt !== ex.answer,
                }"
                @click="!ex._checked && selectAnswer(i, opt)"
              >
                {{ opt }}
              </div>
            </div>
            <div v-else class="ex-input">
              <textarea v-model="ex._userAnswer" :disabled="ex._checked" placeholder="输入你的答案" rows="3"></textarea>
            </div>
            <button v-if="!ex._checked" class="check-btn" @click="checkAnswer(i)" :disabled="!ex._userAnswer">
              检查
            </button>
            <div v-if="ex._checked" class="ex-result" :class="{ right: ex._correct, wrong: !ex._correct }">
              {{ ex._correct ? '✅ 正确' : '❌ 错误' }}
              <span class="ex-answer">正确答案: {{ ex.answer }}</span>
            </div>
            <div v-if="ex._checked && ex.explanation" class="ex-explanation">
              {{ ex.explanation }}
            </div>
          </div>
        </div>

        <div v-if="!result.weak_points?.length && !result.exercises?.length" class="empty-state">
          <span class="empty-icon">🎯</span>
          <p>暂无足够错题数据进行分析</p>
          <p class="empty-hint">继续学习，做更多练习题来积累数据</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { min-height: 100vh; background: var(--bg); }
.container { max-width: 700px; margin: 0 auto; padding: 20px; }
.title { font-size: 28px; font-weight: 900; color: var(--primary); margin-bottom: 4px; }
.subtitle { color: var(--text-light); font-size: 14px; margin-bottom: 24px; }
.step-section { margin-bottom: 24px; }
.step-title { font-size: 18px; font-weight: 800; margin-bottom: 12px; }
.course-list { display: flex; gap: 12px; flex-wrap: wrap; }
.course-card { display: flex; align-items: center; gap: 10px; padding: 12px 16px; background: white; border: 2px solid var(--border); border-radius: 12px; cursor: pointer; transition: all 0.2s; }
.course-card.active { border-color: var(--primary); background: var(--primary-bg); }
.course-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.course-emoji { font-size: 28px; }
.course-name { font-weight: 700; }
.course-lang { font-size: 12px; color: var(--text-light); }
.loading-section { text-align: center; padding: 40px; }
.loading-spinner { width: 48px; height: 48px; border: 4px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 1s linear infinite; margin: 0 auto 16px; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-msg { color: #ef4444; text-align: center; padding: 20px; }
.weak-points { margin-bottom: 24px; }
.weak-point-list { display: flex; flex-direction: column; gap: 12px; margin-bottom: 16px; }
.weak-point-card { background: #fef3c7; border: 1px solid #fcd34d; border-radius: 12px; padding: 16px; }
.wp-header { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.wp-icon { font-size: 18px; }
.wp-topic { font-weight: 800; flex: 1; }
.wp-count { font-size: 12px; color: #92400e; background: #fde68a; padding: 2px 8px; border-radius: 10px; }
.wp-desc { font-size: 14px; color: #78350f; }
.summary-box { background: var(--primary-bg); border-radius: 12px; padding: 14px 16px; display: flex; align-items: flex-start; gap: 8px; font-size: 14px; }
.summary-icon { font-size: 20px; }
.exercises-section { margin-bottom: 24px; }
.exercise-card { background: white; border: 2px solid var(--border); border-radius: 12px; padding: 16px; margin-bottom: 12px; }
.ex-question { font-size: 15px; font-weight: 600; margin-bottom: 12px; }
.ex-options { display: flex; flex-direction: column; gap: 8px; }
.ex-option { padding: 10px 14px; border: 2px solid var(--border); border-radius: 8px; cursor: pointer; transition: all 0.2s; }
.ex-option:hover { border-color: var(--primary); }
.ex-option.selected { border-color: var(--primary); background: var(--primary-bg); }
.ex-option.correct { border-color: #22c55e; background: #dcfce7; }
.ex-option.wrong { border-color: #ef4444; background: #fee2e2; }
.ex-input textarea { width: 100%; border: 2px solid var(--border); border-radius: 8px; padding: 10px; font-size: 14px; resize: vertical; }
.check-btn { margin-top: 8px; padding: 8px 20px; background: var(--primary); color: white; border: none; border-radius: 8px; font-weight: 700; cursor: pointer; }
.check-btn:disabled { opacity: 0.5; }
.ex-result { margin-top: 8px; font-weight: 700; }
.ex-result.right { color: #22c55e; }
.ex-result.wrong { color: #ef4444; }
.ex-answer { font-weight: 400; font-size: 13px; margin-left: 8px; }
.ex-explanation { margin-top: 8px; padding: 10px; background: var(--bg); border-radius: 8px; font-size: 13px; color: var(--text-light); }
.empty-state { text-align: center; padding: 40px; }
.empty-icon { font-size: 48px; }
.empty-hint { font-size: 14px; color: var(--text-light); }
</style>
