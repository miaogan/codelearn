package service

import (
	"testing"
	"time"

	"codelearn/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建内存 SQLite 数据库并迁移表结构
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Course{}, &model.Unit{}, &model.Lesson{}, &model.Exercise{}, &model.UserProgress{}, &model.Submission{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

// seedCourse 创建测试课程数据
func seedCourse(db *gorm.DB) (courseID, unitID, lessonID uint) {
	course := model.Course{Language: "go", Title: "Go 基础", Emoji: "🐹", Color: "#00ADD8", Order: 0}
	db.Create(&course)

	unit := model.Unit{CourseID: course.ID, Title: "Unit 1", Icon: "📌", Color: "#4CAF50", Order: 0}
	db.Create(&unit)

	lesson := model.Lesson{UnitID: unit.ID, Title: "Hello World", Icon: "👋", Order: 0}
	db.Create(&lesson)

	return course.ID, unit.ID, lesson.ID
}

// seedUser 创建测试用户
func seedUser(db *gorm.DB) *model.User {
	user := model.User{
		Username: "testuser",
		Email:    "test@example.com",
		XP:       0,
		Hearts:   5,
		MaxHearts: 5,
		StreakDays: 0,
	}
	db.Create(&user)
	return &user
}

// ============ ProgressService 测试 ============

func TestCompleteLesson_FirstTime(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewProgressService(repo, 20, 10, 5)

	courseID, _, lessonID := seedCourse(db)
	user := seedUser(db)
	_ = courseID

	xp, err := svc.CompleteLesson(user.ID, lessonID, 100)
	if err != nil {
		t.Fatalf("CompleteLesson failed: %v", err)
	}
	if xp != 20 {
		t.Errorf("expected 20 XP, got %d", xp)
	}

	updated, _ := repo.GetUserByID(user.ID)
	if updated.XP != 20 {
		t.Errorf("expected user XP=20, got %d", updated.XP)
	}
	if updated.StreakDays != 1 {
		t.Errorf("expected streak=1, got %d", updated.StreakDays)
	}
}

func TestCompleteLesson_DuplicateNoXP(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewProgressService(repo, 20, 10, 5)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)

	svc.CompleteLesson(user.ID, lessonID, 100)
	xp, err := svc.CompleteLesson(user.ID, lessonID, 90)
	if err != nil {
		t.Fatalf("second CompleteLesson failed: %v", err)
	}
	if xp != 0 {
		t.Errorf("expected 0 XP for duplicate, got %d", xp)
	}

	updated, _ := repo.GetUserByID(user.ID)
	if updated.XP != 20 {
		t.Errorf("expected XP=20 (no double), got %d", updated.XP)
	}
}

func TestLoseHeart(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewProgressService(repo, 20, 10, 5)

	seedCourse(db)
	user := seedUser(db)

	hearts, err := svc.LoseHeart(user.ID)
	if err != nil {
		t.Fatalf("LoseHeart failed: %v", err)
	}
	if hearts != 4 {
		t.Errorf("expected 4 hearts, got %d", hearts)
	}
}

func TestLoseHeart_ZeroFloor(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewProgressService(repo, 20, 10, 5)

	seedCourse(db)
	user := seedUser(db)
	user.Hearts = 0
	db.Save(user)

	hearts, err := svc.LoseHeart(user.ID)
	if err != nil {
		t.Fatalf("LoseHeart failed: %v", err)
	}
	if hearts != 0 {
		t.Errorf("expected 0 hearts (floor), got %d", hearts)
	}
}

func TestRestoreHeart(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewProgressService(repo, 20, 10, 5)

	seedCourse(db)
	user := seedUser(db)
	user.Hearts = 3
	db.Save(user)

	hearts, err := svc.RestoreHeart(user.ID)
	if err != nil {
		t.Fatalf("RestoreHeart failed: %v", err)
	}
	if hearts != 4 {
		t.Errorf("expected 4 hearts, got %d", hearts)
	}
}

func TestRestoreHeart_MaxCap(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewProgressService(repo, 20, 10, 5)

	seedCourse(db)
	user := seedUser(db)

	hearts, err := svc.RestoreHeart(user.ID)
	if err != nil {
		t.Fatalf("RestoreHeart failed: %v", err)
	}
	if hearts != 5 {
		t.Errorf("expected 5 hearts (max cap), got %d", hearts)
	}
}

func TestGetStats(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewProgressService(repo, 20, 10, 5)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)

	svc.CompleteLesson(user.ID, lessonID, 100)

	stats, err := svc.GetStats(user.ID)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats.XP != 20 {
		t.Errorf("expected XP=20, got %d", stats.XP)
	}
	if stats.Hearts != 5 {
		t.Errorf("expected hearts=5, got %d", stats.Hearts)
	}
	if stats.MaxHearts != 5 {
		t.Errorf("expected maxHearts=5, got %d", stats.MaxHearts)
	}
	if stats.StreakDays != 1 {
		t.Errorf("expected streak=1, got %d", stats.StreakDays)
	}
	if stats.CompletedToday != 1 {
		t.Errorf("expected 1 completed, got %d", stats.CompletedToday)
	}
}

func TestUpdateStreak_FirstTime(t *testing.T) {
	user := &model.User{StreakDays: 0, LastStreakAt: nil}
	updateStreak(user)

	if user.StreakDays != 1 {
		t.Errorf("expected streak=1, got %d", user.StreakDays)
	}
	if user.LastStreakAt == nil {
		t.Error("expected LastStreakAt to be set")
	}
}

func TestUpdateStreak_SameDay(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	user := &model.User{StreakDays: 3, LastStreakAt: &today}

	updateStreak(user)

	if user.StreakDays != 3 {
		t.Errorf("expected streak=3 (same day), got %d", user.StreakDays)
	}
}

func TestUpdateStreak_ConsecutiveDay(t *testing.T) {
	now := time.Now()
	yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	user := &model.User{StreakDays: 3, LastStreakAt: &yesterday}

	updateStreak(user)

	if user.StreakDays != 4 {
		t.Errorf("expected streak=4 (consecutive), got %d", user.StreakDays)
	}
}

func TestUpdateStreak_Broken(t *testing.T) {
	now := time.Now()
	threeDaysAgo := time.Date(now.Year(), now.Month(), now.Day()-3, 0, 0, 0, 0, now.Location())
	user := &model.User{StreakDays: 5, LastStreakAt: &threeDaysAgo}

	updateStreak(user)

	if user.StreakDays != 1 {
		t.Errorf("expected streak=1 (broken), got %d", user.StreakDays)
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  Hello  ", "hello"},
		{"GoLang", "golang"},
		{"  A  ", "a"},
		{"", ""},
	}
	for _, tt := range tests {
		got := normalize(tt.input)
		if got != tt.want {
			t.Errorf("normalize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
