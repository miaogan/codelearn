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
