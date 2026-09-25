package service

import (
	"errors"
	"testing"
	"time"

	"codelearn/model"
)

func TestFunnel_OrderAndDedup(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAnalyticsService(repo)

	userA := seedUser(db)
	userB := &model.User{Username: "userb", Email: "b@example.com", Hearts: 5, MaxHearts: 5}
	db.Create(userB)

	now := time.Now()
	// userA 走完全流程，注册埋点两次（去重后应只计 1）
	svc.Record(EventRegister, userA.ID, 0)
	svc.Record(EventRegister, userA.ID, 0)
	svc.Record(EventLessonComplete, userA.ID, 1)
	svc.Record(EventUnitExamPassed, userA.ID, 1)
	svc.Record(EventCertIssued, userA.ID, 1)
	// userB 只注册
	svc.Record(EventRegister, userB.ID, 0)
	_ = now

	funnel := svc.Funnel(30)
	if len(funnel) != 4 {
		t.Fatalf("expected 4 funnel rows, got %d", len(funnel))
	}

	expect := []struct {
		key   string
		users int64
	}{
		{EventRegister, 2},
		{EventLessonComplete, 1},
		{EventUnitExamPassed, 1},
		{EventCertIssued, 1},
	}
	for i, e := range expect {
		if funnel[i].Key != e.key {
			t.Errorf("row %d: expected key %s, got %s", i, e.key, funnel[i].Key)
		}
		if funnel[i].Users != e.users {
			t.Errorf("row %d (%s): expected %d users, got %d", i, e.key, e.users, funnel[i].Users)
		}
	}
}

func TestFunnel_DefaultWindow(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAnalyticsService(repo)

	user := seedUser(db)
	// 30 天内的注册应计入；90 天前的事件应被窗口过滤
	svc.Record(EventRegister, user.ID, 0)
	db.Create(&model.AnalyticsEvent{
		UserID:    user.ID,
		EventType: EventRegister,
		CreatedAt: time.Now().AddDate(0, 0, -90),
	})

	// days<=0 时回退默认 30 天
	funnel := svc.Funnel(0)
	if len(funnel) != 4 {
		t.Fatalf("expected 4 funnel rows, got %d", len(funnel))
	}
	if funnel[0].Users != 1 {
		t.Errorf("expected 1 registered user within window, got %d", funnel[0].Users)
	}
}

func TestFunnel_NoData(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAnalyticsService(repo)

	funnel := svc.Funnel(30)
	if len(funnel) != 4 {
		t.Fatalf("expected 4 funnel rows even without data, got %d", len(funnel))
	}
	for _, r := range funnel {
		if r.Users != 0 {
			t.Errorf("expected 0 users for %s, got %d", r.Key, r.Users)
		}
	}
}

func TestRecord_IgnoresInvalid(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAnalyticsService(repo)

	svc.Record("", 1, 0)
	svc.Record(EventRegister, 0, 0)

	var count int64
	db.Model(&model.AnalyticsEvent{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 events for invalid inputs, got %d", count)
	}
}

func TestSubmitFeedback(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAnalyticsService(repo)

	user := seedUser(db)

	if err := svc.SubmitFeedback(user.ID, "bug", "页面报错", "wx123"); err != nil {
		t.Fatalf("SubmitFeedback failed: %v", err)
	}

	var fb model.UserFeedback
	if err := db.First(&fb).Error; err != nil {
		t.Fatalf("failed to load feedback: %v", err)
	}
	if fb.UserID != user.ID || fb.Category != "bug" || fb.Content != "页面报错" || fb.Contact != "wx123" {
		t.Errorf("unexpected feedback record: %+v", fb)
	}
	if fb.Status != "open" {
		t.Errorf("expected status open, got %s", fb.Status)
	}
}

func TestSubmitFeedback_EmptyContent(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewAnalyticsService(repo)

	user := seedUser(db)
	err := svc.SubmitFeedback(user.ID, "bug", "   ", "")
	if !errors.Is(err, ErrFeedbackEmpty) {
		t.Fatalf("expected ErrFeedbackEmpty, got %v", err)
	}

	var count int64
	db.Model(&model.UserFeedback{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no feedback saved, got %d", count)
	}
}
