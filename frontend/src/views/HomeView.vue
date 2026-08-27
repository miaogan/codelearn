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
</style>
