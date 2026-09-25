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
  freeze_cards: number
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
  source?: string
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

export interface TypeAccuracy {
  type: string
  total: number
  correct: number
  accuracy: number
}

export interface KnowledgePoint {
  lesson_id: number
  title: string
  total: number
  correct: number
  mastery: number
}

export interface ExamReport {
  exam_id: number
  exam_type: string
  course_id: number
  score: number
  correct_count: number
  total_count: number
  passed: boolean
  pass_score: number
  duration_sec: number
  tab_switches: number
  attempts: number
  best_score: number
  by_type: TypeAccuracy[]
  knowledge: KnowledgePoint[]
  results: ExamReportItem[]
  certificate?: Certificate
}

export interface CertStatus {
  eligible: boolean
  certified: boolean
  certificate?: Certificate
  units_total: number
  units_passed: number
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

// ===== 能力图谱 / 复习推荐（Sprint 3 新增） =====

export interface SkillNode {
  lesson_id: number
  lesson_title: string
  icon: string
  unit_id: number
  unit_title: string
  mastery: number
  status: string
  completed: boolean
}

export interface SkillMap {
  course: Course
  nodes: SkillNode[]
  overall: number
}

export interface ReviewRecommendation {
  lesson_id: number
  lesson_title: string
  course_id: number
  course_title: string
  unit_id: number
  weak_count: number
  mastery: number
  reason: string
}

// ===== Sprint 6：SRS 复习 / 成就徽章 / 学习周报 =====

export interface SRSReviewItem {
  id: number
  exercise_id: number
  type: string
  question: string
  options: string
  correct_answer: string
  explanation: string
  review_stage: number
}

export interface AchievementBadge {
  code: string
  title: string
  icon: string
  description: string
  unlocked_at: string
}

export interface AchievementTitle {
  name: string
  icon: string
  level: number
  current_xp: number
  xp_to_next: number
  next_name?: string
}

export interface AchievementSummary {
  badges: AchievementBadge[]
  total: number
  title: AchievementTitle
}

export interface WeakTopic {
  lesson_id: number
  title: string
  wrong_count: number
}

export interface WeeklyReport {
  week_start: string
  week_end: string
  xp_total: number
  active_days: number
  lessons_completed: number
  exercises_done: number
  exams_passed: number
  wrong_added: number
  streak_days: number
  weak_topics: WeakTopic[]
  summary: string
  advice: string
}

// ===== Sprint 6 第二批：项目工坊 / 能力档案 / 分数预测 / AI 学习计划 =====

export interface Project {
  id: number
  user_id: number
  course_id: number
  course_title: string
  title: string
  description: string
  language: string
  main_file: string
  status: string
  completed_at?: string
  run_count: number
  last_run_output: string
  last_run_error: string
  created_at: string
  updated_at: string
}

export interface ProjectFile {
  id: number
  project_id: number
  name: string
  content: string
  order: number
}

export interface ProjectTemplate {
  language: string
  title: string
  files: { name: string; content: string }[]
}

export interface PortfolioCourse {
  course_id: number
  course_title: string
  language: string
  emoji: string
  units_total: number
  units_passed: number
  mastery: number
  certified: boolean
  cert_level?: string
  cert_score?: number
  cert_no?: string
  project_count: number
}

export interface Portfolio {
  username: string
  xp: number
  streak_days: number
  title_name: string
  title_icon: string
  title_level: number
  badge_count: number
  courses: PortfolioCourse[]
  projects: Project[]
}

export interface PredictionFactor {
  label: string
  score: number
  weight: number
  detail: string
}

export interface PrepTask {
  title: string
  type: string
  target_id?: number
  reason: string
}

export interface ExamPrediction {
  course_id: number
  course_title: string
  probability: number
  level: string
  eligible: boolean
  units_passed: number
  units_total: number
  factors: PredictionFactor[]
  advice: string
  prep_tasks: PrepTask[]
}

export interface PlanWeek {
  week: number
  focus: string
  tasks: string[]
}

export interface StudyPlan {
  goal: string
  language: string
  total_weeks: number
  weeks: PlanWeek[]
  summary: string
}
