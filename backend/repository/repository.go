package repository

import (
	"time"

	"codelearn/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB { return r.db }

// User

func (r *Repository) CreateUser(u *model.User) error {
	return r.db.Create(u).Error
}

func (r *Repository) GetUserByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Where("username = ?", username).First(&u).Error
	return &u, err
}

func (r *Repository) GetUserByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.Where("email = ?", email).First(&u).Error
	return &u, err
}

func (r *Repository) GetUserByID(id uint) (*model.User, error) {
	var u model.User
	err := r.db.First(&u, id).Error
	return &u, err
}

func (r *Repository) UpdateUser(u *model.User) error {
	return r.db.Save(u).Error
}

// Course

func (r *Repository) ListCourses() ([]model.Course, error) {
	var courses []model.Course
	err := r.db.Order("\"order\" ASC").Find(&courses).Error
	return courses, err
}

func (r *Repository) GetCourse(id uint) (*model.Course, error) {
	var c model.Course
	err := r.db.First(&c, id).Error
	return &c, err
}

// Unit

func (r *Repository) ListUnitsByCourse(courseID uint) ([]model.Unit, error) {
	var units []model.Unit
	err := r.db.Where("course_id = ?", courseID).Order("\"order\" ASC").Find(&units).Error
	return units, err
}

func (r *Repository) GetUnit(id uint) (*model.Unit, error) {
	var u model.Unit
	err := r.db.First(&u, id).Error
	return &u, err
}

func (r *Repository) ListLessonsByUnit(unitID uint) ([]model.Lesson, error) {
	var lessons []model.Lesson
	err := r.db.Where("unit_id = ?", unitID).Order("\"order\" ASC").Find(&lessons).Error
	return lessons, err
}

func (r *Repository) GetLesson(id uint) (*model.Lesson, error) {
	var l model.Lesson
	err := r.db.First(&l, id).Error
	return &l, err
}

// Exercise

func (r *Repository) ListExercisesByLesson(lessonID uint) ([]model.Exercise, error) {
	var exercises []model.Exercise
	err := r.db.Where("lesson_id = ?", lessonID).Order("\"order\" ASC").Find(&exercises).Error
	return exercises, err
}

func (r *Repository) GetExercise(id uint) (*model.Exercise, error) {
	var e model.Exercise
	err := r.db.First(&e, id).Error
	return &e, err
}

func (r *Repository) CreateExercise(e *model.Exercise) error {
	return r.db.Create(e).Error
}

func (r *Repository) CreateExercises(exercises []model.Exercise) error {
	if len(exercises) == 0 {
		return nil
	}
	return r.db.Create(&exercises).Error
}

// Progress

func (r *Repository) GetProgress(userID, lessonID uint) (*model.UserProgress, error) {
	var p model.UserProgress
	err := r.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).First(&p).Error
	return &p, err
}

func (r *Repository) UpsertProgress(p *model.UserProgress) error {
	var existing model.UserProgress
	result := r.db.Where("user_id = ? AND lesson_id = ?", p.UserID, p.LessonID).First(&existing)
	if result.Error != nil {
		return r.db.Create(p).Error
	}
	existing.Completed = p.Completed
	existing.Score = p.Score
	existing.XPEarned = p.XPEarned
	if p.Completed {
		now := time.Now()
		existing.CompletedAt = &now
	}
	return r.db.Save(&existing).Error
}

func (r *Repository) ListProgressByUser(userID uint) ([]model.UserProgress, error) {
	var progress []model.UserProgress
	err := r.db.Where("user_id = ?", userID).Find(&progress).Error
	return progress, err
}

func (r *Repository) CountCompletedLessons(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserProgress{}).Where("user_id = ? AND completed = ?", userID, true).Count(&count).Error
	return count, err
}

// Submission

func (r *Repository) CreateSubmission(s *model.Submission) error {
	return r.db.Create(s).Error
}

func (r *Repository) CountTodaySubmissions(userID uint) (int64, error) {
	var count int64
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	err := r.db.Model(&model.Submission{}).Where("user_id = ? AND created_at >= ?", userID, start).Count(&count).Error
	return count, err
}

// WrongExercise 错题本

func (r *Repository) UpsertWrongExercise(userID, exerciseID uint, userAnswer, source string) error {
	var existing model.WrongExercise
	result := r.db.Where("user_id = ? AND exercise_id = ?", userID, exerciseID).First(&existing)
	if result.Error != nil {
		// 不存在，新建
		return r.db.Create(&model.WrongExercise{
			UserID:     userID,
			ExerciseID: exerciseID,
			UserAnswer: userAnswer,
			WrongCount: 1,
			Source:     source,
			LastWrongAt: time.Now(),
		}).Error
	}
	// 已存在，增加错误次数，重置掌握状态
	existing.WrongCount++
	existing.Mastered = false
	existing.UserAnswer = userAnswer
	existing.LastWrongAt = time.Now()
	existing.ReviewedAt = nil
	if source != "" {
		existing.Source = source
	}
	return r.db.Save(&existing).Error
}

func (r *Repository) ListWrongExercises(userID uint, onlyUnmastered bool) ([]model.WrongExercise, error) {
	var list []model.WrongExercise
	q := r.db.Where("user_id = ?", userID)
	if onlyUnmastered {
		q = q.Where("mastered = ?", false)
	}
	err := q.Order("last_wrong_at DESC").Find(&list).Error
	return list, err
}

func (r *Repository) GetWrongExercise(userID, exerciseID uint) (*model.WrongExercise, error) {
	var w model.WrongExercise
	err := r.db.Where("user_id = ? AND exercise_id = ?", userID, exerciseID).First(&w).Error
	return &w, err
}

func (r *Repository) MarkWrongExerciseMastered(userID, exerciseID uint) error {
	now := time.Now()
	return r.db.Model(&model.WrongExercise{}).
		Where("user_id = ? AND exercise_id = ?", userID, exerciseID).
		Updates(map[string]interface{}{"mastered": true, "reviewed_at": &now}).Error
}

func (r *Repository) CountWrongExercises(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.WrongExercise{}).Where("user_id = ? AND mastered = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *Repository) ListExercisesByIDs(ids []uint) ([]model.Exercise, error) {
	var exercises []model.Exercise
	if len(ids) == 0 {
		return exercises, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&exercises).Error
	return exercises, err
}

// ListLessonsByIDs 批量获取课时
func (r *Repository) ListLessonsByIDs(ids []uint) ([]model.Lesson, error) {
	var lessons []model.Lesson
	if len(ids) == 0 {
		return lessons, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&lessons).Error
	return lessons, err
}

// ListSubmissionsForExercises 获取用户在指定习题集合上的所有提交记录
func (r *Repository) ListSubmissionsForExercises(userID uint, exerciseIDs []uint) ([]model.Submission, error) {
	var list []model.Submission
	if len(exerciseIDs) == 0 {
		return list, nil
	}
	err := r.db.Where("user_id = ? AND exercise_id IN ?", userID, exerciseIDs).
		Order("created_at ASC").Find(&list).Error
	return list, err
}

// LessonSearchResult 课程内容检索结果
type LessonSearchResult struct {
	ID         uint   `json:"id"`
	UnitID     uint   `json:"unit_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CourseID   uint   `json:"course_id"`
	CourseName string `json:"course_name"`
}

// SearchLessonsByKeyword 按关键词检索课程内容（RAG Retriever）
func (r *Repository) SearchLessonsByKeyword(keyword string, limit int) []LessonSearchResult {
	if keyword == "" {
		return []LessonSearchResult{}
	}
	if limit <= 0 {
		limit = 5
	}

	var results []LessonSearchResult
	// JOIN lessons + units + courses，按标题和内容模糊搜索
	r.db.Table("lessons").
		Select("lessons.id, lessons.unit_id, lessons.title, lessons.content, units.course_id, courses.title as course_name").
		Joins("JOIN units ON units.id = lessons.unit_id").
		Joins("JOIN courses ON courses.id = units.course_id").
		Where("lessons.title LIKE ? OR lessons.content LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Limit(limit).
		Scan(&results)

	return results
}

// ListExercisesByUnit 获取单元下所有课时包含的习题
func (r *Repository) ListExercisesByUnit(unitID uint) ([]model.Exercise, error) {
	var exercises []model.Exercise
	err := r.db.Table("exercises").
		Joins("JOIN lessons ON lessons.id = exercises.lesson_id").
		Where("lessons.unit_id = ?", unitID).
		Order("lessons.\"order\" ASC, exercises.\"order\" ASC").
		Find(&exercises).Error
	return exercises, err
}

func (r *Repository) GetCourseLanguage(courseID uint) (string, error) {
	var c model.Course
	err := r.db.Select("language").First(&c, courseID).Error
	return c.Language, err
}

// XPEvent 排行榜与学习日历

func (r *Repository) CreateXPEvent(e *model.XPEvent) error {
	return r.db.Create(e).Error
}

func (r *Repository) CountXPEventsByReasonSince(userID uint, reason string, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&model.XPEvent{}).
		Where("user_id = ? AND reason = ? AND created_at >= ?", userID, reason, since).
		Count(&count).Error
	return count, err
}

func (r *Repository) SumXPBetween(userID uint, start, end time.Time) (int, error) {
	var sum int64
	err := r.db.Model(&model.XPEvent{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sum).Error
	return int(sum), err
}

// LeaderboardRow 排行榜聚合行
type LeaderboardRow struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	XP       int    `json:"xp"`
}

func (r *Repository) WeeklyLeaderboard(since time.Time, limit int) ([]LeaderboardRow, error) {
	var rows []LeaderboardRow
	if limit <= 0 {
		limit = 20
	}
	err := r.db.Table("xp_events").
		Select("xp_events.user_id, users.username, SUM(xp_events.amount) as xp").
		Joins("JOIN users ON users.id = xp_events.user_id").
		Where("xp_events.created_at >= ?", since).
		Group("xp_events.user_id, users.username").
		Order("xp DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// CalendarDay 学习日历中的某一天
type CalendarDay struct {
	Date string `json:"date"`
	XP   int    `json:"xp"`
}

func (r *Repository) CalendarDays(userID uint, start, end time.Time) ([]CalendarDay, error) {
	var days []CalendarDay
	err := r.db.Table("xp_events").
		Select("date(created_at) as date, SUM(amount) as xp").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Group("date(created_at)").
		Order("date(created_at) ASC").
		Scan(&days).Error
	return days, err
}

// Exam 考试

func (r *Repository) CreateExam(e *model.Exam) error {
	return r.db.Create(e).Error
}

func (r *Repository) GetExam(id uint) (*model.Exam, error) {
	var e model.Exam
	err := r.db.First(&e, id).Error
	return &e, err
}

// ListExamsByUnit 获取单元下的所有考试
func (r *Repository) ListExamsByUnit(unitID uint) ([]model.Exam, error) {
	var list []model.Exam
	err := r.db.Where("unit_id = ?", unitID).Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *Repository) CreateExamQuestions(qs []model.ExamQuestion) error {
	if len(qs) == 0 {
		return nil
	}
	return r.db.Create(&qs).Error
}

func (r *Repository) ListExamQuestions(examID uint) ([]model.ExamQuestion, error) {
	var qs []model.ExamQuestion
	err := r.db.Where("exam_id = ?", examID).Order("\"order\" ASC").Find(&qs).Error
	return qs, err
}

// ExamSubmission 考试提交

func (r *Repository) CreateExamSubmission(s *model.ExamSubmission) error {
	return r.db.Create(s).Error
}

func (r *Repository) GetLatestExamSubmission(userID, examID uint) (*model.ExamSubmission, error) {
	var s model.ExamSubmission
	err := r.db.Where("user_id = ? AND exam_id = ?", userID, examID).Order("created_at DESC").First(&s).Error
	return &s, err
}

// ListExamSubmissions 获取用户在指定考试上的全部提交记录（按时间正序）
func (r *Repository) ListExamSubmissions(userID, examID uint) ([]model.ExamSubmission, error) {
	var list []model.ExamSubmission
	err := r.db.Where("user_id = ? AND exam_id = ?", userID, examID).Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *Repository) CountExamSubmissions(userID, examID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ExamSubmission{}).
		Where("user_id = ? AND exam_id = ?", userID, examID).
		Count(&count).Error
	return count, err
}

// Certificate 证书

func (r *Repository) CreateCertificate(c *model.Certificate) error {
	return r.db.Create(c).Error
}

func (r *Repository) GetCertificateByNo(certNo string) (*model.Certificate, error) {
	var c model.Certificate
	err := r.db.Where("cert_no = ?", certNo).First(&c).Error
	return &c, err
}

func (r *Repository) ListCertificatesByUser(userID uint) ([]model.Certificate, error) {
	var list []model.Certificate
	err := r.db.Where("user_id = ?", userID).Order("issued_at DESC").Find(&list).Error
	return list, err
}

func (r *Repository) GetCertificate(userID, courseID uint) (*model.Certificate, error) {
	var c model.Certificate
	err := r.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&c).Error
	return &c, err
}

func (r *Repository) CountCertificates(userID, courseID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Certificate{}).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Count(&count).Error
	return count, err
}
