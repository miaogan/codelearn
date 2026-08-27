<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { courseApi } from '@/api/client'
import type { LearningPath, SkillTreeLesson } from '@/types'

const route = useRoute()
const router = useRouter()
const path = ref<LearningPath | null>(null)
const loading = ref(true)

onMounted(async () => {
  const id = Number(route.params.id)
  try {
    const res = await courseApi.path(id)
    path.value = res.data
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
})

function enterLesson(lesson: SkillTreeLesson) {
  if (lesson.unlocked) {
    router.push(`/lesson/${lesson.id}`)
  }
}

function offsetClass(index: number) {
  const offsets = ['center', 'left', 'right']
  return offsets[index % 3]
}
</script>

<template>
  <div class="container path-page" v-if="!loading && path">
    <div class="path-header" :style="{ background: path.course.color }">
      <span class="course-emoji">{{ path.course.emoji }}</span>
      <h1 class="course-name">{{ path.course.title }}</h1>
      <p class="course-desc">{{ path.course.description }}</p>
    </div>

    <div class="units">
      <div v-for="(unit, ui) in path.units" :key="unit.id" class="unit-block">
        <div class="unit-header" :style="{ borderColor: unit.color }">
          <span class="unit-icon">{{ unit.icon }}</span>
          <div class="unit-info">
            <div class="unit-title">{{ unit.title }}</div>
            <div class="unit-desc">{{ unit.description }}</div>
          </div>
        </div>

        <div class="lessons-path">
          <template v-for="(lesson, li) in unit.lessons" :key="lesson.id">
            <div class="connector" v-if="li > 0"></div>
            <div
              class="lesson-node"
              :class="[offsetClass(li), { completed: lesson.completed, locked: !lesson.unlocked, current: lesson.unlocked && !lesson.completed }]"
              @click="enterLesson(lesson)"
            >
              <div class="node-circle" :style="{ '--node-color': unit.color }">
                <span v-if="lesson.completed" class="check">✓</span>
                <span v-else-if="lesson.unlocked" class="icon">{{ lesson.icon }}</span>
                <span v-else class="lock">🔒</span>
              </div>
              <div class="node-label">{{ lesson.title }}</div>
            </div>
          </template>
        </div>
      </div>
    </div>

    <button class="btn-ghost back-btn" @click="router.push('/')">← 返回课程列表</button>
  </div>

  <div v-else class="container loading">加载中...</div>
</template>

<style scoped>
.path-page { padding-bottom: 40px; }

.path-header {
  border-radius: var(--radius);
  padding: 28px 20px;
  text-align: center;
  color: white;
  margin-bottom: 32px;
}

.course-emoji { font-size: 48px; display: block; margin-bottom: 8px; }
.course-name { font-size: 24px; margin-bottom: 4px; }
.course-desc { font-size: 14px; opacity: 0.9; }

.unit-block { margin-bottom: 24px; }

.unit-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: white;
  border: 3px solid;
  border-radius: var(--radius-sm);
  margin-bottom: 16px;
  box-shadow: 0 2px 0 #e5e5e5;
}

.unit-icon { font-size: 32px; }
.unit-title { font-size: 18px; font-weight: 800; }
.unit-desc { font-size: 13px; color: var(--text-light); }

.lessons-path {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
}

.connector {
  width: 4px;
  height: 32px;
  background: repeating-linear-gradient(to bottom, var(--border) 0, var(--border) 6px, transparent 6px, transparent 12px);
}

.lesson-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  transition: transform 0.15s;
}

.lesson-node.left { align-self: flex-start; margin-left: 15%; }
.lesson-node.right { align-self: flex-end; margin-right: 15%; }

.lesson-node:hover:not(.locked) {
  transform: scale(1.05);
}

.node-circle {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  position: relative;
  box-shadow: 0 4px 0 var(--node-color, var(--primary-dark));
  transition: all 0.15s;
}

.lesson-node.completed .node-circle {
  background: var(--primary);
  color: white;
  border: 4px solid var(--primary-dark);
}

.lesson-node.current .node-circle {
  background: var(--node-color, var(--primary));
  color: white;
  border: 4px solid;
  border-color: color-mix(in srgb, var(--node-color, var(--primary)) 70%, black);
}

.lesson-node.locked .node-circle {
  background: var(--bg-gray);
  color: #bbb;
  border: 4px solid #e5e5e5;
  box-shadow: 0 4px 0 #e5e5e5;
  cursor: not-allowed;
}

.lesson-node.locked { pointer-events: none; }

.check { font-size: 32px; font-weight: 900; }
.lock { font-size: 24px; }

.node-label {
  margin-top: 8px;
  font-size: 13px;
  font-weight: 700;
  color: var(--text-light);
  text-align: center;
}

.back-btn { margin-top: 32px; width: 100%; }
.loading { text-align: center; padding: 40px; color: var(--text-light); }
</style>
