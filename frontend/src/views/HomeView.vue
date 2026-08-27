<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { courseApi } from '@/api/client'
import type { Course } from '@/types'

const router = useRouter()
const courses = ref<Course[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await courseApi.list()
    courses.value = res.data.courses
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="container home-page">
    <div class="hero">
      <h1 class="hero-title">选择你的编程语言</h1>
      <p class="hero-subtitle">通过互动式学习，像玩游戏一样掌握编程技能</p>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else class="course-grid">
      <div
        v-for="course in courses"
        :key="course.id"
        class="course-card"
        :style="{ '--course-color': course.color }"
        @click="router.push(`/course/${course.id}`)"
      >
        <div class="course-emoji">{{ course.emoji }}</div>
        <h2 class="course-title">{{ course.title }}</h2>
        <p class="course-desc">{{ course.description }}</p>
        <div class="course-lang">{{ course.language }}</div>
      </div>
    </div>

    <div class="features">
      <div class="feature">
        <span class="feature-icon">🎯</span>
        <span>技能树闯关</span>
      </div>
      <div class="feature">
        <span class="feature-icon">🔥</span>
        <span>连续打卡</span>
      </div>
      <div class="feature">
        <span class="feature-icon">🤖</span>
        <span>AI 生成习题</span>
      </div>
      <div class="feature">
        <span class="feature-icon">⚡</span>
        <span>在线代码运行</span>
      </div>
    </div>

    <div class="ai-tools">
      <h2 class="section-title">AI 学习工具</h2>
      <div class="tool-grid">
        <div class="tool-card adaptive" @click="router.push('/adaptive')">
          <span class="tool-icon">📊</span>
          <h3>自适应学习路径</h3>
          <p>AI 分析错题，识别薄弱知识点，生成针对性练习</p>
          <span class="tool-badge">Eino Chain</span>
        </div>
        <div class="tool-card tutor" @click="router.push('/tutor')">
          <span class="tool-icon">🤖</span>
          <h3>AI 编程导师</h3>
          <p>贴入代码，AI 运行、诊断、修复 bug，支持多轮对话</p>
          <span class="tool-badge">Eino Agent</span>
        </div>
        <div class="tool-card knowledge" @click="router.push('/knowledge')">
          <span class="tool-icon">📚</span>
          <h3>知识点问答</h3>
          <p>基于课程内容的 RAG 检索增强问答，个性化解释</p>
          <span class="tool-badge">Eino RAG</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-page { padding-top: 20px; padding-bottom: 40px; }
.hero { text-align: center; margin-bottom: 40px; }
.hero-title { font-size: 28px; color: var(--text); margin-bottom: 8px; }
.hero-subtitle { color: var(--text-light); font-size: 16px; }

.course-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  margin-bottom: 40px;
}

.course-card {
  background: white;
  border: 3px solid var(--course-color, var(--primary));
  border-radius: var(--radius);
  padding: 28px 20px;
  cursor: pointer;
  text-align: center;
  transition: transform 0.15s, box-shadow 0.15s;
  box-shadow: 0 4px 0 var(--course-color, var(--primary));
}

.course-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 0 var(--course-color, var(--primary));
}

.course-emoji { font-size: 48px; margin-bottom: 12px; }
.course-title { font-size: 22px; margin-bottom: 8px; color: var(--course-color, var(--primary)); }
.course-desc { color: var(--text-light); font-size: 14px; margin-bottom: 12px; }
.course-lang {
  display: inline-block;
  background: var(--bg-gray);
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  color: var(--text-light);
}

.features {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  text-align: center;
}

.feature {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 8px;
  background: var(--bg-gray);
  border-radius: var(--radius-sm);
}

.feature-icon { font-size: 28px; }
.feature span:last-child { font-size: 12px; color: var(--text-light); }

.loading { text-align: center; padding: 40px; color: var(--text-light); }

.ai-tools { margin-top: 40px; }
.section-title { font-size: 20px; font-weight: 800; margin-bottom: 16px; text-align: center; }
.tool-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; }
.tool-card { background: white; border: 2px solid var(--border); border-radius: var(--radius); padding: 24px 16px; cursor: pointer; text-align: center; transition: all 0.2s; position: relative; }
.tool-card:hover { transform: translateY(-4px); box-shadow: 0 8px 20px rgba(0,0,0,0.1); }
.tool-card.adaptive { border-color: #8b5cf6; }
.tool-card.tutor { border-color: #3b82f6; }
.tool-card.knowledge { border-color: #10b981; }
.tool-icon { font-size: 40px; display: block; margin-bottom: 8px; }
.tool-card h3 { font-size: 16px; font-weight: 800; margin-bottom: 6px; }
.tool-card p { font-size: 13px; color: var(--text-light); line-height: 1.5; }
.tool-badge { display: inline-block; margin-top: 8px; padding: 2px 10px; border-radius: 10px; font-size: 11px; font-weight: 600; background: var(--bg-gray); color: var(--text-light); }
</style>
