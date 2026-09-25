package service

import (
	"testing"

	"codelearn/model"
)

func TestTitleForXP(t *testing.T) {
	cases := []struct {
		xp   int
		name string
	}{
		{0, "代码萌芽"},
		{99, "代码萌芽"},
		{100, "代码学徒"},
		{299, "代码学徒"},
		{300, "代码骑士"},
		{599, "代码骑士"},
		{600, "代码剑士"},
		{999, "代码剑士"},
		{1000, "代码大师"},
		{1999, "代码大师"},
		{2000, "代码宗师"},
		{99999, "代码宗师"},
	}
	for _, c := range cases {
		name, _, _ := TitleForXP(c.xp)
		if name != c.name {
			t.Errorf("TitleForXP(%d) = %s, want %s", c.xp, name, c.name)
		}
	}
}

func TestUnlock_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAchievementService(repo)

	user := seedUser(db)

	if !svc.Unlock(user.ID, "first_lesson") {
		t.Error("expected first unlock to return true")
	}
	if svc.Unlock(user.ID, "first_lesson") {
		t.Error("expected duplicate unlock to return false")
	}
	var count int64
	db.Model(&model.Achievement{}).Where("user_id = ? AND code = ?", user.ID, "first_lesson").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 achievement row, got %d", count)
	}
}

func TestOnLessonCompleted_UnlocksFirstLesson(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAchievementService(repo)

	user := seedUser(db)
	svc.OnLessonCompleted(user.ID)

	var count int64
	db.Model(&model.Achievement{}).Where("user_id = ? AND code = ?", user.ID, "first_lesson").Count(&count)
	if count != 1 {
		t.Errorf("expected first_lesson unlocked, got %d", count)
	}
}

func TestOnExamPassed_Perfect(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAchievementService(repo)

	user := seedUser(db)
	svc.OnExamPassed(user.ID, 100)

	svc.OnExamPassed(user.ID, 80)
	// 满分徽章应解锁；首考徽章也应解锁
	for _, code := range []string{"first_exam_pass", "perfect_exam"} {
		var count int64
		db.Model(&model.Achievement{}).Where("user_id = ? AND code = ?", user.ID, code).Count(&count)
		if count != 1 {
			t.Errorf("expected %s unlocked, got %d", code, count)
		}
	}
}

func TestOnCertIssued(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAchievementService(repo)

	user := seedUser(db)
	svc.OnCertIssued(user.ID)

	var count int64
	db.Model(&model.Achievement{}).Where("user_id = ? AND code = ?", user.ID, "first_cert").Count(&count)
	if count != 1 {
		t.Errorf("expected first_cert unlocked, got %d", count)
	}
}

func TestPracticeBadges(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAchievementService(repo)

	user := seedUser(db)
	// 记录 50 次练习 XP 事件 → practice_50
	for i := 0; i < 55; i++ {
		db.Create(&model.XPEvent{UserID: user.ID, Amount: 10, Reason: XPReasonExercise})
	}
	svc.OnPracticeCorrect(user.ID)

	var count int64
	db.Model(&model.Achievement{}).Where("user_id = ? AND code = ?", user.ID, "practice_50").Count(&count)
	if count != 1 {
		t.Errorf("expected practice_50 unlocked, got %d", count)
	}
}

func TestSummary(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAchievementService(repo)

	user := seedUser(db)
	user.XP = 350
	db.Save(user)
	svc.Unlock(user.ID, "first_lesson")

	sum, err := svc.Summary(user.ID)
	if err != nil {
		t.Fatalf("Summary failed: %v", err)
	}
	if sum.Title.Name != "代码骑士" {
		t.Errorf("expected 代码骑士, got %s", sum.Title.Name)
	}
	if sum.Total != 1 {
		t.Errorf("expected 1 badge, got %d", sum.Total)
	}
	if sum.Title.XPToNext <= 0 {
		t.Errorf("expected positive xp_to_next (next=600), got %d", sum.Title.XPToNext)
	}
	if sum.Title.NextName != "代码剑士" {
		t.Errorf("expected next 代码剑士, got %s", sum.Title.NextName)
	}
}
