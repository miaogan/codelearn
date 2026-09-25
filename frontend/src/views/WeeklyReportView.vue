<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { weeklyReportApi } from '@/api/client'
import type { WeeklyReport } from '@/types'

const loading = ref(true)
const report = ref<WeeklyReport | null>(null)

onMounted(async () => {
  try {
    const res = await weeklyReportApi.get()
    report.value = res.data
  } catch (e) {
    console.error('加载周报失败', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="report-page">
    <div class="header">
      <router-link to="/profile" class="back-btn">← 返回</router-link>
      <h1>📊 学习周报</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <template v-else-if="report">
      <div class="report-hero">
        <div class="week-range">{{ report.week_start }} ~ {{ report.week_end }}</div>
        <div class="hero-xp">{{ report.xp_total }} <span class="unit">XP</span></div>
        <div class="hero-label">本周累计经验</div>
      </div>

      <div class="stats-grid">
        <div class="stat-card">
          <span class="stat-icon">📅</span>
          <span class="stat-value">{{ report.active_days }}</span>
          <span class="stat-label">学习天数</span>
        </div>
        <div class="stat-card">
          <span class="stat-icon">📚</span>
          <span class="stat-value">{{ report.lessons_completed }}</span>
          <span class="stat-label">完成课时</span>
        </div>
        <div class="stat-card">
          <span class="stat-icon">✏️</span>
          <span class="stat-value">{{ report.exercises_done }}</span>
          <span class="stat-label">练习题目</span>
        </div>
        <div class="stat-card">
          <span class="stat-icon">🏅</span>
          <span class="stat-value">{{ report.exams_passed }}</span>
          <span class="stat-label">通过考试</span>
        </div>
        <div class="stat-card">
          <span class="stat-icon">⚠️</span>
          <span class="stat-value">{{ report.wrong_added }}</span>
          <span class="stat-label">新增错题</span>
        </div>
        <div class="stat-card">
          <span class="stat-icon">🔥</span>
          <span class="stat-value">{{ report.streak_days }}</span>
          <span class="stat-label">连续打卡</span>
        </div>
      </div>

      <div class="section">
        <h2 class="section-title">📝 本周总结</h2>
        <p class="summary-text">{{ report.summary }}</p>
      </div>

      <div class="section" v-if="report.weak_topics.length > 0">
        <h2 class="section-title">🧠 薄弱知识点 TOP{{ report.weak_topics.length }}</h2>
        <div class="weak-list">
          <div v-for="t in report.weak_topics" :key="t.lesson_id" class="weak-item">
            <span class="weak-icon">📌</span>
            <span class="weak-title">{{ t.title || '课时 ' + t.lesson_id }}</span>
            <span class="weak-count">错 {{ t.wrong_count }} 题</span>
          </div>
        </div>
      </div>

      <div class="section advice">
        <h2 class="section-title">💡 下周建议</h2>
        <p class="advice-text">{{ report.advice }}</p>
      </div>

      <div class="actions">
        <router-link to="/" class="btn-primary">继续学习</router-link>
        <router-link to="/srs-reviews" class="btn-secondary">去复习</router-link>
      </div>
    </template>

    <div v-else class="empty">暂无周报数据</div>
  </div>
</template>

<style scoped>
.report-page {
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
  color: var(--text-light);
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

.report-hero {
  background: linear-gradient(135deg, var(--secondary), #0a8acb);
  border-radius: var(--radius);
  color: white;
  text-align: center;
  padding: 32px 20px;
  margin-bottom: 20px;
  box-shadow: 0 4px 0 #0a8acb;
}

.week-range {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 8px;
}

.hero-xp {
  font-size: 48px;
  font-weight: 900;
}

.hero-xp .unit { font-size: 20px; font-weight: 700; }

.hero-label { font-size: 14px; opacity: 0.9; }

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}

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

.stat-icon { font-size: 24px; }
.stat-value { font-size: 22px; font-weight: 900; }
.stat-label { font-size: 11px; color: var(--text-light); }

.section {
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 20px;
  margin-bottom: 16px;
}

.section-title { font-size: 16px; font-weight: 800; margin-bottom: 10px; }

.summary-text, .advice-text {
  font-size: 14px;
  line-height: 1.7;
  color: var(--text);
}

.weak-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.weak-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: var(--bg-gray);
  border-radius: 10px;
}

.weak-icon { font-size: 16px; }
.weak-title { flex: 1; font-size: 14px; font-weight: 700; }
.weak-count { font-size: 12px; font-weight: 800; color: var(--danger); }

.section.advice {
  border-color: #6ee7b7;
  background: #ecfdf5;
}

.actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.actions a {
  flex: 1;
  text-align: center;
  text-decoration: none;
}

@media (max-width: 480px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
