package service

import (
	"testing"
	"time"

	"codelearn/model"

	"gorm.io/gorm"
)

// createExercise 创建一道测试习题
func createExercise(db *gorm.DB, lessonID uint) *model.Exercise {
	ex := &model.Exercise{
		LessonID:    lessonID,
		Type:        "choice",
		Question:    "测试题目",
		Options:     "A|B|C|D",
		Answer:      "A",
		Explanation: "解析",
		Difficulty:  "easy",
	}
	db.Create(ex)
	return ex
}

func TestSRS_TodayReviews(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewSRSReviewService(repo)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)
	ex := createExercise(db, lessonID)

	// 新错题立即进入今日复习
	if err := repo.UpsertWrongExercise(user.ID, ex.ID, "B", "exercise"); err != nil {
		t.Fatalf("upsert wrong failed: %v", err)
	}

	reviews, err := svc.TodayReviews(user.ID)
	if err != nil {
		t.Fatalf("TodayReviews failed: %v", err)
	}
	if len(reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(reviews))
	}
	if reviews[0].ExerciseID != ex.ID || reviews[0].ReviewStage != 0 {
		t.Errorf("unexpected review item: %+v", reviews[0])
	}

	// 答对 → 熟练度 +1，下次复习排到 1 天后（今日不再到期）
	if err := svc.SubmitReview(user.ID, reviews[0].ID, true); err != nil {
		t.Fatalf("SubmitReview failed: %v", err)
	}
	reviews2, _ := svc.TodayReviews(user.ID)
	if len(reviews2) != 0 {
		t.Errorf("expected no due reviews after correct, got %d", len(reviews2))
	}

	var w model.WrongExercise
	db.First(&w, reviews[0].ID)
	if w.ReviewStage != 1 {
		t.Errorf("expected stage=1, got %d", w.ReviewStage)
	}
	if w.NextReviewAt == nil || !w.NextReviewAt.After(time.Now()) {
		t.Errorf("expected next_review_at in future, got %v", w.NextReviewAt)
	}
}

func TestSRS_WrongResets(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewSRSReviewService(repo)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)
	ex := createExercise(db, lessonID)

	repo.UpsertWrongExercise(user.ID, ex.ID, "B", "exercise")
	reviews, _ := svc.TodayReviews(user.ID)
	wrongID := reviews[0].ID

	// 答错 → 重置熟练度 0，次日再复习
	if err := svc.SubmitReview(user.ID, wrongID, false); err != nil {
		t.Fatalf("SubmitReview failed: %v", err)
	}
	var w model.WrongExercise
	db.First(&w, wrongID)
	if w.ReviewStage != 0 {
		t.Errorf("expected stage reset to 0, got %d", w.ReviewStage)
	}
	if w.NextReviewAt == nil || !w.NextReviewAt.After(time.Now()) {
		t.Errorf("expected next_review_at in future after wrong, got %v", w.NextReviewAt)
	}
}

func TestSRS_MasteredAfterMaxStage(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewSRSReviewService(repo)

	_, _, lessonID := seedCourse(db)
	user := seedUser(db)
	ex := createExercise(db, lessonID)

	repo.UpsertWrongExercise(user.ID, ex.ID, "B", "exercise")
	reviews, _ := svc.TodayReviews(user.ID)
	wrongID := reviews[0].ID

	// 连续答对到最高级（len(srsIntervals)=5 次）
	for i := 0; i < len(srsIntervals); i++ {
		if err := svc.SubmitReview(user.ID, wrongID, true); err != nil {
			t.Fatalf("SubmitReview #%d failed: %v", i, err)
		}
	}
	var w model.WrongExercise
	db.First(&w, wrongID)
	if !w.Mastered {
		t.Error("expected wrong exercise auto-mastered after max stage")
	}
	if w.ReviewStage != len(srsIntervals) {
		t.Errorf("expected stage=%d, got %d", len(srsIntervals), w.ReviewStage)
	}
	// 已掌握不再出现在今日复习
	reviews2, _ := svc.TodayReviews(user.ID)
	if len(reviews2) != 0 {
		t.Errorf("expected no reviews after mastered, got %d", len(reviews2))
	}
}

func TestSRS_NoData(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewSRSReviewService(repo)

	user := seedUser(db)
	reviews, err := svc.TodayReviews(user.ID)
	if err != nil {
		t.Fatalf("TodayReviews failed: %v", err)
	}
	if len(reviews) != 0 {
		t.Errorf("expected 0 reviews, got %d", len(reviews))
	}
}
