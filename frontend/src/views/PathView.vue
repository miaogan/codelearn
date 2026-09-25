<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { courseApi, examApi } from '@/api/client'
import type { CertStatus, LearningPath, SkillTreeLesson } from '@/types'

const route = useRoute()
const router = useRouter()
const path = ref<LearningPath | null>(null)
const certStatus = ref<CertStatus | null>(null)
const loading = ref(true)

onMounted(async () => {
  const id = Number(route.params.id)
  try {
    const res = await courseApi.path(id)
    path.value = res.data
  } catch (e) {
    // ignore
  }
  try {
    const res = await examApi.certStatus(id)
    certStatus.value = res.data
  } catch (e) {
    // ignore
  }
  loading.value = false
})

function enterLesson(lesson: SkillTreeLesson) {
  if (lesson.unlocked) {
    router.push(`/lesson/${lesson.id}`)
  }
}

function unitCompleted(unit: { lessons: SkillTreeLesson[] }): boolean {
  return unit.lessons.length > 0 && unit.lessons.every((l) => l.completed)
}

function enterUnitExam(unit: { id: number; lessons: SkillTreeLesson[] }) {
  if (unitCompleted(unit)) {
    router.push(`/unit-exam/${unit.id}`)
  }
}

function enterCertExam() {
  if (certStatus.value?.eligible) {
    router.push(`/cert-exam/${route.params.id}`)
  }
}

function offsetClass(index: number) {
  const offsets = ['center', 'left', 'right']
  return offsets[index % 3]
}

function certLevelText(level?: string) {
  const map: Record<string, string> = { beginner: '入门', intermediate: '进阶', advanced: '熟练' }
  return (level && map[level]) || level || ''
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

          <!-- 单元考试入口：完成全部课时后解锁 -->
          <div class="connector"></div>
          <div
            class="lesson-node unit-exam-node"
            :class="{ completed: unitCompleted(unit), locked: !unitCompleted(unit) }"
            @click="enterUnitExam(unit)"
          >
            <div class="node-circle exam-circle" :style="{ '--node-color': unit.color }">
              <span v-if="unitCompleted(unit)" class="icon">📝</span>
              <span v-else class="lock">🔒</span>
            </div>
            <div class="node-label">单元考试</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 课程认证考试节点：全部单元考试通过后解锁 -->
    <div class="cert-block" v-if="certStatus">
      <div
        class="cert-node"
        :class="{ eligible: certStatus.eligible, certified: certStatus.certified, locked: !certStatus.eligible }"
        @click="enterCertExam"
      >
        <div class="cert-circle">
          <span v-if="certStatus.certified" class="cert-icon">🎖️</span>
          <span v-else-if="certStatus.eligible" class="cert-icon">🏆</span>
          <span v-else class="lock">🔒</span>
        </div>
        <div class="cert-info">
          <div class="cert-title">课程认证考试</div>
          <div class="cert-desc">
            <template v-if="certStatus.certified">已获得认证 · 等级：{{ certLevelText(certStatus.certificate?.level) }}</template>
            <template v-else-if="certStatus.eligible">全部单元考试已通过，可以挑战！</template>
            <template v-else>通过全部 {{ certStatus.units_total }} 个单元考试后解锁（已通过 {{ certStatus.units_passed }}）</template>
          </div>
        </div>
      </div>
    </div>

    <div class="action-bar">
      <button class="btn-exam" @click="router.push(`/course/${route.params.id}/exam`)">
        <span class="btn-icon">🤖</span>
        <span>AI 模拟考试</span>
      </button>
      <button class="btn-predict" @click="router.push(`/course/${route.params.id}/prediction`)">
        <span class="btn-icon">🎯</span>
        <span>考试预测</span>
      </button>
      <button class="btn-skill" @click="router.push(`/course/${route.params.id}/skill`)">
        <span class="btn-icon">🧠</span>
        <span>能力图谱</span>
      </button>
      <button class="btn-wrong" @click="router.push('/wrong-exercises')">
        <span class="btn-icon">📝</span>
        <span>错题本</span>
      </button>
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

.unit-exam-node.completed .exam-circle {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border: 4px solid #4f46e5;
  box-shadow: 0 4px 0 #4f46e5;
  animation: exam-pulse 2s infinite;
}
@keyframes exam-pulse {
  0%, 100% { box-shadow: 0 4px 0 #4f46e5; }
  50% { box-shadow: 0 4px 0 #4f46e5, 0 0 0 6px rgba(99, 102, 241, 0.2); }
}

.back-btn { margin-top: 32px; width: 100%; }

/* 课程认证考试节点 */
.cert-block {
  display: flex;
  justify-content: center;
  margin-top: 28px;
}
.cert-node {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
  max-width: 480px;
  padding: 18px 20px;
  background: white;
  border: 3px solid var(--border);
  border-radius: var(--radius);
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 0 #e5e5e5;
}
.cert-node.eligible:not(.certified) {
  border-color: #f59e0b;
  box-shadow: 0 4px 0 #d97706;
  background: linear-gradient(135deg, #fffbeb, #fef3c7);
}
.cert-node.certified {
  border-color: #f59e0b;
  box-shadow: 0 4px 0 #d97706;
  background: linear-gradient(135deg, #fffbeb, #fde68a);
}
.cert-node.locked {
  pointer-events: none;
  opacity: 0.75;
  background: var(--bg-gray);
}
.cert-node.eligible:hover { transform: translateY(-2px); }
.cert-circle {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: white;
  box-shadow: 0 4px 0 #b45309;
  flex-shrink: 0;
}
.cert-node.locked .cert-circle {
  background: #e5e5e5;
  color: #bbb;
  box-shadow: 0 4px 0 #d4d4d4;
}
.cert-icon { font-size: 30px; }
.cert-info { text-align: left; }
.cert-title { font-size: 17px; font-weight: 900; }
.cert-desc { font-size: 13px; color: var(--text-light); margin-top: 2px; }
.cert-node.eligible .cert-desc { color: #92400e; }

.action-bar {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.btn-exam, .btn-skill, .btn-wrong, .btn-predict {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 16px 20px;
  border-radius: var(--radius-sm);
  border: none;
  font-size: 16px;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-exam {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  box-shadow: 0 4px 0 #4f46e5;
}

.btn-exam:hover { transform: translateY(-2px); box-shadow: 0 6px 0 #4f46e5; }
.btn-exam:active { transform: translateY(2px); box-shadow: 0 2px 0 #4f46e5; }

.btn-skill {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  box-shadow: 0 4px 0 #047857;
}

.btn-skill:hover { transform: translateY(-2px); box-shadow: 0 6px 0 #047857; }
.btn-skill:active { transform: translateY(2px); box-shadow: 0 2px 0 #047857; }

.btn-predict {
  background: linear-gradient(135deg, #f59e0b, #ea580c);
  color: white;
  box-shadow: 0 4px 0 #c2410c;
}

.btn-predict:hover { transform: translateY(-2px); box-shadow: 0 6px 0 #c2410c; }
.btn-predict:active { transform: translateY(2px); box-shadow: 0 2px 0 #c2410c; }

.btn-wrong {
  background: white;
  color: var(--text);
  border: 2px solid var(--border);
  box-shadow: 0 4px 0 #e5e5e5;
}

.btn-wrong:hover { transform: translateY(-2px); box-shadow: 0 6px 0 #e5e5e5; }
.btn-wrong:active { transform: translateY(2px); box-shadow: 0 2px 0 #e5e5e5; }

.btn-icon { font-size: 20px; }
.loading { text-align: center; padding: 40px; color: var(--text-light); }

@media (max-width: 640px) {
  .action-bar { flex-wrap: wrap; }
  .action-bar button { min-width: calc(50% - 6px); flex: 1 1 auto; }
}
</style>
