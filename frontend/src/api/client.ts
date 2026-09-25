import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'
import type {
  User, Course, LearningPath, Lesson, Exercise,
  UserStats, RunResult, JudgeResult, SubmitResult, WrongExerciseItem, ExamResult,
  Exam, ExamReport, LeaderboardEntry, CalendarDay, Certificate, CertStatus,
  SkillMap, ReviewRecommendation, SRSReviewItem, AchievementSummary, WeeklyReport,
  Project, ProjectFile, ProjectTemplate, Portfolio, ExamPrediction, StudyPlan,
} from '@/types'

const api = axios.create({
  baseURL: '/api',
})

api.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      // 避免在登录/注册页重复跳转
      const currentRoute = router.currentRoute.value
      if (currentRoute.name !== 'login' && currentRoute.name !== 'register') {
        router.push({ name: 'login' })
      }
    }
    return Promise.reject(err)
  }
)

export const authApi = {
  register: (data: { username: string; email: string; password: string }) =>
    api.post<{ token: string; user: User }>('/auth/register', data),
  login: (data: { account: string; password: string }) =>
    api.post<{ token: string; user: User }>('/auth/login', data),
}

export const courseApi = {
  list: () => api.get<{ courses: Course[] }>('/courses'),
  path: (id: number) => api.get<LearningPath>(`/courses/${id}`),
  lesson: (id: number) => api.get<{ lesson: Lesson }>(`/lessons/${id}`),
  exercises: (id: number) => api.get<{ exercises: Exercise[] }>(`/lessons/${id}/exercises`),
}

export const exerciseApi = {
  submit: (id: number, answer: string) =>
    api.post<SubmitResult>(`/exercises/${id}/submit`, { answer }),
  generate: (lessonId: number, data: { language: string; topic: string; count?: number; type?: string; difficulty?: string }) =>
    api.post<{ exercises: Exercise[] }>(`/lessons/${lessonId}/generate`, data),
  hint: (data: { question: string; user_answer: string; language: string }) =>
    api.post<{ hint: string }>('/exercises/hint', data),
  template: (id: number) =>
    api.get<{ code_template: string; type: string; question: string; difficulty: string }>(`/exercises/${id}/template`),
  examSubmit: (answers: { exercise_id: number; answer: string }[]) =>
    api.post<ExamResult>('/exercises/exam-submit', { answers }),
}

export const codeApi = {
  run: (language: string, code: string) =>
    api.post<RunResult>('/code/run', { language, code }),
  judge: (exerciseId: number, language: string, code: string) =>
    api.post<JudgeResult>('/code/judge', { exercise_id: exerciseId, language, code }),
  complete: (lessonId: number) =>
    api.post<{ xp_earned: number; hearts: number; message: string }>(`/lessons/${lessonId}/complete`),
}

export const userApi = {
  stats: () => api.get<UserStats>('/users/me/stats'),
  progress: () => api.get<{ progress: any[] }>('/users/me/progress'),
  updateDailyGoal: (goal: number) =>
    api.put<{ daily_goal: number }>('/users/me/daily-goal', { goal }),
}

export const wrongApi = {
  list: (unmastered = false) =>
    api.get<{ wrong_exercises: WrongExerciseItem[]; total: number }>('/wrong-exercises', {
      params: unmastered ? { unmastered: 1 } : {},
    }),
  master: (exerciseId: number) =>
    api.post<{ message: string }>(`/wrong-exercises/${exerciseId}/master`),
  count: () => api.get<{ count: number }>('/wrong-exercises/count'),
}

export const adaptiveApi = {
  recommend: (courseId?: number, language?: string) =>
    api.get('/adaptive/recommend', {
      params: { course_id: courseId, language },
    }),
}

// ===== 单元考试（Sprint 0 新增） =====

export const examApi = {
  startUnit: (unitId: number) =>
    api.post<Exam>(`/units/${unitId}/exam`),
  startCert: (courseId: number) =>
    api.post<Exam>(`/courses/${courseId}/cert-exam`),
  certStatus: (courseId: number) =>
    api.get<CertStatus>(`/courses/${courseId}/cert-status`),
  submit: (examId: number, answers: { question_id: number; exercise_id: number; answer: string }[], durationSec = 0, tabSwitches = 0) =>
    api.post<ExamReport>(`/exams/${examId}/submit`, { answers, duration_sec: durationSec, tab_switches: tabSwitches }),
  report: (examId: number) =>
    api.get<ExamReport>(`/exams/${examId}/report`),
}

// ===== 能力图谱 / 复习推荐（Sprint 3 新增） =====

export const skillApi = {
  map: (courseId: number) =>
    api.get<SkillMap>(`/courses/${courseId}/skill-map`),
  recommendations: () =>
    api.get<{ recommendations: ReviewRecommendation[] }>('/users/me/review-recommendations'),
}

// ===== 排行榜 / 学习日历（Sprint 0 新增） =====

export const leaderboardApi = {
  weekly: () =>
    api.get<{ entries: LeaderboardEntry[]; me: LeaderboardEntry }>('/leaderboard'),
  calendar: (month?: string) =>
    api.get<{ year: number; month: number; days: CalendarDay[] }>('/users/me/calendar', {
      params: month ? { month } : {},
    }),
}

// ===== 能力认证证书（Sprint 0 新增） =====

export const certificateApi = {
  list: () =>
    api.get<{ certificates: Certificate[] }>('/certificates'),
  verify: (certNo: string) =>
    api.post<{ valid: boolean; certificate?: Certificate; message?: string }>('/certificates/verify', { cert_no: certNo }),
}

export const tutorApi = {
  debug: (data: { language: string; code: string; question: string }) =>
    api.post('/tutor/debug', data),
  chat: (data: { messages: { role: string; content: string }[]; code: string; language: string }) =>
    api.post<{ reply: string }>('/tutor/chat', data),
  review: (data: { code: string; language: string }) =>
    api.post<{ review: string }>('/tutor/review', data),
  run: (data: { language: string; code: string; input?: string }) =>
    api.post<RunResult>('/tutor/run', data),
}

export const knowledgeApi = {
  ask: (data: { question: string; language: string }) =>
    api.post('/knowledge/ask', data),
}

export const feedbackApi = {
  submit: (data: { category: string; content: string; contact?: string }) =>
    api.post<{ message: string }>('/feedback', data),
}

// ===== Sprint 6：SRS 复习 / 成就徽章 / 学习周报 =====

export const srsApi = {
  reviews: () => api.get<{ reviews: SRSReviewItem[]; due_count: number }>('/users/me/srs/reviews'),
  submitReview: (wrongId: number, correct: boolean) =>
    api.post<{ message: string }>(`/wrong-exercises/${wrongId}/srs-review`, { correct }),
}

export const achievementApi = {
  summary: () => api.get<AchievementSummary>('/users/me/achievements'),
}

export const weeklyReportApi = {
  get: () => api.get<WeeklyReport>('/users/me/weekly-report'),
}

// ===== Sprint 6 第二批：项目工坊 / 能力档案 / 分数预测 / AI 学习计划 =====

export const projectApi = {
  templates: () => api.get<{ templates: ProjectTemplate[] }>('/projects/templates'),
  list: () => api.get<{ projects: Project[] }>('/projects'),
  get: (id: number) => api.get<{ project: Project; files: ProjectFile[] }>(`/projects/${id}`),
  create: (data: { course_id?: number; title?: string; description?: string; language?: string; main_file?: string }) =>
    api.post<Project>('/projects', data),
  saveFiles: (id: number, mainFile: string, files: ProjectFile[]) =>
    api.put<{ message: string }>(`/projects/${id}/files`, { main_file: mainFile, files }),
  run: (id: number) => api.post<RunResult>(`/projects/${id}/run`),
  complete: (id: number) => api.post<{ project: Project; message: string }>(`/projects/${id}/complete`),
  remove: (id: number) => api.delete<{ message: string }>(`/projects/${id}`),
}

export const portfolioApi = {
  get: (username: string) => api.get<Portfolio>(`/me/${username}`),
}

export const predictionApi = {
  get: (courseId: number) => api.get<ExamPrediction>(`/courses/${courseId}/prediction`),
}

export const studyPlanApi = {
  generate: (data: { goal?: string; language?: string; weeks?: number }) =>
    api.post<StudyPlan>('/study-plan/generate', data),
}

export default api
