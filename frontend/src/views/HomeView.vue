<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { courseApi, userApi, wrongApi, skillApi, srsApi } from '@/api/client'
import type { Course, SkillTreeLesson, WrongExerciseItem, ReviewRecommendation } from '@/types'

const router = useRouter()
const courses = ref<Course[]>([])
const loading = ref(true)
const todayXP = ref(0)
const dailyGoal = ref(50)
const streakDays = ref(0)
const srsDueCount = ref(0)
const nextLesson = ref<{ lessonId: number; courseId: number; courseTitle: string; lessonTitle: string } | null>(null)
const recommendWrong = ref<WrongExerciseItem[]>([])
const recommendReview = ref<ReviewRecommendation[]>([])

onMounted(async () => {
  try {
    const [coursesRes, statsRes, wrongRes, reviewRes, srsRes] = await Promise.all([
      courseApi.list(),
      userApi.stats(),
      wrongApi.list(true),
      skillApi.recommendations(),
      srsApi.reviews(),
    ])
    courses.value = coursesRes.data.courses
    todayXP.value = statsRes.data.today_xp
    dailyGoal.value = statsRes.data.daily_goal
    streakDays.value = statsRes.data.streak_days
    recommendWrong.value = wrongRes.data.wrong_exercises.slice(0, 3)
    recommendReview.value = reviewRes.data.recommendations.slice(0, 3)
    srsDueCount.value = srsRes.data.due_count || 0
    await findNextLesson()
  } catch (e) {
    // ignore
  } finally {
    loading.value = false
  }
})

async function findNextLesson() {
  for (const course of courses.value) {
    try {
      const res = await courseApi.path(course.id)
      for (const unit of res.data.units) {
        for (const lesson of unit.lessons) {
          if (lesson.unlocked && !lesson.completed) {
            nextLesson.value = {
              lessonId: lesson.id,
              courseId: course.id,
              courseTitle: course.title,
              lessonTitle: lesson.title,
            }
            return
          }
        }
      }
    } catch (e) {
      // ignore
    }
  }
}

const goalPercent = computed(() => {
  if (dailyGoal.value <= 0) return 0
  return Math.min(100, Math.round((todayXP.value / dailyGoal.value) * 100))
})
const goalDone = computed(() => goalPercent.value >= 100)

async function setGoal(goal: number) {
  try {
    const res = await userApi.updateDailyGoal(goal)
    dailyGoal.value = res.data.daily_goal
  } catch (e) {
    // ignore
  }
}

const GOAL_OPTIONS = [20, 50, 100]

const R = 52
const CIRC = 2 * Math.PI * R
</script>

<template>
  <div class="container home-page">
    <!-- 今日学习卡片 -->
    <div class="daily-card" :class="{ done: goalDone }">
      <div class="daily-ring-wrap">
        <svg class="daily-ring" viewBox="0 0 120 120">
          <circle class="ring-bg" cx="60" cy="60" :r="R" />
          <circle
            class="ring-fill"
            cx="60"
            cy="60"
            :r="R"
            :stroke-dasharray="CIRC"
            :stroke-dashoffset="CIRC * (1 - goalPercent / 100)"
          />
        </svg>
        <div class="ring-center">
          <span class="ring-xp">{{ todayXP }}</span>
          <span class="ring-label">/ {{ dailyGoal }} XP</span>
        </div>
      </div>

      <div class="daily-info">
        <div class="daily-head">
          <h2 class="daily-title">{{ goalDone ? '今日目标已达成 🎉' : '今日学习' }}</h2>
          <span class="streak-badge" title="连续打卡天数">🔥 {{ streakDays }} 天</span>
        </div>

        <p class="daily-desc">
          {{ goalDone ? '太棒了！明天继续保持，连续打卡别断哦。' : `今日进度 ${goalPercent}%，完成每日目标即可打卡。` }}
        </p>

        <div class="goal-options">
          <span class="goal-label">每日目标：</span>
          <button
            v-for="g in GOAL_OPTIONS"
            :key="g"
            class="goal-option"
            :class="{ active: dailyGoal === g }"
            @click="setGoal(g)"
          >{{ g }} XP</button>
        </div>

        <div class="daily-actions">
          <button
            v-if="nextLesson"
            class="btn-primary"
            @click="router.push(`/lesson/${nextLesson!.lessonId}`)"
          >
            继续学习：{{ nextLesson!.lessonTitle }}
          </button>
          <button v-else class="btn-primary" @click="router.push('/course/1')">
            开始学习
          </button>
          <button class="btn-secondary" @click="router.push('/leaderboard')">
            🏆 排行榜
          </button>
        </div>
      </div>
    </div>

    <!-- 今日快捷入口 -->
    <div class="quick-strip">
      <div class="quick-item srs" @click="router.push('/srs-reviews')">
        <span class="quick-icon">🧠</span>
        <div class="quick-body">
          <div class="quick-title">今日复习</div>
          <div class="quick-meta" :class="{ urgent: srsDueCount > 0 }">
            {{ srsDueCount > 0 ? srsDueCount + ' 题待复习' : '今日无待复习' }}
          </div>
        </div>
        <span class="quick-arrow">→</span>
      </div>
      <div class="quick-item report" @click="router.push('/weekly-report')">
        <span class="quick-icon">📊</span>
        <div class="quick-body">
          <div class="quick-title">学习周报</div>
          <div class="quick-meta">看看本周的学习成果</div>
        </div>
        <span class="quick-arrow">→</span>
      </div>
      <div class="quick-item badge" @click="router.push('/achievements')">
        <span class="quick-icon">🏆</span>
        <div class="quick-body">
          <div class="quick-title">成就墙</div>
          <div class="quick-meta">收集徽章，提升段位</div>
        </div>
        <span class="quick-arrow">→</span>
      </div>
      <div class="quick-item project" @click="router.push('/projects')">
        <span class="quick-icon">🛠️</span>
        <div class="quick-body">
          <div class="quick-title">项目工坊</div>
          <div class="quick-meta">动手实战，创作作品</div>
        </div>
        <span class="quick-arrow">→</span>
      </div>
      <div class="quick-item plan" @click="router.push('/study-plan')">
        <span class="quick-icon">🤖</span>
        <div class="quick-body">
          <div class="quick-title">AI 学习计划</div>
          <div class="quick-meta">定制你的学习路线</div>
        </div>
        <span class="quick-arrow">→</span>
      </div>
    </div>

    <!-- 推荐练习 -->
    <div class="recommend" v-if="recommendWrong.length > 0">
      <h2 class="section-title">今日推荐练习（错题巩固）</h2>
      <div class="recommend-list">
        <div
          v-for="w in recommendWrong"
          :key="w.exercise_id"
          class="recommend-item"
          @click="router.push('/wrong-exercises')"
        >
          <span class="rec-icon">📝</span>
          <div class="rec-body">
            <div class="rec-question">{{ w.question }}</div>
            <div class="rec-meta">错了 {{ w.wrong_count }} 次 · {{ w.difficulty }}</div>
          </div>
          <span class="rec-arrow">→</span>
        </div>
      </div>
    </div>

    <!-- 薄弱章节复习推荐（自适应联动） -->
    <div class="recommend" v-if="recommendReview.length > 0">
      <h2 class="section-title">🧠 薄弱章节复习推荐</h2>
      <div class="recommend-list">
        <div
          v-for="r in recommendReview"
          :key="r.lesson_id"
          class="recommend-item review"
          @click="router.push(`/lesson/${r.lesson_id}`)"
        >
          <span class="rec-icon">📚</span>
          <div class="rec-body">
            <div class="rec-question">{{ r.lesson_title }}</div>
            <div class="rec-meta">{{ r.course_title }} · {{ r.reason }}</div>
          </div>
          <span class="rec-arrow">→</span>
        </div>
      </div>
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

/* 今日学习卡片 */
.daily-card {
  display: flex;
  gap: 24px;
  align-items: center;
  background: white;
  border: 3px solid var(--primary);
  border-radius: var(--radius);
  padding: 24px;
  margin-bottom: 28px;
  box-shadow: 0 4px 0 var(--primary-dark);
}

.daily-card.done {
  border-color: var(--secondary);
  box-shadow: 0 4px 0 var(--secondary-dark, #d97706);
}

.daily-ring-wrap { position: relative; flex-shrink: 0; }
.daily-ring { width: 120px; height: 120px; transform: rotate(-90deg); }
.ring-bg { fill: none; stroke: var(--bg-gray); stroke-width: 10; }
.ring-fill {
  fill: none;
  stroke: var(--primary);
  stroke-width: 10;
  stroke-linecap: round;
  transition: stroke-dashoffset 0.6s;
}
.daily-card.done .ring-fill { stroke: var(--secondary); }
.ring-center {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.ring-xp { font-size: 26px; font-weight: 900; color: var(--text); }
.ring-label { font-size: 11px; color: var(--text-light); }

.daily-info { flex: 1; min-width: 0; }
.daily-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 8px; }
.daily-title { font-size: 20px; font-weight: 900; }
.streak-badge {
  background: #fff3e0;
  color: var(--warning);
  font-weight: 800;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 14px;
  white-space: nowrap;
}
.daily-desc { font-size: 14px; color: var(--text-light); margin-bottom: 12px; }

.goal-options { display: flex; align-items: center; gap: 8px; margin-bottom: 16px; flex-wrap: wrap; }
.goal-label { font-size: 13px; color: var(--text-light); }
.goal-option {
  border: 2px solid var(--border);
  background: white;
  border-radius: 16px;
  padding: 3px 12px;
  font-size: 12px;
  font-weight: 700;
  color: var(--text-light);
  cursor: pointer;
  transition: all 0.15s;
}
.goal-option.active {
  border-color: var(--primary);
  color: var(--primary);
  background: rgba(74, 137, 255, 0.08);
}

.daily-actions { display: flex; gap: 12px; flex-wrap: wrap; }
.btn-primary, .btn-secondary {
  padding: 12px 20px;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-primary { background: var(--primary); color: white; box-shadow: 0 4px 0 var(--primary-dark); }
.btn-primary:hover { transform: translateY(-2px); box-shadow: 0 6px 0 var(--primary-dark); }
.btn-primary:active { transform: translateY(2px); box-shadow: 0 2px 0 var(--primary-dark); }
.btn-secondary {
  background: white;
  color: var(--text);
  border: 2px solid var(--border);
  box-shadow: 0 4px 0 #e5e5e5;
}
.btn-secondary:hover { transform: translateY(-2px); box-shadow: 0 6px 0 #e5e5e5; }
.btn-secondary:active { transform: translateY(2px); box-shadow: 0 2px 0 #e5e5e5; }

/* 推荐练习 */
.recommend { margin-bottom: 28px; }
.section-title { font-size: 18px; font-weight: 800; margin-bottom: 12px; }

/* 今日快捷入口 */
.quick-strip {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  margin-bottom: 28px;
}
.quick-item {
  display: flex;
  align-items: center;
  gap: 10px;
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 14px 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.quick-item:hover { border-color: var(--primary); transform: translateY(-2px); }
.quick-item.srs { border-color: #a78bfa; }
.quick-item.srs:hover { border-color: #7c3aed; }
.quick-item.report { border-color: #6ee7b7; }
.quick-item.report:hover { border-color: #10b981; }
.quick-item.badge { border-color: #fcd34d; }
.quick-item.badge:hover { border-color: #f59e0b; }
.quick-item.project { border-color: #fb923c; }
.quick-item.project:hover { border-color: #ea580c; }
.quick-item.plan { border-color: #818cf8; }
.quick-item.plan:hover { border-color: #4f46e5; }
.quick-icon { font-size: 26px; flex-shrink: 0; }
.quick-body { flex: 1; min-width: 0; }
.quick-title { font-size: 14px; font-weight: 800; }
.quick-meta { font-size: 11px; color: var(--text-light); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.quick-meta.urgent { color: var(--danger); font-weight: 800; }
.quick-arrow { color: var(--text-light); }
.recommend-list { display: flex; flex-direction: column; gap: 10px; }
.recommend-item {
  display: flex;
  align-items: center;
  gap: 12px;
  background: white;
  border: 2px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.15s;
}
.recommend-item:hover { border-color: var(--primary); transform: translateX(4px); }
.recommend-item.review { border-color: #6ee7b7; }
.recommend-item.review:hover { border-color: #10b981; }
.rec-icon { font-size: 22px; }
.rec-body { flex: 1; min-width: 0; }
.rec-question { font-size: 14px; font-weight: 700; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.rec-meta { font-size: 12px; color: var(--text-light); }
.rec-arrow { color: var(--text-light); }

/* 课程卡片（沿用原样式） */
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

.loading { text-align: center; padding: 40px; color: var(--text-light); }

.ai-tools { margin-top: 40px; }
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

@media (max-width: 640px) {
  .daily-card { flex-direction: column; text-align: center; }
  .daily-head { justify-content: center; }
  .daily-actions { justify-content: center; }
  .quick-strip { grid-template-columns: 1fr; }
}
</style>
