<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { courseApi, exerciseApi } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { Exercise, Course } from '@/types'
import ChoiceExercise from '@/components/ChoiceExercise.vue'
import FillBlankExercise from '@/components/FillBlankExercise.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const course = ref<Course | null>(null)
const phase = ref<'setup' | 'loading' | 'exam' | 'result'>('setup')
const exercises = ref<Exercise[]>([])
const currentIdx = ref(0)
const userAnswer = ref('')
const checked = ref(false)
const isCorrect = ref(false)
const explanation = ref('')
const answers = ref<{ exercise: Exercise; answer: string; correct: boolean }[]>([])

const examConfig = ref({
  type: 'choice',
  difficulty: 'easy',
  count: 5,
})

const currentExercise = computed(() => exercises.value[currentIdx.value])
const progress = computed(() => {
  if (exercises.value.length === 0) return 0
  return Math.round((currentIdx.value / exercises.value.length) * 100)
})
const score = computed(() => answers.value.filter(a => a.correct).length)

onMounted(async () => {
  const id = Number(route.params.id)
  try {
    const res = await courseApi.path(id)
    course.value = res.data.course
  } catch {
    // ignore
  }
})

async function startExam() {
  if (!course.value) return
  phase.value = 'loading'
  try {
    const res = await exerciseApi.generate(0, {
      language: course.value.language,
      topic: course.value.title,
      count: examConfig.value.count,
      type: examConfig.value.type,
      difficulty: examConfig.value.difficulty,
    })
    if (res.data.exercises.length === 0) {
      alert('AI 未生成任何习题，请检查 LLM 配置')
      phase.value = 'setup'
      return
    }
    exercises.value = res.data.exercises
    phase.value = 'exam'
  } catch (e: any) {
    alert(e.response?.data?.error || 'AI 生成失败，请检查 LLM 配置')
    phase.value = 'setup'
  }
}

function resetCurrent() {
  userAnswer.value = ''
  checked.value = false
  isCorrect.value = false
  explanation.value = ''
}

async function checkAnswer() {
  if (!currentExercise.value || !userAnswer.value) return
  try {
    const res = await exerciseApi.submit(currentExercise.value.id, userAnswer.value)
    checked.value = true
    isCorrect.value = res.data.correct
    explanation.value = res.data.explanation
    if (res.data.hearts !== undefined) {
      auth.stats.hearts = res.data.hearts
    }
    answers.value.push({
      exercise: currentExercise.value,
      answer: userAnswer.value,
      correct: res.data.correct,
    })
  } catch {
    checked.value = true
    isCorrect.value = false
  }
}

function nextQuestion() {
  if (currentIdx.value < exercises.value.length - 1) {
    currentIdx.value++
    resetCurrent()
  } else {
    phase.value = 'result'
  }
}

function selectOption(option: string) {
  if (checked.value) return
  userAnswer.value = option
}
</script>

<template>
  <div class="container exam-page" v-if="course">
    <!-- 配置阶段 -->
    <div v-if="phase === 'setup'" class="setup-phase">
      <div class="setup-header" :style="{ background: course.color }">
        <span class="emoji">{{ course.emoji }}</span>
        <h1>AI 模拟考试</h1>
        <p>{{ course.title }} · {{ course.description }}</p>
      </div>

      <div class="setup-form">
        <div class="form-group">
          <label>题型</label>
          <div class="option-row">
            <button
              v-for="t in [{v:'choice',l:'选择题'},{v:'fillblank',l:'填空题'}]"
              :key="t.v"
              class="option-btn"
              :class="{ active: examConfig.type === t.v }"
              @click="examConfig.type = t.v"
            >{{ t.l }}</button>
          </div>
        </div>

        <div class="form-group">
          <label>难度</label>
          <div class="option-row">
            <button
              v-for="d in [{v:'easy',l:'简单'},{v:'medium',l:'中等'},{v:'hard',l:'困难'}]"
              :key="d.v"
              class="option-btn"
              :class="{ active: examConfig.difficulty === d.v }"
              @click="examConfig.difficulty = d.v"
            >{{ d.l }}</button>
          </div>
        </div>

        <div class="form-group">
          <label>题量</label>
          <div class="option-row">
            <button
              v-for="n in [3, 5, 10]"
              :key="n"
              class="option-btn"
              :class="{ active: examConfig.count === n }"
              @click="examConfig.count = n"
            >{{ n }} 题</button>
          </div>
        </div>

        <button class="btn-primary start-exam" @click="startExam">开始考试</button>
        <button class="btn-ghost" @click="router.back()">← 返回</button>
      </div>
    </div>

    <!-- 加载阶段 -->
    <div v-else-if="phase === 'loading'" class="loading-phase">
      <div class="loading-spinner"></div>
      <p>AI 正在生成试卷...</p>
    </div>

    <!-- 考试阶段 -->
    <div v-else-if="phase === 'exam'" class="exam-phase">
      <div class="progress-bar-wrap">
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: progress + '%' }"></div>
        </div>
        <span class="progress-text">{{ currentIdx + 1 }}/{{ exercises.length }}</span>
      </div>

      <div class="exam-card" :class="{ shake: checked && !isCorrect, pop: checked && isCorrect }">
        <div class="question">{{ currentExercise.question }}</div>

        <div v-if="currentExercise.type === 'choice'" class="options">
          <button
            v-for="opt in JSON.parse(currentExercise.options || '[]')"
            :key="opt"
            class="option"
            :class="{
              selected: userAnswer === opt,
              correct: checked && opt === currentExercise.answer,
              wrong: checked && userAnswer === opt && opt !== currentExercise.answer,
            }"
            @click="selectOption(opt)"
            :disabled="checked"
          >{{ opt }}</button>
        </div>

        <div v-else-if="currentExercise.type === 'fillblank'" class="fillblank">
          <input
            v-model="userAnswer"
            placeholder="输入答案..."
            :disabled="checked"
            @keyup.enter="checkAnswer"
          />
        </div>
      </div>

      <div v-if="!checked" class="actions">
        <button class="btn-primary" @click="checkAnswer" :disabled="!userAnswer">检查</button>
      </div>

      <div v-else class="feedback" :class="{ ok: isCorrect, no: !isCorrect }">
        <div class="feedback-header">
          <span class="feedback-icon">{{ isCorrect ? '✅' : '❌' }}</span>
          <span>{{ isCorrect ? '答对了！' : '答错了' }}</span>
        </div>
        <div class="feedback-explanation">{{ explanation }}</div>
        <button class="btn-primary" @click="nextQuestion">
          {{ currentIdx < exercises.length - 1 ? '下一题' : '查看结果' }}
        </button>
      </div>
    </div>

    <!-- 结果阶段 -->
    <div v-else-if="phase === 'result'" class="result-phase">
      <div class="result-icon">{{ score === exercises.length ? '🏆' : score >= exercises.length / 2 ? '👍' : '💪' }}</div>
      <h1 class="result-title">考试完成！</h1>
      <div class="score-display">
        <span class="score-value">{{ score }}</span>
        <span class="score-divider">/</span>
        <span class="score-total">{{ exercises.length }}</span>
      </div>
      <div class="score-label">{{ Math.round((score / exercises.length) * 100) }} 分</div>

      <div class="answer-review">
        <h3>答题回顾</h3>
        <div v-for="(a, i) in answers" :key="i" class="review-item" :class="{ ok: a.correct, no: !a.correct }">
          <div class="review-icon">{{ a.correct ? '✅' : '❌' }}</div>
          <div class="review-content">
            <div class="review-question">Q{{ i + 1 }}: {{ a.exercise.question }}</div>
            <div class="review-answer">你的答案: {{ a.answer || '(空)' }}</div>
            <div v-if="!a.correct" class="review-correct">正确答案: {{ a.exercise.answer }}</div>
          </div>
        </div>
      </div>

      <div class="result-actions">
        <button class="btn-primary" @click="router.push(`/course/${route.params.id}`)">返回学习路径</button>
        <button class="btn-ghost" @click="router.push('/wrong-exercises')">查看错题本</button>
      </div>
    </div>
  </div>

  <div v-else class="container loading">加载中...</div>
</template>

<style scoped>
.exam-page { padding-top: 16px; padding-bottom: 40px; }

.setup-header {
  border-radius: var(--radius);
  padding: 28px 20px;
  text-align: center;
  color: white;
  margin-bottom: 24px;
}
.setup-header .emoji { font-size: 48px; display: block; margin-bottom: 8px; }
.setup-header h1 { font-size: 24px; margin-bottom: 4px; }
.setup-header p { font-size: 14px; opacity: 0.9; }

.setup-form { display: flex; flex-direction: column; gap: 24px; }
.form-group label { font-size: 14px; font-weight: 700; color: var(--text-light); display: block; margin-bottom: 10px; }
.option-row { display: flex; gap: 8px; }
.option-btn {
  flex: 1; padding: 12px; border: 2px solid var(--border); border-radius: 12px;
  background: white; font-size: 15px; font-weight: 600; cursor: pointer; transition: all 0.2s;
}
.option-btn:hover { border-color: var(--primary); }
.option-btn.active { border-color: var(--primary); background: #e0f2fe; color: var(--primary); }
.start-exam { width: 100%; }
.btn-ghost { width: 100%; margin-top: 8px; }

.loading-phase { text-align: center; padding: 60px 20px; }
.loading-spinner {
  width: 48px; height: 48px; border: 4px solid var(--border); border-top-color: var(--primary);
  border-radius: 50%; margin: 0 auto 16px; animation: spin 1s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.progress-bar-wrap { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; }
.progress-bar { flex: 1; height: 16px; background: var(--bg-gray); border-radius: 8px; overflow: hidden; }
.progress-fill { height: 100%; background: linear-gradient(90deg, #6366f1, #8b5cf6); border-radius: 8px; transition: width 0.3s; }
.progress-text { font-size: 14px; color: var(--text-light); white-space: nowrap; }

.exam-card { background: white; border: 2px solid var(--border); border-radius: var(--radius); padding: 24px; margin-bottom: 16px; }
.exam-card.shake { animation: shake 0.4s; }
.exam-card.pop { animation: pop 0.3s; }
@keyframes shake { 0%,100%{transform:translateX(0)} 25%{transform:translateX(-8px)} 75%{transform:translateX(8px)} }
@keyframes pop { 0%{transform:scale(1)} 50%{transform:scale(1.02)} 100%{transform:scale(1)} }

.question { font-size: 20px; font-weight: 800; margin-bottom: 20px; line-height: 1.5; }

.options { display: flex; flex-direction: column; gap: 12px; }
.option {
  text-align: left; padding: 14px 18px; border: 2px solid var(--border); border-radius: 12px;
  background: white; font-size: 15px; cursor: pointer; transition: all 0.2s;
}
.option:hover:not(:disabled) { border-color: var(--primary); background: #f0f7ff; }
.option.selected { border-color: var(--primary); background: #e0f2fe; }
.option.correct { border-color: #22c55e; background: #d1fae5; }
.option.wrong { border-color: #ef4444; background: #fee2e2; }

.fillblank input {
  width: 100%; padding: 14px 18px; border: 2px solid var(--border); border-radius: 12px; font-size: 15px; outline: none;
}
.fillblank input:focus { border-color: var(--primary); }

.actions .btn-primary { width: 100%; }

.feedback { border-radius: var(--radius); padding: 20px; margin-top: 12px; }
.feedback.ok { background: #d1fae5; }
.feedback.no { background: #fee2e2; }
.feedback-header { display: flex; align-items: center; gap: 8px; font-size: 20px; font-weight: 900; margin-bottom: 12px; }
.feedback-explanation { font-size: 14px; line-height: 1.6; margin-bottom: 16px; }
.feedback .btn-primary { width: 100%; }

.result-phase { text-align: center; padding-top: 32px; }
.result-icon { font-size: 72px; margin-bottom: 12px; }
.result-title { font-size: 28px; color: var(--primary); margin-bottom: 16px; }
.score-display { font-size: 48px; font-weight: 900; }
.score-value { color: var(--primary); }
.score-divider { color: var(--text-light); margin: 0 4px; }
.score-total { color: var(--text-light); }
.score-label { font-size: 18px; color: var(--text-light); margin-top: 4px; margin-bottom: 32px; }

.answer-review { margin-bottom: 32px; }
.answer-review h3 { font-size: 18px; margin-bottom: 16px; text-align: left; }
.review-item {
  display: flex; gap: 12px; padding: 16px; border-radius: 12px; margin-bottom: 8px; text-align: left;
}
.review-item.ok { background: #d1fae5; }
.review-item.no { background: #fee2e2; }
.review-icon { font-size: 20px; flex-shrink: 0; }
.review-content { flex: 1; }
.review-question { font-size: 15px; font-weight: 700; margin-bottom: 4px; }
.review-answer { font-size: 13px; color: var(--text-light); }
.review-correct { font-size: 13px; color: #059669; font-weight: 600; margin-top: 2px; }

.result-actions { display: flex; gap: 12px; }
.result-actions button { flex: 1; }

.loading { text-align: center; padding: 40px; color: var(--text-light); }
</style>
