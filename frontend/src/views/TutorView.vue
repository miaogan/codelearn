<script setup lang="ts">
import { ref } from 'vue'
import { tutorApi } from '@/api/client'
import StatsBar from '@/components/StatsBar.vue'

const language = ref('go')
const code = ref('')
const question = ref('')
const loading = ref(false)
const result = ref<any>(null)
const error = ref('')

const chatMessages = ref<{ role: string; content: string }[]>([])
const chatInput = ref('')
const chatLoading = ref(false)

const reviewResult = ref('')
const reviewLoading = ref(false)

const runResult = ref<any>(null)

async function debug() {
  if (!code.value.trim()) return
  loading.value = true
  error.value = ''
  result.value = null
  try {
    const res = await tutorApi.debug({
      language: language.value,
      code: code.value,
      question: question.value || '这段代码有什么问题？',
    })
    result.value = res.data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'AI 诊断失败'
  } finally {
    loading.value = false
  }
}

async function runCode() {
  if (!code.value.trim()) return
  try {
    const res = await tutorApi.run({
      language: language.value,
      code: code.value,
    })
    runResult.value = res.data
  } catch (e: any) {
    runResult.value = { error: e.response?.data?.error || '运行失败' }
  }
}

async function review() {
  if (!code.value.trim()) return
  reviewLoading.value = true
  reviewResult.value = ''
  try {
    const res = await tutorApi.review({
      language: language.value,
      code: code.value,
    })
    reviewResult.value = res.data.review
  } catch (e: any) {
    reviewResult.value = '审查失败: ' + (e.response?.data?.error || '')
  } finally {
    reviewLoading.value = false
  }
}

async function chat() {
  if (!chatInput.value.trim()) return
  const userMsg = chatInput.value
  chatMessages.value.push({ role: 'user', content: userMsg })
  chatInput.value = ''
  chatLoading.value = true
  try {
    const res = await tutorApi.chat({
      messages: chatMessages.value,
      code: code.value,
      language: language.value,
    })
    chatMessages.value.push({ role: 'assistant', content: res.data.reply })
  } catch (e: any) {
    chatMessages.value.push({ role: 'assistant', content: '抱歉，回复失败: ' + (e.response?.data?.error || '') })
  } finally {
    chatLoading.value = false
  }
}

function useFixCode() {
  if (result.value?.fix_code) {
    code.value = result.value.fix_code
  }
}
</script>

<template>
  <div class="page">
    <StatsBar />
    <div class="container">
      <h1 class="title">AI 编程导师</h1>
      <p class="subtitle">贴入你的代码，AI 帮你运行、诊断、修复</p>

      <div class="toolbar">
        <select v-model="language" class="lang-select">
          <option value="go">Go</option>
          <option value="python">Python</option>
        </select>
        <button class="tool-btn" @click="runCode" :disabled="!code.trim()">▶ 运行</button>
        <button class="tool-btn" @click="review" :disabled="!code.trim() || reviewLoading">🔍 审查</button>
        <button class="tool-btn primary" @click="debug" :disabled="!code.trim() || loading">🤖 AI 诊断</button>
      </div>

      <div class="editor-section">
        <textarea v-model="code" class="code-editor" placeholder="在这里粘贴你的代码..." spellcheck="false"></textarea>
      </div>

      <input v-model="question" class="question-input" placeholder="描述你遇到的问题（可选）" />

      <div v-if="loading" class="loading-section">
        <div class="loading-spinner"></div>
        <p>AI 正在运行代码并诊断问题...</p>
      </div>

      <div v-if="error" class="error-msg">{{ error }}</div>

      <div v-if="runResult" class="result-box">
        <h3>运行结果</h3>
        <pre class="output" :class="{ 'has-error': runResult.error }">{{ runResult.error || runResult.output || '(无输出)' }}</pre>
      </div>

      <div v-if="result" class="result-box">
        <h3>AI 诊断结果</h3>
        <div class="diagnosis">
          <div class="diag-item"><strong>诊断：</strong>{{ result.diagnosis }}</div>
          <div class="diag-item"><strong>建议：</strong>{{ result.suggestion }}</div>
          <div class="diag-item"><strong>运行说明：</strong>{{ result.run_result }}</div>
        </div>
        <div v-if="result.fix_code" class="fix-section">
          <div class="fix-header">
            <h4>修复后的代码</h4>
            <button class="use-btn" @click="useFixCode">使用此代码</button>
          </div>
          <pre class="fix-code">{{ result.fix_code }}</pre>
        </div>
      </div>

      <div v-if="reviewResult" class="result-box">
        <h3>代码审查</h3>
        <pre class="output review-output">{{ reviewResult }}</pre>
      </div>

      <div class="chat-section">
        <h3>💬 追问对话</h3>
        <div class="chat-messages">
          <div v-for="(msg, i) in chatMessages" :key="i" class="chat-msg" :class="msg.role">
            {{ msg.content }}
          </div>
          <div v-if="chatLoading" class="chat-msg assistant loading-msg">
            AI 正在回复...
          </div>
        </div>
        <div class="chat-input-row">
          <input v-model="chatInput" class="chat-input" placeholder="追问关于代码的问题..." @keyup.enter="chat" :disabled="chatLoading" />
          <button class="send-btn" @click="chat" :disabled="!chatInput.trim() || chatLoading">发送</button>
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
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
.lang-select { padding: 8px 12px; border: 2px solid var(--border); border-radius: 8px; font-size: 14px; background: white; }
.tool-btn { padding: 8px 16px; border: 2px solid var(--border); border-radius: 8px; font-size: 14px; font-weight: 600; background: white; cursor: pointer; transition: all 0.2s; }
.tool-btn:hover { border-color: var(--primary); }
.tool-btn.primary { background: var(--primary); color: white; border-color: var(--primary); }
.tool-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.editor-section { margin-bottom: 12px; }
.code-editor { width: 100%; min-height: 200px; border: 2px solid var(--border); border-radius: 12px; padding: 12px; font-family: 'Fira Code', monospace; font-size: 14px; resize: vertical; line-height: 1.6; }
.question-input { width: 100%; padding: 10px 14px; border: 2px solid var(--border); border-radius: 8px; font-size: 14px; margin-bottom: 16px; }
.loading-section { text-align: center; padding: 30px; }
.loading-spinner { width: 40px; height: 40px; border: 4px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 1s linear infinite; margin: 0 auto 12px; }
@keyframes spin { to { transform: rotate(360deg); } }
.error-msg { color: #ef4444; text-align: center; padding: 12px; }
.result-box { background: white; border: 2px solid var(--border); border-radius: 12px; padding: 16px; margin-bottom: 16px; }
.result-box h3 { font-size: 16px; font-weight: 800; margin-bottom: 12px; }
.output { background: #1e293b; color: #e2e8f0; border-radius: 8px; padding: 12px; font-size: 13px; overflow-x: auto; white-space: pre-wrap; }
.output.has-error { color: #fca5a5; }
.review-output { white-space: pre-wrap; }
.diagnosis { display: flex; flex-direction: column; gap: 8px; }
.diag-item { font-size: 14px; line-height: 1.6; }
.fix-section { margin-top: 12px; }
.fix-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.fix-header h4 { font-size: 14px; font-weight: 700; }
.use-btn { padding: 4px 12px; background: var(--primary); color: white; border: none; border-radius: 6px; font-size: 12px; cursor: pointer; }
.fix-code { background: #1e293b; color: #86efac; border-radius: 8px; padding: 12px; font-size: 13px; overflow-x: auto; white-space: pre; }
.chat-section { margin-top: 24px; }
.chat-section h3 { font-size: 16px; font-weight: 800; margin-bottom: 12px; }
.chat-messages { max-height: 300px; overflow-y: auto; margin-bottom: 12px; }
.chat-msg { padding: 10px 14px; border-radius: 12px; margin-bottom: 8px; font-size: 14px; line-height: 1.6; }
.chat-msg.user { background: var(--primary-bg); margin-left: 40px; }
.chat-msg.assistant { background: white; border: 1px solid var(--border); margin-right: 40px; }
.chat-msg.loading-msg { opacity: 0.6; }
.chat-input-row { display: flex; gap: 8px; }
.chat-input { flex: 1; padding: 10px 14px; border: 2px solid var(--border); border-radius: 8px; font-size: 14px; }
.send-btn { padding: 10px 20px; background: var(--primary); color: white; border: none; border-radius: 8px; font-weight: 700; cursor: pointer; }
.send-btn:disabled { opacity: 0.5; }
</style>
