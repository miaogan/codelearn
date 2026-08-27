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

// Lesson

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

func (r *Repository) UpsertWrongExercise(userID, exerciseID uint, userAnswer string) error {
	var existing model.WrongExercise
	result := r.db.Where("user_id = ? AND exercise_id = ?", userID, exerciseID).First(&existing)
	if result.Error != nil {
		// 不存在，新建
		return r.db.Create(&model.WrongExercise{
			UserID:     userID,
			ExerciseID: exerciseID,
			UserAnswer: userAnswer,
			WrongCount: 1,
			LastWrongAt: time.Now(),
		}).Error
	}
	// 已存在，增加错误次数，重置掌握状态
	existing.WrongCount++
	existing.Mastered = false
	existing.UserAnswer = userAnswer
	existing.LastWrongAt = time.Now()
	existing.ReviewedAt = nil
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
