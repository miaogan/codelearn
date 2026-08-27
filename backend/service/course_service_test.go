package service

import (
	"testing"

	"codelearn/model"
	"codelearn/repository"

	"gorm.io/gorm"
)

// newTestRepo 创建测试用 repository
func newTestRepo(db *gorm.DB) *repository.Repository {
	return repository.New(db)
}

func TestCourseService_ListCourses(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCourseService(repo)

	seedCourse(db)

	courses, err := svc.ListCourses()
	if err != nil {
		t.Fatalf("ListCourses failed: %v", err)
	}
	if len(courses) != 1 {
		t.Errorf("expected 1 course, got %d", len(courses))
	}
	if courses[0].Title != "Go 基础" {
		t.Errorf("expected title 'Go 基础', got '%s'", courses[0].Title)
	}
}

func TestCourseService_GetLearningPath(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCourseService(repo)

	courseID, _, lessonID := seedCourse(db)
	user := seedUser(db)

	path, err := svc.GetLearningPath(courseID, user.ID)
	if err != nil {
		t.Fatalf("GetLearningPath failed: %v", err)
	}

	if path.Course.ID != courseID {
		t.Errorf("expected course ID=%d, got %d", courseID, path.Course.ID)
	}
	if len(path.Units) != 1 {
		t.Fatalf("expected 1 unit, got %d", len(path.Units))
	}
	if len(path.Units[0].Lessons) != 1 {
		t.Fatalf("expected 1 lesson, got %d", len(path.Units[0].Lessons))
	}

	lesson := path.Units[0].Lessons[0]
	if lesson.ID != lessonID {
		t.Errorf("expected lesson ID=%d, got %d", lessonID, lesson.ID)
	}
	if !lesson.Unlocked {
		t.Error("expected first lesson to be unlocked")
	}
	if lesson.Completed {
		t.Error("expected lesson to be not completed")
	}
}

func TestCourseService_LearningPath_UnlockChain(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCourseService(repo)

	courseID, unitID, _ := seedCourse(db)

	// seedCourse 已创建 1 课，再添加 2 课共 3 课
	lesson2 := model.Lesson{UnitID: unitID, Title: "L2", Icon: "2️⃣", Order: 1}
	lesson3 := model.Lesson{UnitID: unitID, Title: "L3", Icon: "3️⃣", Order: 2}
	db.Create(&lesson2)
	db.Create(&lesson3)

	user := seedUser(db)

	// 初始：所有课程解锁（因为前一课未完成时 prevCompleted 初始为 true）
	path, err := svc.GetLearningPath(courseID, user.ID)
	if err != nil {
		t.Fatalf("GetLearningPath failed: %v", err)
	}

	lessons := path.Units[0].Lessons
	if len(lessons) != 3 {
		t.Fatalf("expected 3 lessons, got %d", len(lessons))
	}

	// 第一课完成后，第二课应解锁，第三课也解锁
	repo.UpsertProgress(&model.UserProgress{
		UserID: user.ID, LessonID: lessons[0].ID, Completed: true,
	})

	path2, _ := svc.GetLearningPath(courseID, user.ID)
	lessons2 := path2.Units[0].Lessons

	if !lessons2[0].Completed {
		t.Error("lesson 0 should be completed")
	}
	if !lessons2[1].Unlocked {
		t.Error("lesson 1 should be unlocked")
	}
}

func TestCourseService_GetExercises_HideAnswer(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCourseService(repo)

	_, _, lessonID := seedCourse(db)

	ex1 := model.Exercise{
		LessonID:    lessonID,
		Type:        "choice",
		Question:    "What is 1+1?",
		Options:     `["1","2","3"]`,
		Answer:      "2",
		Explanation:  "Basic math",
		TestCases:   `[{"input":"1","expected":"2"}]`,
		Order:       0,
	}
	db.Create(&ex1)

	exercises, err := svc.GetExercises(lessonID)
	if err != nil {
		t.Fatalf("GetExercises failed: %v", err)
	}
	if len(exercises) != 1 {
		t.Fatalf("expected 1 exercise, got %d", len(exercises))
	}

	if exercises[0].Answer != "" {
		t.Error("answer should be hidden from frontend")
	}
	if exercises[0].TestCases != "" {
		t.Error("test cases should be hidden from frontend")
	}
	if exercises[0].Question != "What is 1+1?" {
		t.Errorf("expected question, got '%s'", exercises[0].Question)
	}
}

func TestCourseService_SubmitAnswer_Correct(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCourseService(repo)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)

	ex := model.Exercise{
		LessonID:   lessonID,
		Type:       "choice",
		Question:   "Go is a ___ language",
		Answer:     "compiled",
		Explanation: "Go is compiled, not interpreted",
		Order:      0,
	}
	db.Create(&ex)

	correct, explanation, err := svc.SubmitAnswer(user.ID, ex.ID, "compiled")
	if err != nil {
		t.Fatalf("SubmitAnswer failed: %v", err)
	}
	if !correct {
		t.Error("expected correct=true")
	}
	if explanation != "Go is compiled, not interpreted" {
		t.Errorf("unexpected explanation: %s", explanation)
	}
}

func TestCourseService_SubmitAnswer_Incorrect(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCourseService(repo)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)

	ex := model.Exercise{
		LessonID:   lessonID,
		Type:       "choice",
		Question:   "Go is a ___ language",
		Answer:     "compiled",
		Explanation: "Go is compiled",
		Order:      0,
	}
	db.Create(&ex)

	correct, _, err := svc.SubmitAnswer(user.ID, ex.ID, "interpreted")
	if err != nil {
		t.Fatalf("SubmitAnswer failed: %v", err)
	}
	if correct {
		t.Error("expected correct=false")
	}
}

func TestCourseService_SubmitAnswer_CaseInsensitive(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCourseService(repo)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)

	ex := model.Exercise{
		LessonID: lessonID,
		Type:     "fillblank",
		Answer:   "Hello",
		Order:    0,
	}
	db.Create(&ex)

	correct, _, err := svc.SubmitAnswer(user.ID, ex.ID, "  hello  ")
	if err != nil {
		t.Fatalf("SubmitAnswer failed: %v", err)
	}
	if !correct {
		t.Error("expected correct=true (case insensitive + trim)")
	}
}

func TestCourseService_SubmitAnswer_CodeType(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCourseService(repo)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)

	ex := model.Exercise{
		LessonID: lessonID,
		Type:     "code",
		Answer:   "some code",
		Order:    0,
	}
	db.Create(&ex)

	correct, _, err := svc.SubmitAnswer(user.ID, ex.ID, "some code")
	if err != nil {
		t.Fatalf("SubmitAnswer failed: %v", err)
	}
	if correct {
		t.Error("expected correct=false for code type (judged by sandbox)")
	}
}

func TestCheckAnswer(t *testing.T) {
	tests := []struct {
		name     string
		exType   string
		answer   string
		input    string
		expected bool
	}{
		{"choice correct", "choice", "B", "b", true},
		{"choice wrong", "choice", "B", "C", false},
		{"fillblank correct", "fillblank", "variable", " Variable ", true},
		{"fillblank wrong", "fillblank", "variable", "constant", false},
		{"code always false", "code", "answer", "answer", false},
		{"order correct", "order", "ABC", "abc", true},
		{"unknown type", "unknown", "X", "x", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ex := &model.Exercise{Type: tt.exType, Answer: tt.answer}
			got := checkAnswer(ex, tt.input)
			if got != tt.expected {
				t.Errorf("checkAnswer(%s, %q) = %v, want %v", tt.exType, tt.input, got, tt.expected)
			}
		})
	}
}
