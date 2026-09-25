<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { achievementApi } from '@/api/client'
import type { AchievementSummary, AchievementBadge } from '@/types'

const loading = ref(true)
const summary = ref<AchievementSummary | null>(null)

const BADGE_DEFS = [
  { code: 'first_lesson', title: '初出茅庐', icon: '🎓', description: '完成第一个课时' },
  { code: 'streak_7', title: '七日之约', icon: '🔥', description: '连续打卡 7 天' },
  { code: 'streak_30', title: '月度战士', icon: '👑', description: '连续打卡 30 天' },
  { code: 'first_exam_pass', title: '首战告捷', icon: '🏅', description: '首次通过单元考试' },
  { code: 'perfect_exam', title: '满分学霸', icon: '💯', description: '单场考试满分' },
  { code: 'first_cert', title: '能力认证', icon: '🎖️', description: '首次获得能力认证证书' },
  { code: 'practice_50', title: '勤学苦练', icon: '✏️', description: '累计练习答对 50 题' },
  { code: 'practice_200', title: '千锤百炼', icon: '🚀', description: '累计练习答对 200 题' },
]

const TITLE_LEVELS = [
  { min: 0, name: '代码萌芽', icon: '🌱' },
  { min: 100, name: '代码学徒', icon: '🧑‍🎓' },
  { min: 300, name: '代码骑士', icon: '⚔️' },
  { min: 600, name: '代码剑士', icon: '🗡️' },
  { min: 1000, name: '代码大师', icon: '🧙' },
  { min: 2000, name: '代码宗师', icon: '🏆' },
]

const unlockedMap = computed(() => {
  const map = new Map<string, AchievementBadge>()
  for (const b of summary.value?.badges || []) {
    map.set(b.code, b)
  }
  return map
})

const title = computed(() => summary.value?.title)

const titleProgress = computed(() => {
  const t = title.value
  if (!t) return 0
  const cur = TITLE_LEVELS[t.level]
  const next = TITLE_LEVELS[t.level + 1]
  if (!cur || !next) return 100
  const span = next.min - cur.min
  if (span <= 0) return 100
  const got = Math.max(0, t.current_xp - cur.min)
  return Math.min(100, Math.max(2, Math.round((got / span) * 100)))
})

const unlockedCount = computed(() => summary.value?.total || 0)

onMounted(async () => {
  try {
    const res = await achievementApi.summary()
    summary.value = res.data
  } catch (e) {
    console.error('加载成就失败', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="ach-page">
    <div class="header">
      <router-link to="/profile" class="back-btn">← 返回</router-link>
      <h1>🏆 成就与段位</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <template v-else-if="title">
      <!-- 段位称号卡片 -->
      <div class="title-card">
        <div class="title-icon">{{ title.icon }}</div>
        <div class="title-info">
          <div class="title-name">{{ title.name }}</div>
          <div class="title-meta">
            Lv.{{ title.level + 1 }} · 当前 {{ title.current_xp }} XP
            <template v-if="title.next_name"> · 距「{{ title.next_name }}」还需 {{ title.xp_to_next }} XP</template>
            <template v-else> · 已达最高段位</template>
          </div>
          <div class="title-bar">
            <div class="title-fill" :style="{ width: titleProgress + '%' }"></div>
          </div>
        </div>
      </div>

      <!-- 徽章墙 -->
      <div class="badge-section">
        <h2 class="section-title">徽章墙 · {{ unlockedCount }} / {{ BADGE_DEFS.length }}</h2>
        <div class="badge-grid">
          <div
            v-for="def in BADGE_DEFS"
            :key="def.code"
            class="badge-card"
            :class="{ locked: !unlockedMap.has(def.code) }"
          >
            <div class="badge-icon">{{ def.icon }}</div>
            <div class="badge-title">{{ def.title }}</div>
            <div class="badge-desc">{{ def.description }}</div>
            <div v-if="unlockedMap.has(def.code)" class="badge-time">
              {{ new Date(unlockedMap.get(def.code)!.unlocked_at).toLocaleDateString() }} 解锁
            </div>
            <div v-else class="badge-lock">🔒 未解锁</div>
          </div>
        </div>
      </div>
    </template>

    <div v-else class="empty">暂无成就数据</div>
  </div>
</template>

<style scoped>
.ach-page {
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
  color: var(--text-secondary, #777);
  text-decoration: none;
  font-size: 16px;
}

.header h1 {
  font-size: 28px;
  font-weight: 900;
}

.loading, .empty {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-light);
}

.title-card {
  display: flex;
  align-items: center;
  gap: 20px;
  background: white;
  border: 3px solid var(--secondary);
  border-radius: var(--radius);
  padding: 24px;
  margin-bottom: 32px;
  box-shadow: 0 4px 0 #0a8acb;
}

.title-icon {
  font-size: 56px;
  flex-shrink: 0;
}

.title-info { flex: 1; min-width: 0; }

.title-name {
  font-size: 26px;
  font-weight: 900;
  color: var(--secondary);
}

.title-meta {
  font-size: 14px;
  color: var(--text-light);
  margin: 6px 0 12px;
}

.title-bar {
  height: 16px;
  background: var(--bg-gray);
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid var(--border);
}

.title-fill {
  height: 100%;
  background: var(--secondary);
  border-radius: 8px;
  transition: width 0.5s;
}

.badge-section { margin-bottom: 32px; }
.section-title { font-size: 18px; font-weight: 800; margin-bottom: 16px; }

.badge-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.badge-card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 20px 14px;
  text-align: center;
  box-shadow: 0 3px 0 var(--border);
}

.badge-card.locked {
  opacity: 0.55;
  background: var(--bg-gray);
  box-shadow: none;
}

.badge-icon { font-size: 40px; margin-bottom: 8px; }

.badge-title { font-size: 16px; font-weight: 800; margin-bottom: 4px; }

.badge-desc { font-size: 12px; color: var(--text-light); line-height: 1.4; min-height: 32px; }

.badge-time {
  margin-top: 8px;
  font-size: 11px;
  color: var(--primary-dark);
  font-weight: 700;
}

.badge-lock {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-light);
}

@media (max-width: 480px) {
  .badge-grid { grid-template-columns: 1fr; }
}
</style>
