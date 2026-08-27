<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { userApi } from '@/api/client'

const auth = useAuthStore()
const router = useRouter()
const account = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(account.value, password.value)
    const res = await userApi.stats()
    auth.setStats(res.data)
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="container auth-page">
    <h1 class="title">📚 CodeLearn</h1>
    <p class="subtitle">登录开始你的编程学习之旅</p>
    <form @submit.prevent="handleLogin" class="auth-form">
      <div class="field">
        <label>用户名或邮箱</label>
        <input v-model="account" type="text" placeholder="输入用户名或邮箱" required />
      </div>
      <div class="field">
        <label>密码</label>
        <input v-model="password" type="password" placeholder="输入密码" required />
      </div>
      <p v-if="error" class="error">{{ error }}</p>
      <button type="submit" class="btn-primary" :disabled="loading">
        {{ loading ? '登录中...' : '登录' }}
      </button>
    </form>
    <p class="switch-link">
      还没有账号？<router-link to="/register">立即注册</router-link>
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

.title {
  font-size: 36px;
  color: var(--primary);
  margin-bottom: 8px;
}

.subtitle {
  color: var(--text-light);
  margin-bottom: 32px;
  font-weight: 700;
}

.auth-form {
  width: 100%;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field label {
  font-size: 14px;
  color: var(--text-light);
  font-weight: 800;
}

.error {
  color: var(--danger);
  text-align: center;
  font-size: 14px;
}

.auth-form .btn-primary {
  margin-top: 8px;
  width: 100%;
}

.switch-link {
  margin-top: 24px;
  color: var(--text-light);
}

.switch-link a {
  color: var(--secondary);
  font-weight: 800;
}
</style>
