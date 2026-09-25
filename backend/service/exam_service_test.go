package service

import (
	"encoding/json"
	"errors"
	"testing"

	"codelearn/model"

	"gorm.io/gorm"
)

// seedExamUnit 创建带习题的考试测试单元
// completed=true 时将该用户的所有课时标记为已完成
func seedExamUnit(db *gorm.DB, userID uint, completed bool) uint {
	course := model.Course{Language: "python", Title: "Python 基础", Emoji: "🐍", Color: "#3776AB", Order: 0}
	db.Create(&course)

	unit := model.Unit{CourseID: course.ID, Title: "Unit 1", Icon: "📌", Color: "#4CAF50", Order: 0, ExamDurationMin: 15, ExamQuestionCount: 4}
	db.Create(&unit)

	lesson1 := model.Lesson{UnitID: unit.ID, Title: "L1", Icon: "1", Order: 0}
	db.Create(&lesson1)
	lesson2 := model.Lesson{UnitID: unit.ID, Title: "L2", Icon: "2", Order: 1}
	db.Create(&lesson2)

	optsA, _ := json.Marshal([]string{"a", "b", "c"})
	optsX, _ := json.Marshal([]string{"x", "y"})
	exs := []model.Exercise{
		{LessonID: lesson1.ID, Type: "choice", Question: "选择 a", Options: string(optsA), Answer: "a", Explanation: "因为 a", Difficulty: "easy", Order: 1},
		{LessonID: lesson1.ID, Type: "fillblank", Question: "填空 42", Answer: "42", Explanation: "因为 42", Difficulty: "easy", Order: 2},
		{LessonID: lesson2.ID, Type: "choice", Question: "选择 x", Options: string(optsX), Answer: "x", Explanation: "因为 x", Difficulty: "easy", Order: 1},
	}
	db.Create(&exs)

	if completed {
		db.Create(&model.UserProgress{UserID: userID, LessonID: lesson1.ID, Completed: true, Score: 100})
		db.Create(&model.UserProgress{UserID: userID, LessonID: lesson2.ID, Completed: true, Score: 100})
	}
	return unit.ID
}

func TestAssembleUnitExam_LessonsNotCompleted(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewExamService(repo)

	user := seedUser(db)
	unitID := seedExamUnit(db, user.ID, false)

	_, err := svc.AssembleUnitExam(unitID, user.ID)
	if !errors.Is(err, ErrUnitNotCompleted) {
		t.Fatalf("expected ErrUnitNotCompleted, got %v", err)
	}
}

func TestAssembleUnitExam_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewExamService(repo)

	user := seedUser(db)
	unitID := seedExamUnit(db, user.ID, true)

	exam, err := svc.AssembleUnitExam(unitID, user.ID)
	if err != nil {
		t.Fatalf("AssembleUnitExam failed: %v", err)
	}
	if exam.ID == 0 {
		t.Error("expected exam ID to be set")
	}
	if len(exam.Questions) != 3 {
		t.Errorf("expected 3 questions, got %d", len(exam.Questions))
	}
	if exam.PassScore != 60 {
		t.Errorf("expected pass score 60, got %d", exam.PassScore)
	}
	// 下发的考题不得包含答案
	for _, q := range exam.Questions {
		if q.Options == "" && q.Type != "fillblank" {
			t.Errorf("question %d missing options", q.ID)
		}
	}
}

func TestSubmitExam_Scoring(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewExamService(repo)

	user := seedUser(db)
	unitID := seedExamUnit(db, user.ID, true)

	exam, err := svc.AssembleUnitExam(unitID, user.ID)
	if err != nil {
		t.Fatalf("AssembleUnitExam failed: %v", err)
	}

	// 前两题答对，第三题答错 => 66 分，通过
	// 注意：映射键为习题 ID（种子数据固定为 1/2/3），与试卷洗牌顺序无关
	exAnswers := map[uint]string{
		1: "a",
		2: "42",
		3: "wrong",
	}
	answers := make([]ExamAnswerItem, 0, len(exam.Questions))
	for _, q := range exam.Questions {
		answers = append(answers, ExamAnswerItem{QuestionID: q.ID, ExerciseID: q.ExerciseID, Answer: exAnswers[q.ExerciseID]})
	}

	report, err := svc.SubmitExam(user.ID, exam.ID, answers, 300, 0)
	if err != nil {
		t.Fatalf("SubmitExam failed: %v", err)
	}
	if report.TotalCount != 3 {
		t.Errorf("expected total 3, got %d", report.TotalCount)
	}
	if report.CorrectCount != 2 {
		t.Errorf("expected 2 correct, got %d", report.CorrectCount)
	}
	if report.Score != 66 {
		t.Errorf("expected score 66, got %d", report.Score)
	}
	if !report.Passed {
		t.Error("expected passed=true (66 >= 60)")
	}
	if report.Attempts != 1 {
		t.Errorf("expected attempts=1, got %d", report.Attempts)
	}
	if len(report.ByType) == 0 {
		t.Error("expected by_type breakdown")
	}
	if len(report.Knowledge) == 0 {
		t.Error("expected knowledge breakdown")
	}

	// 答错的题应进入错题本
	wrong, _ := repo.ListWrongExercises(user.ID, false)
	if len(wrong) != 1 {
		t.Errorf("expected 1 wrong exercise recorded, got %d", len(wrong))
	}
}

func TestSubmitExam_Failed(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewExamService(repo)

	user := seedUser(db)
	unitID := seedExamUnit(db, user.ID, true)

	exam, _ := svc.AssembleUnitExam(unitID, user.ID)

	answers := make([]ExamAnswerItem, 0, len(exam.Questions))
	for _, q := range exam.Questions {
		answers = append(answers, ExamAnswerItem{QuestionID: q.ID, ExerciseID: q.ExerciseID, Answer: "zzz"})
	}

	report, err := svc.SubmitExam(user.ID, exam.ID, answers, 300, 2)
	if err != nil {
		t.Fatalf("SubmitExam failed: %v", err)
	}
	if report.Score != 0 {
		t.Errorf("expected score 0, got %d", report.Score)
	}
	if report.Passed {
		t.Error("expected passed=false")
	}
	if report.TabSwitches != 2 {
		t.Errorf("expected tab_switches 2, got %d", report.TabSwitches)
	}
}

func TestGetReport(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewExamService(repo)

	user := seedUser(db)
	unitID := seedExamUnit(db, user.ID, true)

	exam, _ := svc.AssembleUnitExam(unitID, user.ID)
	answers := make([]ExamAnswerItem, 0, len(exam.Questions))
	for _, q := range exam.Questions {
		answers = append(answers, ExamAnswerItem{QuestionID: q.ID, ExerciseID: q.ExerciseID, Answer: "a"})
	}
	svc.SubmitExam(user.ID, exam.ID, answers, 120, 0)

	report, err := svc.GetReport(user.ID, exam.ID)
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}
	if report.Attempts != 1 {
		t.Errorf("expected attempts=1, got %d", report.Attempts)
	}
	if report.TotalCount != 3 {
		t.Errorf("expected total 3, got %d", report.TotalCount)
	}
}

// 未通过后 24h 内不可重考
func TestSubmitExam_RetryCoolDown(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewExamService(repo)

	user := seedUser(db)
	unitID := seedExamUnit(db, user.ID, true)
	exam, _ := svc.AssembleUnitExam(unitID, user.ID)

	// 全部答错 → 未通过
	bad := make([]ExamAnswerItem, 0, len(exam.Questions))
	for _, q := range exam.Questions {
		bad = append(bad, ExamAnswerItem{QuestionID: q.ID, ExerciseID: q.ExerciseID, Answer: "zzz"})
	}
	if _, err := svc.SubmitExam(user.ID, exam.ID, bad, 60, 0); err != nil {
		t.Fatalf("first submit failed: %v", err)
	}

	// 立即重考应被冷却拦截
	_, err := svc.SubmitExam(user.ID, exam.ID, bad, 60, 0)
	if !errors.Is(err, ErrRetryCoolDown) {
		t.Fatalf("expected ErrRetryCoolDown, got %v", err)
	}
}

// 已通过后不可重考
func TestSubmitExam_AlreadyPassed(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewExamService(repo)

	user := seedUser(db)
	unitID := seedExamUnit(db, user.ID, true)
	exam, _ := svc.AssembleUnitExam(unitID, user.ID)

	// 全部答对 → 通过
	good := make([]ExamAnswerItem, 0, len(exam.Questions))
	for _, q := range exam.Questions {
		answer := "a"
		if q.Type == "fillblank" {
			answer = "42"
		}
		good = append(good, ExamAnswerItem{QuestionID: q.ID, ExerciseID: q.ExerciseID, Answer: answer})
	}
	report, err := svc.SubmitExam(user.ID, exam.ID, good, 60, 0)
	if err != nil {
		t.Fatalf("first submit failed: %v", err)
	}
	if !report.Passed {
		t.Fatal("expected first submit to pass")
	}

	// 已通过后重考被拦截
	_, err = svc.SubmitExam(user.ID, exam.ID, good, 60, 0)
	if !errors.Is(err, ErrAlreadyPassed) {
		t.Fatalf("expected ErrAlreadyPassed, got %v", err)
	}
}

func TestPickExamQuestions_TypeDistribution(t *testing.T) {
	exs := []model.Exercise{
		{Type: "choice"}, {Type: "choice"}, {Type: "fillblank"},
		{Type: "choice"}, {Type: "code"}, {Type: "code"},
	}
	selected := pickExamQuestions(exs, 4)
	if len(selected) != 4 {
		t.Fatalf("expected 4 selected, got %d", len(selected))
	}
	codeCount := 0
	for _, ex := range selected {
		if ex.Type == "code" {
			codeCount++
		}
	}
	// 4 题中代码题最多 1 道（30% 向下取整）
	if codeCount > 1 {
		t.Errorf("expected at most 1 code question, got %d", codeCount)
	}
}

func TestPickExamQuestions_AllWhenFew(t *testing.T) {
	exs := []model.Exercise{{Type: "choice"}, {Type: "fillblank"}}
	selected := pickExamQuestions(exs, 10)
	if len(selected) != 2 {
		t.Errorf("expected all 2 exercises when fewer than target, got %d", len(selected))
	}
}
