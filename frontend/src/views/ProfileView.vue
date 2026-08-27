<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { userApi } from '@/api/client'

const auth = useAuthStore()
const router = useRouter()
const todayPercent = ref(0)

onMounted(async () => {
  try {
    const res = await userApi.stats()
    auth.setStats(res.data)
    todayPercent.value = Math.min(100, Math.round((res.data.today_xp / res.data.daily_goal) * 100))
  } catch (e) {
    // ignore
  }
})

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="container profile-page">
    <div class="profile-header">
      <div class="avatar">{{ auth.user?.username?.charAt(0).toUpperCase() || '?' }}</div>
      <h1 class="username">{{ auth.user?.username }}</h1>
      <p class="email">{{ auth.user?.email }}</p>
    </div>

    <div class="stats-grid">
      <div class="stat-card xp-card">
        <span class="stat-icon">💎</span>
        <span class="stat-value">{{ auth.stats.xp }}</span>
        <span class="stat-label">总经验值</span>
      </div>
      <div class="stat-card streak-card">
        <span class="stat-icon">🔥</span>
        <span class="stat-value">{{ auth.stats.streak_days }}</span>
        <span class="stat-label">连续天数</span>
      </div>
      <div class="stat-card hearts-card">
        <span class="stat-icon">❤️</span>
        <span class="stat-value">{{ auth.stats.hearts }}</span>
        <span class="stat-label">剩余心数</span>
      </div>
      <div class="stat-card completed-card">
        <span class="stat-icon">📚</span>
        <span class="stat-value">{{ auth.stats.completed_today }}</span>
        <span class="stat-label">已完成课程</span>
      </div>
    </div>

    <div class="daily-goal">
      <h2 class="section-title">今日目标</h2>
      <div class="goal-progress">
        <div class="goal-bar">
          <div class="goal-fill" :style="{ width: todayPercent + '%' }"></div>
        </div>
        <span class="goal-text">{{ auth.stats.today_xp }} / {{ auth.stats.daily_goal }} XP</span>
      </div>
    </div>

    <div class="actions">
      <router-link to="/" class="btn-secondary">继续学习</router-link>
      <button class="btn-ghost" @click="logout">退出登录</button>
    </div>
  </div>
</template>

<style scoped>
.profile-page { padding-top: 24px; padding-bottom: 40px; }

.profile-header { text-align: center; margin-bottom: 32px; }
.avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--primary);
  color: white;
  font-size: 36px;
  font-weight: 900;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 12px;
}
.username { font-size: 24px; }
.email { color: var(--text-light); font-size: 14px; }

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 32px;
}

.stat-card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  text-align: center;
  box-shadow: 0 3px 0 var(--border);
}

.stat-icon { font-size: 32px; }
.stat-value { font-size: 28px; font-weight: 900; }
.stat-label { font-size: 12px; color: var(--text-light); }

.xp-card { border-color: var(--secondary); box-shadow: 0 3px 0 var(--secondary); }
.xp-card .stat-value { color: var(--secondary); }

.streak-card { border-color: var(--warning); box-shadow: 0 3px 0 var(--warning); }
.streak-card .stat-value { color: var(--warning); }

.hearts-card { border-color: var(--danger); box-shadow: 0 3px 0 var(--danger); }
.hearts-card .stat-value { color: var(--danger); }

.daily-goal { margin-bottom: 32px; }
.section-title { font-size: 18px; margin-bottom: 12px; }

.goal-progress {
  display: flex;
  align-items: center;
  gap: 12px;
}

.goal-bar {
  flex: 1;
  height: 24px;
  background: var(--bg-gray);
  border-radius: 12px;
  overflow: hidden;
  border: 2px solid var(--border);
}

.goal-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 12px;
  transition: width 0.5s;
  min-width: 4px;
}

.goal-text { font-size: 14px; color: var(--text-light); white-space: nowrap; }

.actions {
  display: flex;
  gap: 12px;
}
.actions > * { flex: 1; text-align: center; }
</style>
