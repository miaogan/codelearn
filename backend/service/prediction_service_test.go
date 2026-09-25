package service

import (
	"testing"
	"time"

	"codelearn/model"
	"codelearn/repository"

	"gorm.io/gorm"
)

// seedUnitExamWithSubmission 为指定单元创建考试与一次提交
func seedUnitExamWithSubmission(t *testing.T, db *gorm.DB, courseID, unitID, userID uint, score int, passed bool, at time.Time) {
	t.Helper()
	exam := model.Exam{CourseID: courseID, UnitID: unitID, Title: "Unit Exam", ExamType: "unit", PassScore: 60, QuestionCount: 8}
	if err := db.Create(&exam).Error; err != nil {
		t.Fatalf("create exam: %v", err)
	}
	sub := model.ExamSubmission{
		UserID: userID, ExamID: exam.ID, Score: score,
		CorrectCount: 0, TotalCount: 8, Passed: passed, CreatedAt: at,
	}
	if err := db.Create(&sub).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}
}

func findExamFactor(p *ExamPrediction) *PredictionFactor {
	for i := range p.Factors {
		if p.Factors[i].Label == "单元考试成绩" {
			return &p.Factors[i]
		}
	}
	return nil
}

// P2-1：完全无数据时中性返回 50，而不是误导性的 0
func TestPredict_NoDataNeutral(t *testing.T) {
	db := setupTestDB(t)
	courseID, _, _ := seedCourse(db)
	user := seedUser(db)

	svc := NewPredictionService(repository.New(db))
	p, err := svc.Predict(user.ID, courseID)
	if err != nil {
		t.Fatalf("predict: %v", err)
	}
	if p.Probability != 50 {
		t.Errorf("expected neutral 50 with no data, got %d", p.Probability)
	}
	if p.Eligible {
		t.Error("expected not eligible with no data")
	}
	examFactor := findExamFactor(p)
	if examFactor == nil {
		t.Fatal("exam factor missing")
	}
	if examFactor.Detail != "0 个单元有考试记录，平均 0 分" {
		t.Errorf("unexpected exam factor detail: %s", examFactor.Detail)
	}
}

// P2-1：挂科尝试也计入单元考试平均分（避免只看通过场次的乐观偏差）
func TestPredict_FailedAttemptCountsInAverage(t *testing.T) {
	db := setupTestDB(t)
	courseID, unitID, _ := seedCourse(db)
	user := seedUser(db)

	seedUnitExamWithSubmission(t, db, courseID, unitID, user.ID, 40, false, time.Now())

	svc := NewPredictionService(repository.New(db))
	p, err := svc.Predict(user.ID, courseID)
	if err != nil {
		t.Fatalf("predict: %v", err)
	}
	if p.UnitsPassed != 0 {
		t.Errorf("expected 0 units passed, got %d", p.UnitsPassed)
	}
	if p.Eligible {
		t.Error("expected not eligible with failed exam")
	}
	// 有数据时如实反映：概率 = 40*30/100 = 12，而不是中性 50
	if p.Probability != 12 {
		t.Errorf("expected prob 12 from failed attempt, got %d", p.Probability)
	}
	examFactor := findExamFactor(p)
	if examFactor == nil {
		t.Fatal("exam factor missing")
	}
	if examFactor.Score != 40 {
		t.Errorf("expected avg 40 (failed attempt included), got %d", examFactor.Score)
	}
	if examFactor.Detail != "1 个单元有考试记录，平均 40 分" {
		t.Errorf("unexpected exam factor detail: %s", examFactor.Detail)
	}
}

// P2-1：通过单元与挂科单元混合时，平均分同时包含两者
func TestPredict_PassedAndFailedMix(t *testing.T) {
	db := setupTestDB(t)
	courseID, unitID, _ := seedCourse(db)
	unit2 := model.Unit{CourseID: courseID, Title: "Unit 2", Icon: "📌", Color: "#4CAF50", Order: 1}
	if err := db.Create(&unit2).Error; err != nil {
		t.Fatalf("create unit2: %v", err)
	}
	user := seedUser(db)

	now := time.Now()
	seedUnitExamWithSubmission(t, db, courseID, unitID, user.ID, 85, true, now.Add(-time.Hour))
	seedUnitExamWithSubmission(t, db, courseID, unit2.ID, user.ID, 50, false, now)

	svc := NewPredictionService(repository.New(db))
	p, err := svc.Predict(user.ID, courseID)
	if err != nil {
		t.Fatalf("predict: %v", err)
	}
	if p.UnitsPassed != 1 {
		t.Errorf("expected 1 unit passed, got %d", p.UnitsPassed)
	}
	examFactor := findExamFactor(p)
	if examFactor == nil {
		t.Fatal("exam factor missing")
	}
	if examFactor.Score != 67 { // (85+50)/2
		t.Errorf("expected avg 67, got %d", examFactor.Score)
	}
	if examFactor.Detail != "2 个单元有考试记录，平均 67 分" {
		t.Errorf("unexpected exam factor detail: %s", examFactor.Detail)
	}
}
