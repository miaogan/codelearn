package model

import "time"

type User struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Username      string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email         string     `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash  string     `gorm:"size:255;not null" json:"-"`
	XP            int        `gorm:"default:0" json:"xp"`
	StreakDays    int        `gorm:"default:0" json:"streak_days"`
	LastStreakAt  *time.Time `json:"last_streak_at,omitempty"`
	Hearts        int        `gorm:"default:5" json:"hearts"`
	MaxHearts     int        `gorm:"default:5" json:"max_hearts"`
	DailyGoal     int        `gorm:"default:50" json:"daily_goal"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Course struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Language    string `gorm:"size:20;index;not null" json:"language"`
	Title       string `gorm:"size:200;not null" json:"title"`
	Description string `gorm:"size:500" json:"description"`
	Emoji       string `gorm:"size:10" json:"emoji"`
	Color       string `gorm:"size:20" json:"color"`
	Order       int    `gorm:"default:0" json:"order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Unit struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	CourseID          uint      `gorm:"index;not null" json:"course_id"`
	Title             string    `gorm:"size:200;not null" json:"title"`
	Description       string    `gorm:"size:500" json:"description"`
	Icon              string    `gorm:"size:50" json:"icon"`
	Color             string    `gorm:"size:20" json:"color"`
	Order             int       `gorm:"default:0" json:"order"`
	ExamDurationMin   int       `gorm:"default:15" json:"exam_duration_min"`
	ExamQuestionCount int       `gorm:"default:10" json:"exam_question_count"`
	CreatedAt         time.Time `json:"created_at"`
}

type Lesson struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	UnitID      uint   `gorm:"index;not null" json:"unit_id"`
	Title       string `gorm:"size:200;not null" json:"title"`
	Description string `gorm:"size:500" json:"description"`
	Content     string `gorm:"type:text" json:"content"`
	Icon        string `gorm:"size:50" json:"icon"`
	Order       int    `gorm:"default:0" json:"order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Exercise struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	LessonID     uint   `gorm:"index;not null" json:"lesson_id"`
	Type         string `gorm:"size:20;not null" json:"type"`
	Question     string `gorm:"type:text;not null" json:"question"`
	Options      string `gorm:"type:text" json:"options"`
	Answer       string `gorm:"type:text;not null" json:"answer"`
	Explanation  string `gorm:"type:text" json:"explanation"`
	Difficulty   string `gorm:"size:20;default:easy" json:"difficulty"`
	CodeTemplate string `gorm:"type:text" json:"code_template"`
	TestCases    string `gorm:"type:text" json:"test_cases"`
	Order        int    `gorm:"default:0" json:"order"`
	IsAIGen      bool   `gorm:"default:false" json:"is_ai_gen"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserProgress struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"uniqueIndex:idx_user_lesson" json:"user_id"`
	LessonID    uint       `gorm:"uniqueIndex:idx_user_lesson" json:"lesson_id"`
	Completed   bool       `gorm:"default:false" json:"completed"`
	Score       int        `gorm:"default:0" json:"score"`
	XPEarned    int        `gorm:"default:0" json:"xp_earned"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Submission struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	ExerciseID uint      `gorm:"index;not null" json:"exercise_id"`
	Answer     string    `gorm:"type:text" json:"answer"`
	Correct    bool      `json:"correct"`
	CreatedAt  time.Time `json:"created_at"`
}

// WrongExercise 错题本：记录用户答错的习题，支持回顾和已掌握标记
type WrongExercise struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"uniqueIndex:idx_user_exercise_wrong" json:"user_id"`
	ExerciseID   uint       `gorm:"uniqueIndex:idx_user_exercise_wrong" json:"exercise_id"`
	UserAnswer   string     `gorm:"type:text" json:"user_answer"`
	WrongCount   int        `gorm:"default:1" json:"wrong_count"`
	Mastered     bool       `gorm:"default:false" json:"mastered"`
	LastWrongAt time.Time  `json:"last_wrong_at"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// XPEvent XP 变动流水（排行榜与学习日历的数据源）
type XPEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Amount    int       `gorm:"not null" json:"amount"`
	Reason    string    `gorm:"size:20;not null" json:"reason"` // lesson / exercise / exam
	CreatedAt time.Time `json:"created_at"`
}

// Exam 考试（单元考试 / 认证考试）
type Exam struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CourseID      uint      `gorm:"index;not null" json:"course_id"`
	UnitID        uint      `gorm:"index" json:"unit_id"`
	Title         string    `gorm:"size:200;not null" json:"title"`
	ExamType      string    `gorm:"size:20;not null;default:unit" json:"exam_type"` // unit / cert
	DurationMin   int       `gorm:"default:15" json:"duration_min"`
	QuestionCount int       `gorm:"default:10" json:"question_count"`
	PassScore     int       `gorm:"default:60" json:"pass_score"`
	CreatedAt     time.Time `json:"created_at"`
}

// ExamQuestion 考试题目快照（保证同一场考试试卷稳定）
type ExamQuestion struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	ExamID     uint `gorm:"index;not null" json:"exam_id"`
	ExerciseID uint `gorm:"not null" json:"exercise_id"`
	Order      int  `gorm:"default:0" json:"order"`
}

// ExamSubmission 考试提交记录
type ExamSubmission struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	ExamID       uint      `gorm:"index;not null" json:"exam_id"`
	Score        int       `gorm:"not null" json:"score"`
	CorrectCount int       `gorm:"not null" json:"correct_count"`
	TotalCount   int       `gorm:"not null" json:"total_count"`
	DurationSec  int       `gorm:"default:0" json:"duration_sec"`
	Passed       bool      `gorm:"not null" json:"passed"`
	CreatedAt    time.Time `json:"created_at"`
}

// Certificate 能力认证证书
type Certificate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	CourseID    uint      `gorm:"index;not null" json:"course_id"`
	CertNo      string    `gorm:"uniqueIndex;size:40;not null" json:"cert_no"`
	Level       string    `gorm:"size:20;not null" json:"level"` // beginner / intermediate / advanced
	Score       int       `gorm:"not null" json:"score"`
	CourseTitle string    `gorm:"size:200;not null" json:"course_title"`
	UserName    string    `gorm:"size:50;not null" json:"user_name"`
	IssuedAt    time.Time `json:"issued_at"`
	CreatedAt   time.Time `json:"created_at"`
}
