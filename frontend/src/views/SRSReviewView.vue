<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { srsApi } from '@/api/client'
import type { SRSReviewItem } from '@/types'

const loading = ref(true)
const submitting = ref(false)
const reviews = ref<SRSReviewItem[]>([])
const dueCount = ref(0)
const currentIndex = ref(0)
const selectedAnswer = ref('')
const checked = ref(false)
const isCorrect = ref(false)
const feedback = ref('')

const currentItem = computed(() => reviews.value[currentIndex.value])

const parsedOptions = computed(() => {
  if (!currentItem.value?.options) return []
  try {
    return JSON.parse(currentItem.value.options) as string[]
  } catch {
    return []
  }
})

const progressPercent = computed(() => {
  if (reviews.value.length === 0) return 0
  return Math.round((currentIndex.value / reviews.value.length) * 100)
})

async function loadReviews() {
  loading.value = true
  try {
    const res = await srsApi.reviews()
    reviews.value = res.data.reviews || []
    dueCount.value = res.data.due_count || 0
    currentIndex.value = 0
  } catch (e) {
    console.error('加载今日复习失败', e)
  } finally {
    loading.value = false
  }
}

function checkAnswer() {
  if (!currentItem.value) return
  isCorrect.value = selectedAnswer.value.trim().toLowerCase() === currentItem.value.correct_answer.trim().toLowerCase()
  checked.value = true
}

async function submitReview(correct: boolean) {
  if (!currentItem.value || submitting.value) return
  submitting.value = true
  try {
    await srsApi.submitReview(currentItem.value.id, correct)
    reviews.value.splice(currentIndex.value, 1)
    if (reviews.value.length > 0 && currentIndex.value >= reviews.value.length) {
      currentIndex.value = reviews.value.length - 1
    }
    resetCard()
  } catch (e) {
    console.error('提交复习失败', e)
  } finally {
    submitting.value = false
  }
}

function resetCard() {
  checked.value = false
  selectedAnswer.value = ''
  isCorrect.value = false
  feedback.value = ''
}

function selectOption(option: string) {
  if (checked.value) return
  selectedAnswer.value = option
}

onMounted(loadReviews)
</script>

<template>
  <div class="srs-page">
    <div class="header">
      <router-link to="/" class="back-btn">← 返回</router-link>
      <h1>🧠 今日复习</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="reviews.length === 0" class="empty">
      <div class="empty-icon">🎉</div>
      <p>今日没有待复习的错题，去学习新知识吧！</p>
      <router-link to="/" class="btn-primary go-btn">继续学习</router-link>
    </div>

    <div v-else class="content">
      <div class="toolbar">
        <div class="progress">
          {{ currentIndex + 1 }} / {{ reviews.length }}
        </div>
        <div class="due-badge" v-if="dueCount > 0">今日待复习 {{ dueCount }} 题</div>
      </div>

      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
      </div>

      <div v-if="currentItem" class="card" :key="currentItem.id">
        <div class="card-header">
          <span class="badge type-badge">{{ currentItem.type }}</span>
          <span class="stage-badge">熟练度 {{ currentItem.review_stage }}/5</span>
        </div>

        <div class="question">{{ currentItem.question }}</div>

        <div v-if="currentItem.type === 'choice'" class="options">
          <button
            v-for="opt in parsedOptions"
            :key="opt"
            class="option"
            :class="{
              selected: selectedAnswer === opt,
              correct: checked && opt === currentItem.correct_answer,
              wrong: checked && selectedAnswer === opt && opt !== currentItem.correct_answer,
            }"
            @click="selectOption(opt)"
            :disabled="checked"
          >
            {{ opt }}
          </button>
        </div>

        <div v-else-if="currentItem.type === 'fillblank'" class="fillblank">
          <input
            v-model="selectedAnswer"
            placeholder="输入答案..."
            :disabled="checked"
            @keyup.enter="checkAnswer"
          />
        </div>

        <div v-else class="code-section">
          <p class="hint">代码题：回忆一下解题思路，然后选择掌握程度。</p>
        </div>

        <div v-if="checked" class="result" :class="{ ok: isCorrect, no: !isCorrect }">
          <div class="result-icon">{{ isCorrect ? '✅' : '❌' }}</div>
          <div class="result-content">
            <p v-if="!isCorrect">
              你的答案：{{ selectedAnswer || '(空)' }}
            </p>
            <p>正确答案：{{ currentItem.correct_answer }}</p>
            <p class="explanation">{{ currentItem.explanation }}</p>
          </div>
        </div>

        <div class="actions">
          <button
            v-if="!checked && currentItem.type !== 'code'"
            class="btn primary"
            @click="checkAnswer"
            :disabled="!selectedAnswer"
          >
            检查
          </button>
          <template v-else>
            <button class="btn forget" @click="submitReview(false)" :disabled="submitting">😵 忘记了</button>
            <button class="btn primary" @click="submitReview(true)" :disabled="submitting">😎 记住了</button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.srs-page {
  max-width: 700px;
  margin: 0 auto;
  padding: 20px;
}

.header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.back-btn {
  color: var(--text-light);
  text-decoration: none;
  font-size: 16px;
}

.header h1 {
  font-size: 28px;
  font-weight: 900;
}

.loading, .empty {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-light);
}

.empty-icon { font-size: 60px; margin-bottom: 16px; }
.empty p { margin-bottom: 20px; }
.go-btn { display: inline-block; text-decoration: none; }

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.progress {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-light);
}

.due-badge {
  background: #e0f2fe;
  color: #0369a1;
  font-size: 12px;
  font-weight: 700;
  padding: 4px 12px;
  border-radius: 16px;
}

.progress-bar {
  height: 12px;
  background: var(--bg-gray);
  border-radius: 6px;
  overflow: hidden;
  border: 2px solid var(--border);
  margin-bottom: 20px;
}

.progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 6px;
  transition: width 0.3s;
}

.card {
  background: white;
  border-radius: 16px;
  padding: 28px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
}

.badge {
  font-size: 12px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 8px;
}

.type-badge { background: #e0f2fe; color: #0369a1; }
.stage-badge { background: #fef3c7; color: #92400e; }

.question {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 24px;
  line-height: 1.6;
}

.options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.option {
  text-align: left;
  padding: 14px 18px;
  border: 2px solid var(--border);
  border-radius: 12px;
  background: white;
  font-size: 15px;
  cursor: pointer;
  transition: all 0.2s;
}

.option:hover:not(:disabled) {
  border-color: var(--primary);
  background: #f0f7ff;
}

.option.selected {
  border-color: var(--primary);
  background: #e0f2fe;
}

.option.correct {
  border-color: #22c55e;
  background: #d1fae5;
}

.option.wrong {
  border-color: #ef4444;
  background: #fee2e2;
}

.fillblank input {
  width: 100%;
  padding: 14px 18px;
  border: 2px solid var(--border);
  border-radius: 12px;
  font-size: 15px;
  outline: none;
}

.fillblank input:focus {
  border-color: var(--primary);
}

.code-section { text-align: center; }
.hint { color: var(--text-light); font-size: 14px; margin-bottom: 12px; }

.result {
  margin-top: 20px;
  padding: 16px;
  border-radius: 12px;
  display: flex;
  gap: 12px;
}

.result.ok { background: #d1fae5; }
.result.no { background: #fee2e2; }

.result-icon { font-size: 24px; }

.result-content p { margin: 4px 0; font-size: 14px; }
.explanation { color: var(--text-light); }

.actions {
  margin-top: 24px;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn {
  padding: 12px 28px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 700;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:disabled { opacity: 0.5; cursor: default; }

.btn.primary {
  background: var(--primary);
  color: white;
}

.btn.secondary {
  background: #e5e7eb;
  color: var(--text);
}

.btn.forget {
  background: #fee2e2;
  color: #b91c1c;
}
</style>
