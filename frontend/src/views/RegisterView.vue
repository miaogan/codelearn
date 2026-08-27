<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { userApi } from '@/api/client'

const auth = useAuthStore()
const router = useRouter()
const username = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleRegister() {
  error.value = ''
  loading.value = true
  try {
    await auth.register(username.value, email.value, password.value)
    const res = await userApi.stats()
    auth.setStats(res.data)
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="container auth-page">
    <h1 class="title">📚 CodeLearn</h1>
    <p class="subtitle">创建账号，开始学习编程</p>
    <form @submit.prevent="handleRegister" class="auth-form">
      <div class="field">
        <label>用户名</label>
        <input v-model="username" type="text" placeholder="输入用户名" required minlength="2" />
      </div>
      <div class="field">
        <label>邮箱</label>
        <input v-model="email" type="email" placeholder="输入邮箱" required />
      </div>
      <div class="field">
        <label>密码</label>
        <input v-model="password" type="password" placeholder="至少 6 位" required minlength="6" />
      </div>
      <p v-if="error" class="error">{{ error }}</p>
      <button type="submit" class="btn-primary" :disabled="loading">
        {{ loading ? '注册中...' : '注册' }}
      </button>
    </form>
    <p class="switch-link">
      已有账号？<router-link to="/login">去登录</router-link>
    </p>
  </div>
</template>

<style scoped>
.auth-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 80vh;
  padding-top: 40px;
}
.title { font-size: 36px; color: var(--primary); margin-bottom: 8px; }
.subtitle { color: var(--text-light); margin-bottom: 32px; font-weight: 700; }
.auth-form { width: 100%; max-width: 400px; display: flex; flex-direction: column; gap: 16px; }
.field { display: flex; flex-direction: column; gap: 6px; }
.field label { font-size: 14px; color: var(--text-light); font-weight: 800; }
.error { color: var(--danger); text-align: center; font-size: 14px; }
.auth-form .btn-primary { margin-top: 8px; width: 100%; }
.switch-link { margin-top: 24px; color: var(--text-light); }
.switch-link a { color: var(--secondary); font-weight: 800; }
</style>
