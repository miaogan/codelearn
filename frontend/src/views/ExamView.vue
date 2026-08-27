<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { courseApi, exerciseApi } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { Exercise, Course, ExamResult } from '@/types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const course = ref<Course | null>(null)
const phase = ref<'setup' | 'loading' | 'exam' | 'review' | 'submitting' | 'result'>('setup')
const exercises = ref<Exercise[]>([])
const answers = ref<Map<number, string>>(new Map())
const currentIdx = ref(0)
const examResult = ref<ExamResult | null>(null)

const examConfig = ref({
  type: 'choice',
  difficulty: 'easy',
  count: 5,
})

const currentExercise = computed(() => exercises.value[currentIdx.value])
const answeredCount = computed(() => {
  let count = 0
  for (const ex of exercises.value) {
    if (answers.value.get(ex.id)) count++
  }
  return count
})
const allAnswered = computed(() => answeredCount.value === exercises.value.length)

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
  answers.value = new Map()
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
    currentIdx.value = 0
    phase.value = 'exam'
  } catch (e: any) {
    alert(e.response?.data?.error || 'AI 生成失败，请检查 LLM 配置')
    phase.value = 'setup'
  }
}

function selectAnswer(exerciseId: number, answer: string) {
  answers.value.set(exerciseId, answer)
}

function getAnswer(exerciseId: number): string {
  return answers.value.get(exerciseId) || ''
}

function nextQuestion() {
  if (currentIdx.value < exercises.value.length - 1) {
    currentIdx.value++
  }
}

function prevQuestion() {
  if (currentIdx.value > 0) {
    currentIdx.value--
  }
}

function goToQuestion(idx: number) {
  currentIdx.value = idx
}

function goToReview() {
  phase.value = 'review'
}

function backToExam() {
  phase.value = 'exam'
}

async function submitExam() {
  phase.value = 'submitting'
  const payload = exercises.value.map(ex => ({
    exercise_id: ex.id,
    answer: answers.value.get(ex.id) || '',
  }))
  try {
    const res = await exerciseApi.examSubmit(payload)
    examResult.value = res.data
    phase.value = 'result'
  } catch (e: any) {
    alert(e.response?.data?.error || '提交失败')
    phase.value = 'review'
  }
}

function parsedOptions(ex: Exercise): string[] {
  try {
    return JSON.parse(ex.options || '[]')
  } catch {
    return []
  }
}

function getResultItem(exerciseId: number) {
  return examResult.value?.results.find(r => r.exercise_id === exerciseId)
}
</script>

<template>
  <div class="container exam-page" v-if="course">
    <!-- 配置阶段 -->
    <div v-if="phase === 'setup'" class="setup-phase">
      <div class="setup-header" :style="{ background: course.color }">
        <span class="emoji">{{ course.emoji }}</span>
        <h1>AI 模拟考试</h1>
        <p>{{ course.title }}</p>
      </div>

      <div class="setup-form">
        <div class="form-group">
          <label>题型</label>
          <div class="option-row">
            <button v-for="t in [{v:'choice',l:'选择题'},{v:'fillblank',l:'填空题'},{v:'subjective',l:'主观题'},{v:'mixed',l:'混合题'}]" :key="t.v" class="option-btn" :class="{ active: examConfig.type === t.v }" @click="examConfig.type = t.v">{{ t.l }}</button>
          </div>
        </div>

        <div class="form-group">
          <label>难度</label>
          <div class="option-row">
            <button v-for="d in [{v:'easy',l:'简单'},{v:'medium',l:'中等'},{v:'hard',l:'困难'}]" :key="d.v" class="option-btn" :class="{ active: examConfig.difficulty === d.v }" @click="examConfig.difficulty = d.v">{{ d.l }}</button>
          </div>
        </div>

        <div class="form-group">
          <label>题量</label>
          <div class="option-row">
            <button v-for="n in [3, 5, 10]" :key="n" class="option-btn" :class="{ active: examConfig.count === n }" @click="examConfig.count = n">{{ n }} 题</button>
          </div>
        </div>

        <div class="info-box">
          <p>📝 考试模式说明：</p>
          <ul>
            <li>不扣心数，最后统一打分</li>
            <li>可自由切换题目，修改答案</li>
            <li>提交前可检查所有答案</li>
            <li>主观题由 AI 判定正确性</li>
          </ul>
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

    <!-- 答题阶段 -->
    <div v-else-if="phase === 'exam'" class="exam-phase">
      <div class="exam-topbar">
        <span class="q-progress">{{ currentIdx + 1 }} / {{ exercises.length }}</span>
        <span class="answered-count">已答 {{ answeredCount }} / {{ exercises.length }}</span>
      </div>

      <div class="q-nav">
        <button v-for="(ex, i) in exercises" :key="ex.id"
          class="q-dot" :class="{ active: i === currentIdx, answered: answers.get(ex.id) }"
          @click="goToQuestion(i)">{{ i + 1 }}</button>
      </div>

      <div class="exam-card">
        <div class="q-type-badge">{{ currentExercise.type }}</div>
        <div class="question">{{ currentExercise.question }}</div>

        <!-- 选择题 -->
        <div v-if="currentExercise.type === 'choice'" class="options">
          <button v-for="opt in parsedOptions(currentExercise)" :key="opt"
            class="option" :class="{ selected: getAnswer(currentExercise.id) === opt }"
            @click="selectAnswer(currentExercise.id, opt)">{{ opt }}</button>
        </div>

        <!-- 填空题 -->
        <div v-else-if="currentExercise.type === 'fillblank'" class="fillblank">
          <input :value="getAnswer(currentExercise.id)" @input="selectAnswer(currentExercise.id, ($event.target as HTMLInputElement).value)" placeholder="输入答案..." />
        </div>

        <!-- 主观题 -->
        <div v-else-if="currentExercise.type === 'subjective'" class="subjective">
          <textarea :value="getAnswer(currentExercise.id)" @input="selectAnswer(currentExercise.id, ($event.target as HTMLTextAreaElement).value)" placeholder="请输入你的答案，AI 将判定是否正确..." rows="6"></textarea>
        </div>
      </div>

      <div class="exam-actions">
        <button class="btn-ghost" @click="prevQuestion" :disabled="currentIdx === 0">← 上一题</button>
        <button v-if="currentIdx < exercises.length - 1" class="btn-primary" @click="nextQuestion">下一题 →</button>
        <button v-else class="btn-primary" @click="goToReview" :disabled="!allAnswered">检查答案</button>
      </div>
    </div>

    <!-- 检查阶段 -->
    <div v-else-if="phase === 'review'" class="review-phase">
      <h2>检查答案</h2>
      <p class="review-hint">提交前可点击任意题目修改答案</p>
      <div class="review-list">
        <div v-for="(ex, i) in exercises" :key="ex.id" class="review-item" @click="currentIdx = i; backToExam()">
          <div class="review-num" :class="{ answered: answers.get(ex.id) }">{{ i + 1 }}</div>
          <div class="review-content">
            <div class="review-q">{{ ex.question }}</div>
            <div class="review-a">{{ answers.get(ex.id) || '未作答' }}</div>
          </div>
        </div>
      </div>
      <button class="btn-primary submit-btn" @click="submitExam">提交试卷</button>
      <button class="btn-ghost" @click="backToExam">← 返回修改</button>
    </div>

    <!-- 提交中 -->
    <div v-else-if="phase === 'submitting'" class="loading-phase">
      <div class="loading-spinner"></div>
      <p>AI 正在批改试卷...</p>
    </div>

    <!-- 结果阶段 -->
    <div v-else-if="phase === 'result' && examResult" class="result-phase">
      <div class="result-icon">{{ examResult.score === 100 ? '🏆' : examResult.score >= 60 ? '👍' : '💪' }}</div>
      <h1 class="result-title">考试完成</h1>
      <div class="score-display">
        <span class="score-value">{{ examResult.correct_count }}</span>
        <span class="score-divider">/</span>
        <span class="score-total">{{ examResult.total_count }}</span>
      </div>
      <div class="score-label">{{ examResult.score }} 分</div>

      <div class="result-review">
        <h3>答题详情</h3>
        <div v-for="(ex, i) in exercises" :key="ex.id" class="result-item" :class="{ ok: getResultItem(ex.id)?.correct, no: !getResultItem(ex.id)?.correct }">
          <div class="result-icon-sm">{{ getResultItem(ex.id)?.correct ? '✅' : '❌' }}</div>
          <div class="result-content">
            <div class="result-q">Q{{ i + 1 }} [{{ ex.type }}] {{ ex.question }}</div>
            <div class="result-a">你的答案: {{ getAnswer(ex.id) || '(空)' }}</div>
            <div v-if="!getResultItem(ex.id)?.correct" class="result-correct">正确答案: {{ getResultItem(ex.id)?.correct_answer }}</div>
            <div class="result-explain">{{ getResultItem(ex.id)?.explanation }}</div>
            <div v-if="getResultItem(ex.id)?.feedback" class="result-feedback">AI 反馈: {{ getResultItem(ex.id)?.feedback }}</div>
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

.setup-header { border-radius: var(--radius); padding: 28px 20px; text-align: center; color: white; margin-bottom: 24px; }
.setup-header .emoji { font-size: 48px; display: block; margin-bottom: 8px; }
.setup-header h1 { font-size: 24px; margin-bottom: 4px; }
.setup-header p { font-size: 14px; opacity: 0.9; }

.setup-form { display: flex; flex-direction: column; gap: 24px; }
.form-group label { font-size: 14px; font-weight: 700; color: var(--text-light); display: block; margin-bottom: 10px; }
.option-row { display: flex; gap: 8px; flex-wrap: wrap; }
.option-btn { flex: 1; min-width: 70px; padding: 12px; border: 2px solid var(--border); border-radius: 12px; background: white; font-size: 15px; font-weight: 600; cursor: pointer; transition: all 0.2s; }
.option-btn:hover { border-color: var(--primary); }
.option-btn.active { border-color: var(--primary); background: #e0f2fe; color: var(--primary); }

.info-box { background: #f0f7ff; border-radius: 12px; padding: 16px; }
.info-box p { font-size: 14px; font-weight: 700; margin-bottom: 8px; }
.info-box ul { margin: 0; padding-left: 20px; font-size: 13px; color: var(--text-light); }
.info-box li { margin-bottom: 4px; }

.start-exam { width: 100%; }
.btn-ghost { width: 100%; margin-top: 8px; }

.loading-phase { text-align: center; padding: 60px 20px; }
.loading-spinner { width: 48px; height: 48px; border: 4px solid var(--border); border-top-color: var(--primary); border-radius: 50%; margin: 0 auto 16px; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.exam-topbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.q-progress { font-size: 16px; font-weight: 800; }
.answered-count { font-size: 13px; color: var(--text-light); }

.q-nav { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 20px; }
.q-dot { width: 32px; height: 32px; border-radius: 8px; border: 2px solid var(--border); background: white; font-size: 13px; font-weight: 700; cursor: pointer; transition: all 0.15s; }
.q-dot.answered { background: #d1fae5; border-color: #22c55e; }
.q-dot.active { background: var(--primary); color: white; border-color: var(--primary); }

.exam-card { background: white; border: 2px solid var(--border); border-radius: var(--radius); padding: 24px; margin-bottom: 16px; }
.q-type-badge { display: inline-block; font-size: 12px; font-weight: 700; padding: 4px 10px; border-radius: 8px; background: #e0f2fe; color: #0369a1; margin-bottom: 16px; }
.question { font-size: 20px; font-weight: 800; margin-bottom: 20px; line-height: 1.5; }

.options { display: flex; flex-direction: column; gap: 12px; }
.option { text-align: left; padding: 14px 18px; border: 2px solid var(--border); border-radius: 12px; background: white; font-size: 15px; cursor: pointer; transition: all 0.2s; }
.option:hover { border-color: var(--primary); background: #f0f7ff; }
.option.selected { border-color: var(--primary); background: #e0f2fe; }

.fillblank input { width: 100%; padding: 14px 18px; border: 2px solid var(--border); border-radius: 12px; font-size: 15px; outline: none; }
.fillblank input:focus { border-color: var(--primary); }

.subjective textarea { width: 100%; padding: 14px 18px; border: 2px solid var(--border); border-radius: 12px; font-size: 15px; outline: none; resize: vertical; font-family: inherit; }
.subjective textarea:focus { border-color: var(--primary); }

.exam-actions { display: flex; gap: 12px; }
.exam-actions button { flex: 1; }

.review-phase h2 { font-size: 22px; font-weight: 900; margin-bottom: 8px; }
.review-hint { font-size: 14px; color: var(--text-light); margin-bottom: 20px; }
.review-list { display: flex; flex-direction: column; gap: 8px; margin-bottom: 24px; }
.review-item { display: flex; gap: 12px; padding: 16px; border-radius: 12px; background: white; border: 2px solid var(--border); cursor: pointer; transition: all 0.15s; }
.review-item:hover { border-color: var(--primary); }
.review-num { width: 32px; height: 32px; border-radius: 8px; background: var(--bg-gray); display: flex; align-items: center; justify-content: center; font-weight: 700; flex-shrink: 0; }
.review-num.answered { background: #d1fae5; color: #059669; }
.review-content { flex: 1; min-width: 0; }
.review-q { font-size: 15px; font-weight: 700; margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.review-a { font-size: 13px; color: var(--text-light); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.submit-btn { width: 100%; margin-bottom: 8px; }

.result-phase { text-align: center; padding-top: 32px; }
.result-icon { font-size: 72px; margin-bottom: 12px; }
.result-title { font-size: 28px; color: var(--primary); margin-bottom: 16px; }
.score-display { font-size: 48px; font-weight: 900; }
.score-value { color: var(--primary); }
.score-divider { color: var(--text-light); margin: 0 4px; }
.score-total { color: var(--text-light); }
.score-label { font-size: 18px; color: var(--text-light); margin-top: 4px; margin-bottom: 32px; }

.result-review { margin-bottom: 32px; }
.result-review h3 { font-size: 18px; margin-bottom: 16px; text-align: left; }
.result-item { display: flex; gap: 12px; padding: 16px; border-radius: 12px; margin-bottom: 8px; text-align: left; }
.result-item.ok { background: #d1fae5; }
.result-item.no { background: #fee2e2; }
.result-icon-sm { font-size: 20px; flex-shrink: 0; }
.result-content { flex: 1; min-width: 0; }
.result-q { font-size: 15px; font-weight: 700; margin-bottom: 4px; }
.result-a { font-size: 13px; color: var(--text-light); }
.result-correct { font-size: 13px; color: #059669; font-weight: 600; margin-top: 2px; }
.result-explain { font-size: 13px; color: var(--text); margin-top: 4px; line-height: 1.5; }
.result-feedback { font-size: 13px; color: #7c3aed; margin-top: 4px; font-style: italic; }

.result-actions { display: flex; gap: 12px; }
.result-actions button { flex: 1; }

.loading { text-align: center; padding: 40px; color: var(--text-light); }
</style>
