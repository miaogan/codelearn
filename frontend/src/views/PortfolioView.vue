<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { portfolioApi } from '@/api/client'
import type { Portfolio } from '@/types'

const route = useRoute()
const auth = useAuthStore()
const loading = ref(true)
const portfolio = ref<Portfolio | null>(null)
const shareUrl = ref('')

const isShared = computed(() => !!route.params.username)
const backPath = computed(() => (isShared.value ? '/' : '/profile'))

onMounted(async () => {
  const username = (route.params.username as string) || auth.user?.username
  if (!username) return
  shareUrl.value = `${window.location.origin}/portfolio/${username}`
  try {
    const res = await portfolioApi.get(username)
    portfolio.value = res.data
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
})

async function copyLink() {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    alert('链接已复制，可分享给他人查看你的能力档案')
  } catch (e) {
    // ignore
  }
}

function masterColor(m: number) {
  if (m >= 80) return '#10b981'
  if (m >= 50) return '#f59e0b'
  return '#ef4444'
}
</script>

<template>
  <div class="portfolio-page">
    <div class="header">
      <router-link :to="backPath" class="back-btn">← 返回</router-link>
      <h1>📇 数字能力档案</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <template v-else-if="portfolio">
      <div class="card-head">
        <div class="avatar">{{ portfolio.username.charAt(0).toUpperCase() }}</div>
        <div class="head-info">
          <div class="head-name">{{ portfolio.username }}</div>
          <div class="head-title">{{ portfolio.title_icon }} {{ portfolio.title_name }} · Lv.{{ portfolio.title_level }}</div>
        </div>
      <div class="head-share" v-if="!isShared" @click="copyLink">🔗 分享</div>
      </div>

      <div class="stats-grid">
        <div class="stat-card">
          <span class="stat-icon">💎</span>
          <span class="stat-value">{{ portfolio.xp }}</span>
          <span class="stat-label">总经验</span>
        </div>
        <div class="stat-card">
          <span class="stat-icon">🔥</span>
          <span class="stat-value">{{ portfolio.streak_days }}</span>
          <span class="stat-label">连续天数</span>
        </div>
        <div class="stat-card">
          <span class="stat-icon">🏅</span>
          <span class="stat-value">{{ portfolio.badge_count }}</span>
          <span class="stat-label">成就徽章</span>
        </div>
      </div>

      <div class="section">
        <h2 class="section-title">🎓 课程能力</h2>
        <div v-if="portfolio.courses.length === 0" class="empty">暂未开始学习任何课程</div>
        <div class="course-list">
          <div v-for="c in portfolio.courses" :key="c.course_id" class="course-item">
            <div class="course-head">
              <span class="course-emoji">{{ c.emoji }}</span>
              <div class="course-body">
                <div class="course-title">
                  {{ c.course_title }}
                  <span v-if="c.certified" class="cert-badge">🎖️ {{ c.cert_level }}认证</span>
                </div>
                <div class="course-meta">单元 {{ c.units_passed }}/{{ c.units_total }} · 项目 {{ c.project_count }}</div>
              </div>
              <span class="master-num" :style="{ color: masterColor(c.mastery) }">{{ c.mastery }}%</span>
            </div>
            <div class="master-bar">
              <div class="master-fill" :style="{ width: c.mastery + '%', background: masterColor(c.mastery) }"></div>
            </div>
          </div>
        </div>
      </div>

      <div class="section">
        <h2 class="section-title">🛠️ 项目作品</h2>
        <div v-if="portfolio.projects.length === 0" class="empty">暂无项目作品，快去项目工坊创作吧</div>
        <div class="project-list">
          <div v-for="p in portfolio.projects" :key="p.id" class="project-item">
            <span class="project-icon">{{ p.status === 'completed' ? '✅' : '📝' }}</span>
            <div class="project-body">
              <div class="project-title">{{ p.title }}</div>
              <div class="project-meta">{{ p.language }} · {{ p.status === 'completed' ? '已完成' : '进行中' }}</div>
            </div>
          </div>
        </div>
      </div>

      <router-link to="/projects" class="btn-primary">前往项目工坊</router-link>
    </template>

    <div v-else class="loading">暂无档案数据</div>
  </div>
</template>

<style scoped>
.portfolio-page { max-width: 700px; margin: 0 auto; padding: 20px; }
.header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.back-btn { color: var(--text-light); text-decoration: none; font-size: 16px; }
.header h1 { font-size: 28px; font-weight: 900; }
.loading, .empty { text-align: center; padding: 40px 20px; color: var(--text-light); }

.card-head {
  display: flex;
  align-items: center;
  gap: 14px;
  background: linear-gradient(135deg, var(--primary), #6366f1);
  border-radius: var(--radius);
  color: white;
  padding: 24px;
  margin-bottom: 16px;
  box-shadow: 0 4px 0 var(--primary-dark);
}
.avatar {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: rgba(255,255,255,0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 900;
  flex-shrink: 0;
}
.head-info { flex: 1; min-width: 0; }
.head-name { font-size: 20px; font-weight: 900; }
.head-title { font-size: 13px; opacity: 0.9; }
.head-share {
  background: rgba(255,255,255,0.2);
  border-radius: 16px;
  padding: 5px 12px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  flex-shrink: 0;
}

.stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 16px; }
.stat-card {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 16px 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  box-shadow: 0 3px 0 var(--border);
}
.stat-icon { font-size: 22px; }
.stat-value { font-size: 22px; font-weight: 900; }
.stat-label { font-size: 11px; color: var(--text-light); }

.section {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 20px;
  margin-bottom: 16px;
}
.section-title { font-size: 16px; font-weight: 800; margin-bottom: 14px; }

.course-list { display: flex; flex-direction: column; gap: 14px; }
.course-head { display: flex; align-items: center; gap: 12px; margin-bottom: 6px; }
.course-emoji { font-size: 28px; }
.course-body { flex: 1; min-width: 0; }
.course-title { font-size: 15px; font-weight: 800; display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.cert-badge { font-size: 11px; background: #fef3c7; color: #b45309; padding: 2px 8px; border-radius: 10px; font-weight: 700; }
.course-meta { font-size: 12px; color: var(--text-light); }
.master-num { font-size: 18px; font-weight: 900; }
.master-bar { height: 8px; background: var(--bg-gray); border-radius: 4px; overflow: hidden; }
.master-fill { height: 100%; border-radius: 4px; transition: width 0.6s; }

.project-list { display: flex; flex-direction: column; gap: 10px; }
.project-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  background: var(--bg-gray);
  border-radius: 10px;
}
.project-icon { font-size: 18px; }
.project-body { flex: 1; min-width: 0; }
.project-title { font-size: 14px; font-weight: 700; }
.project-meta { font-size: 12px; color: var(--text-light); }

.btn-primary {
  display: block;
  text-align: center;
  text-decoration: none;
  padding: 12px 20px;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 800;
  background: var(--primary);
  color: white;
  box-shadow: 0 4px 0 var(--primary-dark);
}
</style>
