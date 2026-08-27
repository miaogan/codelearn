<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { wrongApi } from '@/api/client'
import type { WrongExerciseItem } from '@/types'

const wrongExercises = ref<WrongExerciseItem[]>([])
const loading = ref(true)
const showMastered = ref(false)
const currentIndex = ref(0)
const selectedAnswer = ref('')
const checked = ref(false)
const isCorrect = ref(false)

const filteredList = computed(() =>
  wrongExercises.value.filter(w => showMastered.value ? true : !w.mastered)
)

const currentItem = computed(() => filteredList.value[currentIndex.value])

const parsedOptions = computed(() => {
  if (!currentItem.value?.options) return []
  try {
    return JSON.parse(currentItem.value.options) as string[]
  } catch {
    return []
  }
})

async function loadWrong() {
  loading.value = true
  try {
    const res = await wrongApi.list(false)
    wrongExercises.value = res.data.wrong_exercises || []
  } catch (e) {
    console.error('加载错题失败', e)
  } finally {
    loading.value = false
  }
}

function checkAnswer() {
  if (!currentItem.value) return
  isCorrect.value = selectedAnswer.value.trim().toLowerCase() === currentItem.value.correct_answer.trim().toLowerCase()
  checked.value = true
}

function next() {
  checked.value = false
  selectedAnswer.value = ''
  if (currentIndex.value < filteredList.value.length - 1) {
    currentIndex.value++
  } else {
    currentIndex.value = 0
  }
}

async function markMastered() {
  if (!currentItem.value) return
  try {
    await wrongApi.master(currentItem.value.exercise_id)
    currentItem.value.mastered = true
    next()
  } catch (e) {
    console.error('标记失败', e)
  }
}

function selectOption(option: string) {
  if (checked.value) return
  selectedAnswer.value = option
}

onMounted(loadWrong)
</script>

<template>
  <div class="wrong-page">
    <div class="header">
      <router-link to="/" class="back-btn">← 返回</router-link>
      <h1>📝 错题本</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="filteredList.length === 0" class="empty">
      <div class="empty-icon">🎉</div>
      <p>没有错题，继续保持！</p>
    </div>

    <div v-else class="content">
      <div class="toolbar">
        <div class="progress">
          {{ currentIndex + 1 }} / {{ filteredList.length }}
        </div>
        <label class="toggle">
          <input type="checkbox" v-model="showMastered" />
          <span>显示已掌握</span>
        </label>
      </div>

      <div v-if="currentItem" class="card" :key="currentItem.id">
        <div class="card-header">
          <span class="badge type-badge">{{ currentItem.type }}</span>
          <span class="badge diff-badge">{{ currentItem.difficulty }}</span>
          <span v-if="currentItem.mastered" class="badge mastered-badge">已掌握</span>
          <span class="wrong-count">错 {{ currentItem.wrong_count }} 次</span>
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

        <div v-else-if="currentItem.type === 'code'" class="code-section">
          <pre class="code-template">{{ currentItem.code_template }}</pre>
          <p class="hint">代码题请去课程页面练习</p>
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
          <button v-if="!checked" class="btn primary" @click="checkAnswer" :disabled="!selectedAnswer">
            检查
          </button>
          <template v-else>
            <button class="btn primary" @click="next">下一题</button>
            <button v-if="!isCorrect" class="btn secondary" @click="markMastered">已掌握</button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.wrong-page {
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
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 16px;
}

.header h1 {
  font-size: 28px;
  font-weight: 900;
  color: var(--text);
}

.loading, .empty {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.empty-icon {
  font-size: 60px;
  margin-bottom: 16px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.progress {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-secondary);
}

.toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
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
  flex-wrap: wrap;
}

.badge {
  font-size: 12px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 8px;
}

.type-badge { background: #e0f2fe; color: #0369a1; }
.diff-badge { background: #fef3c7; color: #92400e; }
.mastered-badge { background: #d1fae5; color: #065f46; }

.wrong-count {
  margin-left: auto;
  font-size: 14px;
  font-weight: 700;
  color: #ef4444;
}

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

.code-template {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 12px;
  font-size: 14px;
  overflow-x: auto;
}

.hint {
  margin-top: 12px;
  color: var(--text-secondary);
  font-size: 14px;
}

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
.explanation { color: var(--text-secondary); }

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
</style>
