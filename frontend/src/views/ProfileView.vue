<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { userApi, leaderboardApi } from '@/api/client'
import type { CalendarDay } from '@/types'

const auth = useAuthStore()
const router = useRouter()
const todayPercent = ref(0)
const calendarDays = ref<CalendarDay[]>([])
const calendarYear = ref(0)
const calendarMonth = ref(0)

onMounted(async () => {
  try {
    const [statsRes, calRes] = await Promise.all([
      userApi.stats(),
      leaderboardApi.calendar(),
    ])
    auth.setStats(statsRes.data)
    todayPercent.value = Math.min(100, Math.round((statsRes.data.today_xp / statsRes.data.daily_goal) * 100))
    calendarDays.value = calRes.data.days
    calendarYear.value = calRes.data.year
    calendarMonth.value = calRes.data.month
  } catch (e) {
    // ignore
  }
})

const xpByDate = computed(() => {
  const map = new Map<string, number>()
  for (const d of calendarDays.value) {
    map.set(d.date, d.xp)
  }
  return map
})

// 生成当月日历网格（周为行，天为列）
const monthGrid = computed(() => {
  const first = new Date(calendarYear.value, calendarMonth.value - 1, 1)
  const daysInMonth = new Date(calendarYear.value, calendarMonth.value, 0).getDate()
  const startOffset = first.getDay() // 0=周日
  const cells: { day: number; xp: number; hasXP: boolean }[] = []
  for (let i = 0; i < startOffset; i++) {
    cells.push({ day: 0, xp: 0, hasXP: false })
  }
  for (let d = 1; d <= daysInMonth; d++) {
    const key = `${calendarYear.value}-${String(calendarMonth.value).padStart(2, '0')}-${String(d).padStart(2, '0')}`
    const xp = xpByDate.value.get(key) || 0
    cells.push({ day: d, xp, hasXP: xp > 0 })
  }
  return cells
})

function cellClass(c: { hasXP: boolean; xp: number }) {
  if (!c.hasXP) return 'cal-day empty'
  if (c.xp >= 80) return 'cal-day hot'
  if (c.xp >= 40) return 'cal-day mid'
  return 'cal-day low'
}

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

    <!-- 学习日历 -->
    <div class="calendar-section">
      <h2 class="section-title">学习日历 · {{ calendarYear }} 年 {{ calendarMonth }} 月</h2>
      <div class="cal-weekdays">
        <span v-for="w in ['日', '一', '二', '三', '四', '五', '六']" :key="w">{{ w }}</span>
      </div>
      <div class="cal-grid">
        <div
          v-for="(c, i) in monthGrid"
          :key="i"
          class="cal-day"
          :class="cellClass(c)"
          :title="c.day ? `${c.day}日 · ${c.xp} XP` : ''"
        >
          <span v-if="c.day">{{ c.day }}</span>
        </div>
      </div>
      <p class="cal-legend">坚持每天学习，点亮你的日历 🔥</p>
    </div>

    <div class="actions">
      <router-link to="/" class="btn-secondary">继续学习</router-link>
      <router-link to="/leaderboard" class="btn-secondary">排行榜</router-link>
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

/* 学习日历 */
.calendar-section {
  margin-bottom: 32px;
}
.cal-weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  text-align: center;
  font-size: 12px;
  color: var(--text-light);
  margin-bottom: 6px;
}
.cal-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 6px;
}
.cal-day {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  border-radius: 8px;
  background: var(--bg-gray);
  color: var(--text-light);
}
.cal-day.empty { color: transparent; background: transparent; }
.cal-day.low { background: #dbeafe; color: var(--primary-dark); }
.cal-day.mid { background: #93c5fd; color: white; }
.cal-day.hot { background: #3b82f6; color: white; }
.cal-legend { margin-top: 10px; font-size: 12px; color: var(--text-light); }
</style>
