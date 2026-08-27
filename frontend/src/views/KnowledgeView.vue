<script setup lang="ts">
import { ref } from 'vue'
import { knowledgeApi } from '@/api/client'
import StatsBar from '@/components/StatsBar.vue'

const language = ref('Go')
const question = ref('')
const loading = ref(false)
const result = ref<any>(null)
const error = ref('')
const chatHistory = ref<{ question: string; answer: string; sources: any[]; followUps: string[] }[]>([])

async function ask() {
  if (!question.value.trim()) return
  loading.value = true
  error.value = ''
  result.value = null
  try {
    const res = await knowledgeApi.ask({
      question: question.value,
      language: language.value,
    })
    result.value = res.data
    chatHistory.value.unshift({
      question: question.value,
      answer: res.data.answer,
      sources: res.data.sources || [],
      followUps: res.data.follow_ups || [],
    })
    question.value = ''
  } catch (e: any) {
    error.value = e.response?.data?.error || '获取回答失败'
  } finally {
    loading.value = false
  }
}

function askFollowUp(q: string) {
  question.value = q
  ask()
}
</script>

<template>
  <div class="page">
    <StatsBar />
    <div class="container">
      <h1 class="title">知识点问答</h1>
      <p class="subtitle">基于课程内容的 AI 问答，检索课程知识并给出个性化解释</p>

      <div class="ask-section">
        <select v-model="language" class="lang-select">
          <option value="Go">Go</option>
          <option value="Python">Python</option>
        </select>
        <div class="input-row">
          <input v-model="question" class="ask-input" placeholder="问任何关于编程的问题，比如：goroutine 和 channel 有什么关系？" @keyup.enter="ask" :disabled="loading" />
          <button class="ask-btn" @click="ask" :disabled="!question.trim() || loading">
            {{ loading ? '思考中...' : '提问' }}
          </button>
        </div>
      </div>

      <div v-if="loading" class="loading-section">
        <div class="loading-spinner"></div>
        <p>正在检索课程内容并生成回答...</p>
      </div>

      <div v-if="error" class="error-msg">{{ error }}</div>

      <div v-if="result" class="answer-box">
        <div class="answer-content">{{ result.answer }}</div>

        <div v-if="result.sources?.length > 0" class="sources-section">
          <h4>📚 参考来源</h4>
          <div v-for="(src, i) in result.sources" :key="i" class="source-card">
            <div class="src-course">{{ src.course_name }} → {{ src.lesson_title }}</div>
            <div class="src-snippet">{{ src.snippet }}</div>
          </div>
        </div>

        <div v-if="result.follow_ups?.length > 0" class="followup-section">
          <h4>💡 相关问题</h4>
          <div class="followup-list">
            <button v-for="(fq, i) in result.follow_ups" :key="i" class="followup-btn" @click="askFollowUp(fq)">
              {{ fq }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="chatHistory.length > 1" class="history-section">
        <h3>📜 历史问答</h3>
        <div v-for="(item, i) in chatHistory.slice(1)" :key="i" class="history-item">
          <div class="hist-q">Q: {{ item.question }}</div>
          <div class="hist-a">{{ item.answer.substring(0, 200) }}{{ item.answer.length > 200 ? '...' : '' }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { min-height: 100vh; background: var(--bg); }
.container { max-width: 700px; margin: 0 auto; padding: 20px; }
.title { font-size: 28px; font-weight: 900; color: var(--primary); margin-bottom: 4px; }
.subtitle { color: var(--text-light); font-size: 14px; margin-bottom: 20px; }
.ask-section { margin-bottom: 20px; }
.lang-select { padding: 8px 12px; border: 2px solid var(--border); border-radius: 8px; font-size: 14px; background: white; margin-bottom: 8px; }
.input-row { display: flex; gap: 8px; }
.ask-input { flex: 1; padding: 12px 16px; border: 2px solid var(--border); border-radius: 8px; font-size: 15px; }
.ask-input:focus { border-color: var(--primary); }
.ask-btn { padding: 12px 24px; background: var(--primary); color: white; border: none; border-radius: 8px; font-weight: 700; cursor: pointer; white-space: nowrap; }
.ask-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.loading-section { text-align: center; padding: 30px; }
.loading-spinner { width: 40px; height: 40px; border: 4px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 1s linear infinite; margin: 0 auto 12px; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-msg { color: #ef4444; text-align: center; padding: 12px; }
.answer-box { background: white; border: 2px solid var(--border); border-radius: 12px; padding: 20px; margin-bottom: 16px; }
.answer-content { font-size: 15px; line-height: 1.8; white-space: pre-wrap; }
.sources-section { margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border); }
.sources-section h4 { font-size: 14px; font-weight: 700; margin-bottom: 8px; color: var(--text-light); }
.source-card { background: var(--bg); border-radius: 8px; padding: 10px 12px; margin-bottom: 8px; }
.src-course { font-size: 13px; font-weight: 700; color: var(--primary); margin-bottom: 4px; }
.src-snippet { font-size: 13px; color: var(--text-light); line-height: 1.5; }
.followup-section { margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border); }
.followup-section h4 { font-size: 14px; font-weight: 700; margin-bottom: 8px; color: var(--text-light); }
.followup-list { display: flex; flex-wrap: wrap; gap: 8px; }
.followup-btn { padding: 8px 14px; background: var(--primary-bg); border: 1px solid var(--primary); border-radius: 20px; font-size: 13px; color: var(--primary); cursor: pointer; transition: all 0.2s; }
.followup-btn:hover { background: var(--primary); color: white; }
.history-section { margin-top: 24px; }
.history-section h3 { font-size: 16px; font-weight: 800; margin-bottom: 12px; }
.history-item { background: white; border: 1px solid var(--border); border-radius: 8px; padding: 12px; margin-bottom: 8px; }
.hist-q { font-size: 14px; font-weight: 600; margin-bottom: 4px; }
.hist-a { font-size: 13px; color: var(--text-light); line-height: 1.5; }
</style>
