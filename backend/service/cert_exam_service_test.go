package service

import (
	"encoding/json"
	"errors"
	"testing"

	"codelearn/model"
	"codelearn/repository"

	"gorm.io/gorm"
)

// seedCertCourse 创建认证测试课程：2 个单元，每单元 1 课时 2 习题
func seedCertCourse(db *gorm.DB) (courseID uint, unitIDs []uint) {
	course := model.Course{Language: "go", Title: "Go 认证课程", Emoji: "🐹", Color: "#00ADD8", Order: 0}
	db.Create(&course)

	for i := 0; i < 2; i++ {
		unit := model.Unit{CourseID: course.ID, Title: "Unit", Icon: "📌", Color: "#4CAF50", Order: i, ExamDurationMin: 15, ExamQuestionCount: 2}
		db.Create(&unit)
		unitIDs = append(unitIDs, unit.ID)

		lesson := model.Lesson{UnitID: unit.ID, Title: "Lesson", Icon: "1", Order: 0}
		db.Create(&lesson)

		opts, _ := json.Marshal([]string{"a", "b"})
		exs := []model.Exercise{
			{LessonID: lesson.ID, Type: "choice", Question: "选择 a", Options: string(opts), Answer: "a", Explanation: "x", Difficulty: "easy", Order: 1},
			{LessonID: lesson.ID, Type: "fillblank", Question: "填空 42", Answer: "42", Explanation: "y", Difficulty: "easy", Order: 2},
		}
		db.Create(&exs)
	}
	return course.ID, unitIDs
}

// passUnitExam 完成课时并全对通过一次单元考试
func passUnitExam(t *testing.T, repo *repository.Repository, svc *ExamService, userID, unitID uint) {
	t.Helper()
	lessons, _ := repo.ListLessonsByUnit(unitID)
	for _, l := range lessons {
		if err := repo.UpsertProgress(&model.UserProgress{UserID: userID, LessonID: l.ID, Completed: true, Score: 100}); err != nil {
			t.Fatalf("UpsertProgress failed: %v", err)
		}
	}
	exam, err := svc.AssembleUnitExam(unitID, userID)
	if err != nil {
		t.Fatalf("AssembleUnitExam failed: %v", err)
	}
	answers := make([]ExamAnswerItem, 0, len(exam.Questions))
	for _, q := range exam.Questions {
		ans := "a"
		if q.Type == "fillblank" {
			ans = "42"
		}
		answers = append(answers, ExamAnswerItem{QuestionID: q.ID, ExerciseID: q.ExerciseID, Answer: ans})
	}
	report, err := svc.SubmitExam(userID, exam.ID, answers, 60, 0)
	if err != nil {
		t.Fatalf("SubmitExam failed: %v", err)
	}
	if !report.Passed {
		t.Fatal("expected unit exam to pass")
	}
}

func TestAssembleCertExam_NotEligible(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewExamService(repo)

	user := seedUser(db)
	courseID, _ := seedCertCourse(db)

	_, err := svc.AssembleCertExam(courseID, user.ID)
	if !errors.Is(err, ErrCertUnitsNotPassed) {
		t.Fatalf("expected ErrCertUnitsNotPassed, got %v", err)
	}
}

func TestAssembleCertExam_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewExamService(repo)

	user := seedUser(db)
	courseID, _ := seedCertCourse(db)

	units, _ := repo.ListUnitsByCourse(courseID)
	for _, u := range units {
		passUnitExam(t, repo, svc, user.ID, u.ID)
	}

	exam, err := svc.AssembleCertExam(courseID, user.ID)
	if err != nil {
		t.Fatalf("AssembleCertExam failed: %v", err)
	}
	if exam.ExamType != ExamTypeCert {
		t.Errorf("expected exam type cert, got %s", exam.ExamType)
	}
	if exam.DurationMin != CertExamDurationMin {
		t.Errorf("expected duration %d, got %d", CertExamDurationMin, exam.DurationMin)
	}
	if exam.PassScore != 60 {
		t.Errorf("expected pass score 60, got %d", exam.PassScore)
	}
	// 覆盖全部单元
	covered := map[uint]bool{}
	for _, q := range exam.Questions {
		ex, err := repo.GetExercise(q.ExerciseID)
		if err != nil {
			t.Fatalf("GetExercise failed: %v", err)
		}
		l, err := repo.GetLesson(ex.LessonID)
		if err != nil {
			t.Fatalf("GetLesson failed: %v", err)
		}
		covered[l.UnitID] = true
	}
	if len(covered) != 2 {
		t.Errorf("expected questions covering 2 units, got %d", len(covered))
	}
}

func TestCertStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewExamService(repo)

	user := seedUser(db)
	courseID, _ := seedCertCourse(db)

	// 未通过任何单元考试 → 不具备资格
	st, err := svc.CertStatus(courseID, user.ID)
	if err != nil {
		t.Fatalf("CertStatus failed: %v", err)
	}
	if st.Eligible {
		t.Error("expected eligible=false before passing unit exams")
	}
	if st.Certified {
		t.Error("expected certified=false before issuance")
	}
	if st.UnitsPassed != 0 {
		t.Errorf("expected units_passed 0, got %d", st.UnitsPassed)
	}

	// 通过全部单元考试 → 具备资格
	units, _ := repo.ListUnitsByCourse(courseID)
	for _, u := range units {
		passUnitExam(t, repo, svc, user.ID, u.ID)
	}
	st, _ = svc.CertStatus(courseID, user.ID)
	if !st.Eligible {
		t.Error("expected eligible=true after passing all unit exams")
	}
	if st.UnitsPassed != 2 {
		t.Errorf("expected units_passed 2, got %d", st.UnitsPassed)
	}
}

func TestPickCertQuestions_CoverageAndCodeCap(t *testing.T) {
	pools := make([][]model.Exercise, 0, 3)
	for i := 0; i < 3; i++ {
		pool := []model.Exercise{
			{ID: uint(i*4 + 1), Type: "choice"},
			{ID: uint(i*4 + 2), Type: "fillblank"},
			{ID: uint(i*4 + 3), Type: "choice"},
			{ID: uint(i*4 + 4), Type: "code"},
		}
		pools = append(pools, pool)
	}

	selected := pickCertQuestions(pools, 10)
	if len(selected) != 10 {
		t.Fatalf("expected 10 selected, got %d", len(selected))
	}

	codeCount := 0
	covered := map[uint]bool{}
	for _, ex := range selected {
		if ex.Type == "code" {
			codeCount++
		}
		covered[(ex.ID-1)/4] = true
	}
	if codeCount > 3 {
		t.Errorf("expected code <= 30%% (3), got %d", codeCount)
	}
	for i := 0; i < 3; i++ {
		if !covered[uint(i)] {
			t.Errorf("unit %d not covered by cert questions", i)
		}
	}
}
