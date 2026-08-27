<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { courseApi, exerciseApi, codeApi, userApi } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { Lesson, Exercise, SubmitResult } from '@/types'
import ChoiceExercise from '@/components/ChoiceExercise.vue'
import FillBlankExercise from '@/components/FillBlankExercise.vue'
import CodeExercise from '@/components/CodeExercise.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const lesson = ref<Lesson | null>(null)
const exercises = ref<Exercise[]>([])
const loading = ref(true)
const phase = ref<'reading' | 'exercising' | 'completed'>('reading')
const currentIdx = ref(0)
const userAnswer = ref('')
const checkResult = ref<SubmitResult | null>(null)
const checked = ref(false)
const codePassed = ref(false)
const aiGenLoading = ref(false)

const currentExercise = computed(() => exercises.value[currentIdx.value])
const progress = computed(() => {
  if (exercises.value.length === 0) return 0
  return Math.round((currentIdx.value / exercises.value.length) * 100)
})

onMounted(async () => {
  const id = Number(route.params.id)
  try {
    const [lessonRes, exRes] = await Promise.all([
      courseApi.lesson(id),
      courseApi.exercises(id),
    ])
    lesson.value = lessonRes.data.lesson
    exercises.value = exRes.data.exercises
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
})

function startExercises() {
  phase.value = 'exercising'
  resetCurrent()
}

function resetCurrent() {
  userAnswer.value = ''
  checkResult.value = null
  checked.value = false
  codePassed.value = false
}

async function checkAnswer() {
  if (!currentExercise.value || !userAnswer.value) return
  try {
    const res = await exerciseApi.submit(currentExercise.value.id, userAnswer.value)
    checkResult.value = res.data
    checked.value = true
    if (res.data.hearts !== undefined) {
      auth.stats.hearts = res.data.hearts
    }
  } catch (e) {
    // ignore
  }
}

function onCodePassed() {
  codePassed.value = true
  checked.value = true
  checkResult.value = { correct: true, explanation: '代码通过所有测试用例！', hearts: auth.stats.hearts }
}

async function nextExercise() {
  if (currentIdx.value < exercises.value.length - 1) {
    currentIdx.value++
    resetCurrent()
  } else {
    await completeLesson()
  }
}

async function completeLesson() {
  try {
    const res = await codeApi.complete(Number(route.params.id))
    const statsRes = await userApi.stats()
    auth.setStats(statsRes.data)
    phase.value = 'completed'
  } catch (e) {
    phase.value = 'completed'
  }
}

async function generateAIExercises() {
  if (!lesson.value) return
  aiGenLoading.value = true
  try {
    const language = lesson.value.title.includes('Go') ? 'go' : 'python'
    const res = await exerciseApi.generate(Number(route.params.id), {
      language,
      topic: lesson.value.title,
      count: 3,
      type: 'choice',
      difficulty: 'easy',
    })
    exercises.value = [...exercises.value, ...res.data.exercises]
  } catch (e: any) {
    alert(e.response?.data?.error || 'AI 生成失败，请检查 LLM 配置')
  } finally {
    aiGenLoading.value = false
  }
}

function back() {
  if (lesson.value) {
    router.push(`/course/${getCid()}`)
  } else {
    router.push('/')
  }
}

function getCid(): number {
  // 从 lesson 的 unit_id 反查不太方便，直接返回 1 作为默认
  return 1
}
</script>

<template>
  <div class="container lesson-page" v-if="!loading && lesson">
    <!-- 阅读阶段 -->
    <div v-if="phase === 'reading'" class="reading-phase">
      <div class="lesson-header">
        <span class="lesson-icon">{{ lesson.icon }}</span>
        <h1 class="lesson-title">{{ lesson.title }}</h1>
        <p class="lesson-desc">{{ lesson.description }}</p>
      </div>
      <div class="lesson-content">
        <pre class="content-text">{{ lesson.content }}</pre>
      </div>
      <button class="btn-primary start-btn" @click="startExercises" :disabled="exercises.length === 0">
        {{ exercises.length > 0 ? '开始练习' : '暂无习题' }}
      </button>
      <button class="btn-secondary ai-btn" @click="generateAIExercises" :disabled="aiGenLoading">
        {{ aiGenLoading ? 'AI 生成中...' : '🤖 AI 生成练习题' }}
      </button>
    </div>

    <!-- 练习阶段 -->
    <div v-else-if="phase === 'exercising'" class="exercise-phase">
      <div class="progress-bar-wrap">
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: progress + '%' }"></div>
        </div>
        <span class="progress-text">{{ currentIdx + 1 }}/{{ exercises.length }}</span>
      </div>

      <div class="exercise-card" :class="{ shake: checked && !checkResult?.correct, pop: checked && checkResult?.correct }">
        <div class="exercise-question">{{ currentExercise.question }}</div>

        <div class="exercise-body">
          <ChoiceExercise v-if="currentExercise.type === 'choice'" :exercise="currentExercise" @answer="userAnswer = $event" />
          <FillBlankExercise v-else-if="currentExercise.type === 'fillblank'" :exercise="currentExercise" @answer="userAnswer = $event" />
          <CodeExercise v-else-if="currentExercise.type === 'code'" :exercise="currentExercise" @passed="onCodePassed" />
        </div>
      </div>

      <div v-if="!checked" class="exercise-actions">
        <button class="btn-primary" @click="checkAnswer" :disabled="!userAnswer && currentExercise.type !== 'code'">
          检查
        </button>
      </div>

      <div v-else class="feedback-panel" :class="{ correct: checkResult?.correct, wrong: !checkResult?.correct }">
        <div class="feedback-header">
          <span v-if="checkResult?.correct" class="feedback-icon">✓</span>
          <span v-else class="feedback-icon">✗</span>
          <span class="feedback-text">{{ checkResult?.correct ? '答对了！' : '答错了' }}</span>
        </div>
        <div class="feedback-explanation">{{ checkResult?.explanation }}</div>
        <button class="btn-primary next-btn" @click="nextExercise">
          {{ currentIdx < exercises.length - 1 ? '继续' : '完成' }}
        </button>
      </div>
    </div>

    <!-- 完成阶段 -->
    <div v-else-if="phase === 'completed'" class="completed-phase">
      <div class="completed-icon">🎉</div>
      <h1 class="completed-title">课程完成！</h1>
      <p class="completed-subtitle">恭喜你完成了「{{ lesson.title }}」</p>
      <div class="completed-rewards">
        <div class="reward-item">
          <span class="reward-icon">💎</span>
          <span class="reward-value">+{{ auth.stats.xp }} XP</span>
        </div>
        <div class="reward-item">
          <span class="reward-icon">🔥</span>
          <span class="reward-value">{{ auth.stats.streak_days }} 天连续</span>
        </div>
      </div>
      <button class="btn-primary" @click="back">返回学习路径</button>
    </div>
  </div>

  <div v-else class="container loading">加载中...</div>
</template>

<style scoped>
.lesson-page { padding-top: 16px; padding-bottom: 40px; }

.reading-phase { display: flex; flex-direction: column; gap: 20px; }
.lesson-header { text-align: center; }
.lesson-icon { font-size: 48px; }
.lesson-title { font-size: 24px; margin: 8px 0 4px; }
.lesson-desc { color: var(--text-light); font-size: 14px; }

.lesson-content {
  background: var(--bg-gray);
  border-radius: var(--radius);
  padding: 20px;
}

.content-text {
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 14px;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.8;
  font-weight: 500;
  color: var(--text);
}

.start-btn { width: 100%; }
.ai-btn { width: 100%; }

.progress-bar-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.progress-bar {
  flex: 1;
  height: 16px;
  background: var(--bg-gray);
  border-radius: 8px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 8px;
  transition: width 0.3s;
}

.progress-text { font-size: 14px; color: var(--text-light); white-space: nowrap; }

.exercise-card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius);
  padding: 24px;
  margin-bottom: 16px;
}

.exercise-card.shake { animation: shake 0.4s; }
.exercise-card.pop { animation: pop 0.3s; }

.exercise-question {
  font-size: 20px;
  font-weight: 800;
  margin-bottom: 20px;
  line-height: 1.5;
}

.exercise-body { margin-bottom: 8px; }

.exercise-actions { margin-top: 16px; }
.exercise-actions .btn-primary { width: 100%; }

.feedback-panel {
  border-radius: var(--radius);
  padding: 20px;
  margin-top: 12px;
}

.feedback-panel.correct { background: var(--primary-light); }
.feedback-panel.wrong { background: #ffe5e5; }

.feedback-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 20px;
  font-weight: 900;
  margin-bottom: 12px;
}

.feedback-panel.correct .feedback-icon { color: var(--primary); }
.feedback-panel.wrong .feedback-icon { color: var(--danger); }

.feedback-explanation {
  font-size: 14px;
  color: var(--text);
  line-height: 1.6;
  margin-bottom: 16px;
  font-weight: 600;
}

.next-btn { width: 100%; }

.completed-phase {
  text-align: center;
  padding-top: 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.completed-icon { font-size: 72px; }
.completed-title { font-size: 28px; color: var(--primary); }
.completed-subtitle { color: var(--text-light); margin-bottom: 20px; }

.completed-rewards {
  display: flex;
  gap: 24px;
  margin-bottom: 32px;
}

.reward-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  background: var(--bg-gray);
  padding: 16px 24px;
  border-radius: var(--radius-sm);
}

.reward-icon { font-size: 32px; }
.reward-value { font-size: 16px; font-weight: 800; }

.loading { text-align: center; padding: 40px; color: var(--text-light); }
</style>
