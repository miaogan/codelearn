package service

import (
	"time"

	"codelearn/model"
	"codelearn/repository"
)

// SRSReviewService 间隔重复复习（U3）：错题按遗忘曲线排期，替代全量错题列表
type SRSReviewService struct {
	repo *repository.Repository
}

func NewSRSReviewService(repo *repository.Repository) *SRSReviewService {
	return &SRSReviewService{repo: repo}
}

// SRSReviewItem 今日待复习题
type SRSReviewItem struct {
	ID            uint   `json:"id"`
	ExerciseID    uint   `json:"exercise_id"`
	Type          string `json:"type"`
	Question      string `json:"question"`
	Options       string `json:"options"`
	CorrectAnswer string `json:"correct_answer"`
	Explanation   string `json:"explanation"`
	ReviewStage   int    `json:"review_stage"`
}

// srsIntervals 各熟练度（1-5）对应的复习间隔天数
var srsIntervals = []int{1, 2, 4, 7, 15}

// TodayReviews 今日到期的复习题（未掌握 且 到期或尚未排期）
func (s *SRSReviewService) TodayReviews(userID uint) ([]SRSReviewItem, error) {
	wrongs, err := s.repo.ListDueWrongExercises(userID, time.Now())
	if err != nil {
		return nil, err
	}
	if len(wrongs) == 0 {
		return []SRSReviewItem{}, nil
	}

	exIDs := make([]uint, 0, len(wrongs))
	for _, w := range wrongs {
		exIDs = append(exIDs, w.ExerciseID)
	}
	exercises, err := s.repo.ListExercisesByIDs(exIDs)
	if err != nil {
		return nil, err
	}
	exMap := make(map[uint]model.Exercise, len(exercises))
	for _, e := range exercises {
		exMap[e.ID] = e
	}

	items := make([]SRSReviewItem, 0, len(wrongs))
	for _, w := range wrongs {
		ex, ok := exMap[w.ExerciseID]
		if !ok {
			continue
		}
		items = append(items, SRSReviewItem{
			ID:            w.ID,
			ExerciseID:    w.ExerciseID,
			Type:          ex.Type,
			Question:      ex.Question,
			Options:       ex.Options,
			CorrectAnswer: ex.Answer,
			Explanation:   ex.Explanation,
			ReviewStage:   w.ReviewStage,
		})
	}
	return items, nil
}

// SubmitReview 提交一次复习结果：答对升级熟练度并拉长间隔；答错重置。达到最高级自动标记掌握。
func (s *SRSReviewService) SubmitReview(userID, wrongID uint, correct bool) error {
	w, err := s.repo.GetWrongExerciseByID(userID, wrongID)
	if err != nil {
		return err
	}
	now := time.Now()
	if correct {
		next := w.ReviewStage + 1
		if next > len(srsIntervals) {
			next = len(srsIntervals)
		}
		w.ReviewStage = next
		days := srsIntervals[next-1]
		t := now.AddDate(0, 0, days)
		w.NextReviewAt = &t
		if w.ReviewStage >= len(srsIntervals) {
			w.Mastered = true
			w.ReviewedAt = &now
		}
	} else {
		w.ReviewStage = 0
		t := now.AddDate(0, 0, 1)
		w.NextReviewAt = &t
		w.Mastered = false
	}
	return s.repo.UpdateWrongExerciseReview(w)
}
