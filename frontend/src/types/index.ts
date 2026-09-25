export interface User {
  id: number
  username: string
  email: string
  xp: number
}

export interface Course {
  id: number
  language: string
  title: string
  description: string
  emoji: string
  color: string
  order: number
}

export interface SkillTreeLesson {
  id: number
  title: string
  icon: string
  order: number
  unlocked: boolean
  completed: boolean
  score: number
}

export interface SkillTreeUnit {
  id: number
  title: string
  description: string
  icon: string
  color: string
  order: number
  lessons: SkillTreeLesson[]
}

export interface LearningPath {
  course: Course
  units: SkillTreeUnit[]
}

export interface Lesson {
  id: number
  unit_id: number
  title: string
  description: string
  content: string
  icon: string
  order: number
}

export interface Exercise {
  id: number
  lesson_id: number
  type: 'choice' | 'fillblank' | 'code' | 'order' | 'subjective'
  question: string
  options: string
  answer: string
  explanation: string
  difficulty: string
  code_template: string
  test_cases: string
  order: number
  is_ai_gen: boolean
}

export interface ExamResultItem {
  exercise_id: number
  correct: boolean
  user_answer: string
  correct_answer: string
  explanation: string
  feedback?: string
}

export interface ExamResult {
  results: ExamResultItem[]
  correct_count: number
  total_count: number
  score: number
}

export interface UserStats {
  xp: number
  streak_days: number
  hearts: number
  max_hearts: number
  daily_goal: number
  today_xp: number
  completed_today: number
}

export interface RunResult {
  output: string
  error?: string
  exit_code: number
  time_ms: number
}

export interface TestCaseResult {
  input: string
  expected: string
  actual: string
  pass: boolean
  error?: string
}

export interface JudgeResult {
  results: TestCaseResult[]
  all_pass: boolean
  pass_count: number
  total_count: number
}

export interface SubmitResult {
  correct: boolean
  explanation: string
  hearts: number
}

export interface WrongExerciseItem {
  id: number
  exercise_id: number
  type: 'choice' | 'fillblank' | 'code' | 'order'
  question: string
  options: string
  difficulty: string
  code_template: string
  user_answer: string
  correct_answer: string
  explanation: string
  wrong_count: number
  mastered: boolean
  last_wrong_at: string
  reviewed_at?: string
}

// ===== 单元考试（Sprint 0 新增） =====

export interface ExamQuestion {
  id: number
  exercise_id: number
  order: number
  type: string
  question: string
  options: string
  code_template: string
  difficulty: string
}

export interface Exam {
  id: number
  title: string
  exam_type: string
  duration_min: number
  pass_score: number
  questions: ExamQuestion[]
}

export interface ExamReportItem {
  question_id: number
  exercise_id: number
  correct: boolean
  user_answer: string
  correct_answer: string
  explanation: string
}

export interface ExamReport {
  exam_id: number
  score: number
  correct_count: number
  total_count: number
  passed: boolean
  pass_score: number
  duration_sec: number
  attempts: number
  results: ExamReportItem[]
}

// ===== 排行榜 / 学习日历（Sprint 0 新增） =====

export interface LeaderboardEntry {
  rank: number
  user_id: number
  username: string
  xp: number
}

export interface CalendarDay {
  date: string
  xp: number
}

// ===== 能力认证证书（Sprint 0 新增） =====

export interface Certificate {
  id: number
  user_id: number
  course_id: number
  cert_no: string
  level: string
  score: number
  course_title: string
  user_name: string
  issued_at: string
}
