import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'
import type {
  User, Course, LearningPath, Lesson, Exercise,
  UserStats, RunResult, JudgeResult, SubmitResult, WrongExerciseItem, ExamResult,
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

export default api
