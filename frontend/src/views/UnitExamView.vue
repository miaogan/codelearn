<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { examApi } from '@/api/client'
import type { Exam, ExamReport, ExamQuestion } from '@/types'

const route = useRoute()
const router = useRouter()

const phase = ref<'loading' | 'intro' | 'exam' | 'result' | 'error'>('loading')
const exam = ref<Exam | null>(null)
const report = ref<ExamReport | null>(null)
const errorMsg = ref('')
const answers = ref<Map<number, string>>(new Map())
const currentIdx = ref(0)
const startTime = ref(0)
const remainSec = ref(0)
let timer: ReturnType<typeof setInterval> | undefined

const unitId = Number(route.params.unitId)

onMounted(async () => {
  try {
    exam.value = await (await examApi.startUnit(unitId)).data
    phase.value = 'intro'
  } catch (e: any) {
    errorMsg.value = e.response?.data?.error || '创建考试失败'
    phase.value = 'error'
  }
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})

const totalQuestions = computed(() => exam.value?.questions.length || 0)
const answeredCount = computed(() => {
  let n = 0
  for (const q of exam.value?.questions || []) {
    if (answers.value.get(q.id)) n++
  }
  return n
})
const allAnswered = computed(() => answeredCount.value >= totalQuestions.value)
const currentQuestion = computed<ExamQuestion | null>(() => {
  const qs = exam.value?.questions
  return qs ? qs[currentIdx.value] ?? null : null
})

function startExam() {
  startTime.value = Date.now()
  remainSec.value = (exam.value?.duration_min || 15) * 60
  phase.value = 'exam'
  timer = setInterval(() => {
    remainSec.value--
    if (remainSec.value <= 0) {
      submitExam()
    }
  }, 1000)
}

function parseOptions(options: string): string[] {
  try {
    const arr = JSON.parse(options || '[]')
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

function selectAnswer(q: ExamQuestion, option: string) {
  answers.value.set(q.id, option)
}

function next() {
  if (currentIdx.value < totalQuestions.value - 1) {
    currentIdx.value++
  }
}

function prev() {
  if (currentIdx.value > 0) {
    currentIdx.value--
  }
}

async function submitExam() {
  if (timer) clearInterval(timer)
  if (phase.value !== 'exam') return
  phase.value = 'loading'
  const durationSec = Math.round((Date.now() - startTime.value) / 1000)
  try {
    const payload = (exam.value?.questions || []).map((q) => ({
      question_id: q.id,
      exercise_id: q.exercise_id,
      answer: answers.value.get(q.id) || '',
    }))
    report.value = (await examApi.submit(exam.value!.id, payload, durationSec)).data
    phase.value = 'result'
  } catch (e: any) {
    errorMsg.value = e.response?.data?.error || '提交失败'
    phase.value = 'error'
  }
}

const fmtTime = computed(() => {
  const m = Math.floor(remainSec.value / 60)
  const s = remainSec.value % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
})

function levelText(level: string) {
  const map: Record<string, string> = { beginner: '入门', intermediate: '进阶', advanced: '熟练' }
  return map[level] || level
}
</script>

<template>
  <div class="container exam-page">
    <div v-if="phase === 'loading'" class="loading">加载中...</div>

    <!-- 考试说明 -->
    <div v-else-if="phase === 'intro' && exam" class="exam-intro card">
      <h1 class="intro-title">📝 {{ exam.title }}</h1>
      <div class="intro-meta">
        <div class="meta-item"><span class="meta-icon">⏱️</span>时长 {{ exam.duration_min }} 分钟</div>
        <div class="meta-item"><span class="meta-icon">📋</span>{{ totalQuestions }} 道题</div>
        <div class="meta-item"><span class="meta-icon">✅</span>及格线 {{ exam.pass_score }} 分</div>
      </div>
      <p class="intro-tip">考试期间不扣心数，请独立完成。提交后自动生成成绩报告，错题会进入错题本。</p>
      <button class="btn-primary start-btn" @click="startExam">开始考试</button>
    </div>

    <!-- 答题中 -->
    <div v-else-if="phase === 'exam' && exam && currentQuestion" class="card exam-body">
      <div class="exam-topbar">
        <div class="exam-progress">第 {{ currentIdx + 1 }} / {{ totalQuestions }} 题</div>
        <div class="exam-timer" :class="{ urgent: remainSec < 60 }">⏱️ {{ fmtTime }}</div>
      </div>

      <div class="question-box">
        <div class="q-badge">{{ currentQuestion.type === 'choice' ? '选择题' : currentQuestion.type === 'fillblank' ? '填空题' : '代码题' }}</div>
        <p class="q-text">{{ currentQuestion.question }}</p>

        <!-- 选择题 -->
        <div v-if="currentQuestion.type === 'choice'" class="options">
          <button
            v-for="opt in parseOptions(currentQuestion.options)"
            :key="opt"
            class="option"
            :class="{ selected: answers.get(currentQuestion.id) === opt }"
            @click="selectAnswer(currentQuestion, opt)"
          >{{ opt }}</button>
        </div>

        <!-- 填空题 -->
        <div v-else-if="currentQuestion.type === 'fillblank'" class="fill-blank">
          <input
            class="fill-input"
            placeholder="输入你的答案"
            :value="answers.get(currentQuestion.id) || ''"
            @input="answers.set(currentQuestion.id, ($event.target as HTMLInputElement).value)"
          />
        </div>

        <!-- 代码题 -->
        <div v-else class="code-box">
          <textarea
            class="code-input"
            :value="answers.get(currentQuestion.id) || currentQuestion.code_template || ''"
            placeholder="在这里写代码"
            spellcheck="false"
            @input="answers.set(currentQuestion.id, ($event.target as HTMLTextAreaElement).value)"
          ></textarea>
          <p class="code-tip">代码将由测试用例自动评判，全部通过才算答对。</p>
        </div>
      </div>

      <div class="exam-nav">
        <button class="btn-ghost" :disabled="currentIdx === 0" @click="prev">← 上一题</button>
        <button
          v-if="currentIdx < totalQuestions - 1"
          class="btn-primary"
          @click="next"
        >下一题 →</button>
        <button
          v-else
          class="btn-primary submit-btn"
          :disabled="!allAnswered"
          @click="submitExam"
        >交卷（{{ answeredCount }}/{{ totalQuestions }}）</button>
      </div>
    </div>

    <!-- 成绩报告 -->
    <div v-else-if="phase === 'result' && report" class="card result-card">
      <div class="result-head" :class="{ pass: report.passed, fail: !report.passed }">
        <span class="result-icon">{{ report.passed ? '🎉' : '💪' }}</span>
        <h1 class="result-title">{{ report.passed ? '考试通过！' : '未通过，再接再厉' }}</h1>
        <div class="result-score">{{ report.score }}<span class="result-unit">分</span></div>
        <p class="result-meta">答对 {{ report.correct_count }}/{{ report.total_count }} · 用时 {{ Math.floor(report.duration_sec / 60) }}分{{ report.duration_sec % 60 }}秒 · 第 {{ report.attempts }} 次</p>
      </div>

      <div class="result-detail">
        <div
          v-for="(r, i) in report.results"
          :key="r.question_id"
          class="result-item"
          :class="{ correct: r.correct }"
        >
          <span class="ri-icon">{{ r.correct ? '✅' : '❌' }}</span>
          <div class="ri-body">
            <div class="ri-question">第 {{ i + 1 }} 题</div>
            <div v-if="!r.correct" class="ri-answer">
              正确答案：<strong>{{ r.correct_answer || '（代码题）' }}</strong>
            </div>
            <div class="ri-explain">{{ r.explanation }}</div>
          </div>
        </div>
      </div>

      <div class="result-actions">
        <button class="btn-primary" @click="router.push('/')">返回首页</button>
        <button class="btn-secondary" @click="router.push('/wrong-exercises')">去复习错题</button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-else-if="phase === 'error'" class="card error-card">
      <h2>😢 {{ errorMsg }}</h2>
      <button class="btn-primary" @click="router.push('/')">返回首页</button>
    </div>
  </div>
</template>

<style scoped>
.exam-page { padding-top: 24px; padding-bottom: 40px; }
.card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius);
  padding: 24px;
  box-shadow: 0 4px 0 #e5e5e5;
}
.loading { text-align: center; padding: 40px; color: var(--text-light); }

/* 考试说明 */
.exam-intro { text-align: center; }
.intro-title { font-size: 24px; margin-bottom: 16px; }
.intro-meta { display: flex; justify-content: center; gap: 24px; flex-wrap: wrap; margin-bottom: 16px; }
.meta-item { display: flex; align-items: center; gap: 6px; font-weight: 700; color: var(--text-light); }
.meta-icon { font-size: 18px; }
.intro-tip { color: var(--text-light); font-size: 13px; margin-bottom: 20px; }
.start-btn { padding: 14px 40px; font-size: 17px; }

/* 答题 */
.exam-topbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.exam-progress { font-weight: 800; }
.exam-timer { font-weight: 900; color: var(--primary); }
.exam-timer.urgent { color: var(--danger); animation: blink 1s infinite; }
@keyframes blink { 50% { opacity: 0.4; } }

.question-box { margin-bottom: 20px; }
.q-badge {
  display: inline-block;
  background: var(--bg-gray);
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 700;
  color: var(--text-light);
  margin-bottom: 10px;
}
.q-text { font-size: 17px; font-weight: 700; margin-bottom: 16px; line-height: 1.6; }

.options { display: flex; flex-direction: column; gap: 10px; }
.option {
  text-align: left;
  padding: 14px 16px;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  background: white;
  font-size: 15px;
  cursor: pointer;
  transition: all 0.15s;
}
.option:hover { border-color: var(--primary); }
.option.selected { border-color: var(--primary); background: rgba(74, 137, 255, 0.1); font-weight: 700; }

.fill-input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 16px;
  outline: none;
}
.fill-input:focus { border-color: var(--primary); }

.code-box { display: flex; flex-direction: column; gap: 8px; }
.code-input {
  width: 100%;
  min-height: 200px;
  font-family: 'SF Mono', Consolas, 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
  padding: 12px;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  background: #f8fafc;
  outline: none;
  resize: vertical;
}
.code-input:focus { border-color: var(--primary); }
.code-tip { font-size: 12px; color: var(--text-light); }

.exam-nav { display: flex; justify-content: space-between; gap: 12px; }
.btn-primary, .btn-secondary, .btn-ghost {
  padding: 12px 20px;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-primary { background: var(--primary); color: white; box-shadow: 0 4px 0 var(--primary-dark); }
.btn-primary:active { transform: translateY(2px); box-shadow: 0 2px 0 var(--primary-dark); }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; box-shadow: none; }
.btn-secondary {
  background: white;
  color: var(--text);
  border: 2px solid var(--border);
  box-shadow: 0 4px 0 #e5e5e5;
}
.btn-ghost { background: transparent; color: var(--text-light); }

/* 成绩报告 */
.result-card { text-align: center; }
.result-head { padding: 20px 0 16px; }
.result-head.pass .result-title { color: var(--secondary); }
.result-head.fail .result-title { color: var(--danger); }
.result-icon { font-size: 56px; display: block; margin-bottom: 8px; }
.result-title { font-size: 26px; margin-bottom: 8px; }
.result-score { font-size: 56px; font-weight: 900; color: var(--text); }
.result-unit { font-size: 20px; color: var(--text-light); margin-left: 4px; }
.result-meta { color: var(--text-light); font-size: 14px; margin-top: 4px; }

.result-detail { text-align: left; margin-top: 20px; display: flex; flex-direction: column; gap: 8px; }
.result-item {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  background: var(--bg-gray);
  border: 2px solid transparent;
}
.result-item.correct { border-color: #86efac; }
.result-item:not(.correct) { border-color: #fca5a5; }
.ri-icon { font-size: 18px; }
.ri-body { flex: 1; }
.ri-question { font-weight: 800; font-size: 14px; }
.ri-answer { font-size: 13px; margin-top: 2px; }
.ri-explain { font-size: 13px; color: var(--text-light); margin-top: 2px; }

.result-actions { display: flex; gap: 12px; margin-top: 20px; }
.result-actions > * { flex: 1; }

.error-card { text-align: center; }
.error-card h2 { margin-bottom: 20px; font-size: 18px; }
</style>
