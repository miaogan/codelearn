import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/client'
import type { User } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)
  const stats = ref({
    xp: 0,
    streak_days: 0,
    hearts: 5,
    max_hearts: 5,
    daily_goal: 50,
    today_xp: 0,
    completed_today: 0,
  })

  const isLoggedIn = computed(() => !!token.value)

  async function register(username: string, email: string, password: string) {
    const res = await authApi.register({ username, email, password })
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
  }

  async function login(account: string, password: string) {
    const res = await authApi.login({ account, password })
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  function setStats(s: typeof stats.value) {
    stats.value = s
  }

  return { token, user, stats, isLoggedIn, register, login, logout, setStats }
})
