<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { feedbackApi } from '@/api/client'

const router = useRouter()
const category = ref('bug')
const content = ref('')
const contact = ref('')
const submitting = ref(false)
const done = ref(false)
const error = ref('')

async function submit() {
  if (!content.value.trim()) {
    error.value = '请填写反馈内容'
    return
  }
  error.value = ''
  submitting.value = true
  try {
    await feedbackApi.submit({
      category: category.value,
      content: content.value.trim(),
      contact: contact.value.trim(),
    })
    done.value = true
  } catch (e: any) {
    error.value = e.response?.data?.error || '提交失败，请稍后再试'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="container feedback-page">
    <h1 class="page-title">📮 问题反馈</h1>
    <p class="page-sub">遇到问题或有建议？告诉我们，我们会尽快处理。</p>

    <div v-if="done" class="card success-card">
      <div class="success-icon">🎉</div>
      <h2>感谢你的反馈！</h2>
      <p class="success-sub">我们会认真对待每一条建议</p>
      <button class="btn-primary" @click="router.push('/')">返回首页</button>
    </div>

    <form v-else @submit.prevent="submit" class="card feedback-form">
      <div class="field">
        <label>反馈类型</label>
        <select v-model="category" class="select-input">
          <option value="bug">🐛 问题/Bug 上报</option>
          <option value="suggestion">💡 功能建议</option>
          <option value="content">📚 内容纠错</option>
          <option value="other">其他</option>
        </select>
      </div>
      <div class="field">
        <label>反馈内容</label>
        <textarea
          v-model="content"
          class="content-input"
          rows="5"
          placeholder="请描述你遇到的问题或建议（如：单元考试第 3 题解析有误...）"
          required
        ></textarea>
      </div>
      <div class="field">
        <label>联系方式（选填）</label>
        <input v-model="contact" class="text-input" placeholder="邮箱 / 微信号，便于我们回复你" />
      </div>
      <p v-if="error" class="error">{{ error }}</p>
      <button type="submit" class="btn-primary" :disabled="submitting">
        {{ submitting ? '提交中...' : '提交反馈' }}
      </button>
    </form>

    <button class="btn-ghost back-btn" @click="router.back()">← 返回</button>
  </div>
</template>

<style scoped>
.feedback-page { padding-top: 24px; padding-bottom: 40px; }
.page-title { font-size: 24px; margin-bottom: 4px; }
.page-sub { color: var(--text-light); font-size: 14px; margin-bottom: 20px; }

.card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius);
  padding: 24px;
  box-shadow: 0 4px 0 #e5e5e5;
}

.feedback-form { display: flex; flex-direction: column; gap: 16px; }
.field { display: flex; flex-direction: column; gap: 6px; }
.field label { font-size: 14px; color: var(--text-light); font-weight: 800; }
.select-input, .text-input, .content-input {
  padding: 12px 14px;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 15px;
  outline: none;
  font-family: inherit;
}
.select-input:focus, .text-input:focus, .content-input:focus { border-color: var(--primary); }
.content-input { resize: vertical; }
.error { color: var(--danger); font-size: 14px; text-align: center; }
.btn-primary {
  padding: 12px 20px;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  background: var(--primary);
  color: white;
  box-shadow: 0 4px 0 var(--primary-dark);
}
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.success-card { text-align: center; padding: 40px 24px; }
.success-icon { font-size: 56px; margin-bottom: 8px; }
.success-card h2 { font-size: 20px; margin-bottom: 6px; }
.success-sub { color: var(--text-light); font-size: 14px; margin-bottom: 20px; }

.btn-ghost { background: transparent; color: var(--text-light); border: none; margin-top: 20px; width: 100%; font-weight: 800; cursor: pointer; }
</style>
