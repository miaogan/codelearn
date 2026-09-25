package service

import (
	"strings"
	"testing"
)

func TestWeeklyReport_NoData(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewWeeklyReportService(repo)

	user := seedUser(db)

	report, err := svc.Report(user.ID)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}
	if report.ActiveDays != 0 {
		t.Errorf("expected 0 active days, got %d", report.ActiveDays)
	}
	if !strings.Contains(report.Advice, "还没有学习记录") {
		t.Errorf("expected empty-state advice, got %q", report.Advice)
	}
	if report.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestWeeklyReport_WithData(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewWeeklyReportService(repo)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)

	// 完成一个课时（记录进度 + XP）
	progressSvc := NewProgressService(repo, 20, 10, 5)
	progressSvc.CompleteLesson(user.ID, lessonID, 100)

	// 新增一道错题 → 薄弱点应含该课时
	ex := createExercise(db, lessonID)
	repo.UpsertWrongExercise(user.ID, ex.ID, "B", "exercise")

	report, err := svc.Report(user.ID)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}
	if report.ActiveDays < 1 {
		t.Errorf("expected >=1 active day, got %d", report.ActiveDays)
	}
	if report.LessonsDone < 1 {
		t.Errorf("expected >=1 lesson done, got %d", report.LessonsDone)
	}
	if len(report.WeakTopics) == 0 {
		t.Error("expected weak topics to include the lesson with wrong exercise")
	}
	if report.Advice == "" || report.Summary == "" {
		t.Error("expected non-empty summary/advice")
	}
}

func TestWeakTopics_MaxThree(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewWeeklyReportService(repo)

	user := seedUser(db)
	// 创建 2 个课时各若干错题
	_, _, lesson1 := seedCourse(db)
	_, _, lesson2 := seedCourse(db)
	for i := 0; i < 3; i++ {
		ex1 := createExercise(db, lesson1)
		ex2 := createExercise(db, lesson2)
		repo.UpsertWrongExercise(user.ID, ex1.ID, "B", "exercise")
		repo.UpsertWrongExercise(user.ID, ex2.ID, "C", "exercise")
	}
	topics := svc.weakTopics(user.ID)
	if len(topics) > 2 {
		t.Errorf("expected at most 2 distinct lessons, got %d topics: %+v", len(topics), topics)
	}
}
