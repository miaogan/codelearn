<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { leaderboardApi } from '@/api/client'
import type { LeaderboardEntry } from '@/types'

const router = useRouter()
const entries = ref<LeaderboardEntry[]>([])
const me = ref<LeaderboardEntry | null>(null)
const loading = ref(true)
const error = ref('')

const medals = ['🥇', '🥈', '🥉']

onMounted(async () => {
  try {
    const res = await leaderboardApi.weekly()
    entries.value = res.data.entries
    me.value = res.data.me
  } catch (e) {
    error.value = '加载排行榜失败，请稍后重试'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="container leaderboard-page">
    <div class="lb-header">
      <h1 class="lb-title">🏆 本周排行榜</h1>
      <p class="lb-sub">每周一结算，通过完成课时、答对练习和通过考试获得 XP</p>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else>
      <!-- 我的排名 -->
      <div class="me-card" v-if="me">
        <span class="me-rank">#{{ me.rank }}</span>
        <span class="me-label">我的本周 XP</span>
        <span class="me-xp">{{ me.xp }} XP</span>
      </div>

      <!-- 排行榜 -->
      <div class="lb-list">
        <div
          v-for="(e, i) in entries"
          :key="e.user_id"
          class="lb-item"
          :class="{ top3: i < 3 }"
        >
          <span class="lb-rank">
            <span v-if="i < 3" class="lb-medal">{{ medals[i] }}</span>
            <span v-else class="lb-num">{{ e.rank }}</span>
          </span>
          <span class="lb-username">{{ e.username }}</span>
          <span class="lb-xp">{{ e.xp }} XP</span>
        </div>
      </div>

      <div v-if="entries.length === 0" class="empty">本周还没有人上榜，快去学习吧！🚀</div>
    </div>

    <button class="btn-ghost back-btn" @click="router.push('/')">← 返回首页</button>
  </div>
</template>

<style scoped>
.leaderboard-page { padding-top: 24px; padding-bottom: 40px; }

.lb-header { text-align: center; margin-bottom: 28px; }
.lb-title { font-size: 26px; font-weight: 900; margin-bottom: 6px; }
.lb-sub { color: var(--text-light); font-size: 14px; }

.me-card {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border-radius: var(--radius);
  padding: 16px 24px;
  margin-bottom: 20px;
  box-shadow: 0 4px 0 #4f46e5;
}
.me-rank { font-size: 22px; font-weight: 900; }
.me-label { font-size: 14px; opacity: 0.9; }
.me-xp { font-size: 20px; font-weight: 900; }

.lb-list { display: flex; flex-direction: column; gap: 8px; }
.lb-item {
  display: flex;
  align-items: center;
  gap: 14px;
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 12px 16px;
}
.lb-item.top3 { border-color: #fbbf24; }

.lb-rank { width: 40px; text-align: center; flex-shrink: 0; }
.lb-medal { font-size: 24px; }
.lb-num { font-size: 18px; font-weight: 800; color: var(--text-light); }
.lb-username { flex: 1; font-weight: 700; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.lb-xp { font-weight: 900; color: var(--secondary); }

.empty { text-align: center; color: var(--text-light); padding: 40px; }
.loading, .error { text-align: center; padding: 40px; color: var(--text-light); }

.back-btn { margin-top: 32px; width: 100%; }
</style>
