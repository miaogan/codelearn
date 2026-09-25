package service

import (
	"testing"
	"time"

	"codelearn/model"
)

func TestRecordExerciseXP_Cap(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewProgressService(repo, 20, 10, 5)

	user := seedUser(db)

	// 记录超过每日上限的练习 XP
	for i := 0; i < MaxExerciseXPEventsPerDay+5; i++ {
		svc.RecordExerciseXP(user.ID)
	}

	var count int64
	db.Model(&model.XPEvent{}).Where("user_id = ? AND reason = ?", user.ID, XPReasonExercise).Count(&count)
	if count != MaxExerciseXPEventsPerDay {
		t.Errorf("expected %d exercise XP events, got %d", MaxExerciseXPEventsPerDay, count)
	}
}

func TestCompleteLesson_RecordsXPEvent(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewProgressService(repo, 20, 10, 5)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)

	svc.CompleteLesson(user.ID, lessonID, 100)

	var events []model.XPEvent
	db.Where("user_id = ?", user.ID).Find(&events)
	if len(events) != 1 {
		t.Fatalf("expected 1 XP event, got %d", len(events))
	}
	if events[0].Reason != XPReasonLesson || events[0].Amount != 20 {
		t.Errorf("unexpected XP event: reason=%s amount=%d", events[0].Reason, events[0].Amount)
	}
}

func TestWeeklyLeaderboard(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewLeaderboardService(repo)

	userA := seedUser(db)
	userB := &model.User{Username: "userb", Email: "b@example.com", Hearts: 5, MaxHearts: 5}
	db.Create(userB)

	now := time.Now()
	db.Create(&model.XPEvent{UserID: userA.ID, Amount: 40, Reason: XPReasonLesson, CreatedAt: now})
	db.Create(&model.XPEvent{UserID: userA.ID, Amount: 10, Reason: XPReasonExercise, CreatedAt: now})
	db.Create(&model.XPEvent{UserID: userB.ID, Amount: 25, Reason: XPReasonLesson, CreatedAt: now})

	entries, err := svc.WeeklyLeaderboard(20)
	if err != nil {
		t.Fatalf("WeeklyLeaderboard failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].UserID != userA.ID {
		t.Errorf("expected userA first (50 XP), got user %d with %d XP", entries[0].UserID, entries[0].XP)
	}
	if entries[0].XP != 50 {
		t.Errorf("expected 50 XP for userA, got %d", entries[0].XP)
	}
}

func TestMyRank(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewLeaderboardService(repo)

	userA := seedUser(db)
	userB := &model.User{Username: "userb", Email: "b@example.com", Hearts: 5, MaxHearts: 5}
	db.Create(userB)

	now := time.Now()
	db.Create(&model.XPEvent{UserID: userA.ID, Amount: 30, Reason: XPReasonLesson, CreatedAt: now})
	db.Create(&model.XPEvent{UserID: userB.ID, Amount: 60, Reason: XPReasonLesson, CreatedAt: now})

	me, err := svc.MyRank(userA.ID)
	if err != nil {
		t.Fatalf("MyRank failed: %v", err)
	}
	if me.XP != 30 {
		t.Errorf("expected 30 XP, got %d", me.XP)
	}
	if me.Rank != 2 {
		t.Errorf("expected rank 2 (one user ahead), got %d", me.Rank)
	}
}

func TestCalendar(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewLeaderboardService(repo)

	user := seedUser(db)
	now := time.Now()
	db.Create(&model.XPEvent{UserID: user.ID, Amount: 20, Reason: XPReasonLesson, CreatedAt: now})
	db.Create(&model.XPEvent{UserID: user.ID, Amount: 10, Reason: XPReasonExercise, CreatedAt: now.Add(-time.Hour * 48)})

	days, err := svc.Calendar(user.ID, now.Year(), int(now.Month()))
	if err != nil {
		t.Fatalf("Calendar failed: %v", err)
	}
	// 今天和两天前各应有一条记录（若跨月则只断言今天）
	totalXP := 0
	for _, d := range days {
		totalXP += d.XP
	}
	if totalXP < 30 {
		t.Errorf("expected at least 30 XP in calendar, got %d", totalXP)
	}
}
