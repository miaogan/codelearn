package service

import (
	"time"

	"codelearn/model"
	"codelearn/repository"
)

type WrongExerciseService struct {
	repo *repository.Repository
}

func NewWrongExerciseService(repo *repository.Repository) *WrongExerciseService {
	return &WrongExerciseService{repo: repo}
}

// WrongExerciseItem 错题列表项（含习题详情）
type WrongExerciseItem struct {
	ID           uint      `json:"id"`
	ExerciseID   uint      `json:"exercise_id"`
	Type         string    `json:"type"`
	Question     string    `json:"question"`
	Options      string    `json:"options"`
	Difficulty   string    `json:"difficulty"`
	CodeTemplate string    `json:"code_template"`
	UserAnswer   string    `json:"user_answer"`
	CorrectAnswer string   `json:"correct_answer"`
	Explanation  string    `json:"explanation"`
	WrongCount   int       `json:"wrong_count"`
	Mastered     bool      `json:"mastered"`
	LastWrongAt  time.Time `json:"last_wrong_at"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
}

func (s *WrongExerciseService) ListWrongExercises(userID uint, onlyUnmastered bool) ([]WrongExerciseItem, error) {
	wrongs, err := s.repo.ListWrongExercises(userID, onlyUnmastered)
	if err != nil {
		return nil, err
	}

	exerciseIDs := make([]uint, 0, len(wrongs))
	for _, w := range wrongs {
		exerciseIDs = append(exerciseIDs, w.ExerciseID)
	}

	exercises, err := s.repo.ListExercisesByIDs(exerciseIDs)
	if err != nil {
		return nil, err
	}

	exMap := make(map[uint]model.Exercise, len(exercises))
	for _, e := range exercises {
		exMap[e.ID] = e
	}

	items := make([]WrongExerciseItem, 0, len(wrongs))
	for _, w := range wrongs {
		ex, ok := exMap[w.ExerciseID]
		if !ok {
			continue
		}
		items = append(items, WrongExerciseItem{
			ID:           w.ID,
			ExerciseID:   w.ExerciseID,
			Type:         ex.Type,
			Question:     ex.Question,
			Options:      ex.Options,
			Difficulty:   ex.Difficulty,
			CodeTemplate: ex.CodeTemplate,
			UserAnswer:   w.UserAnswer,
			CorrectAnswer: ex.Answer,
			Explanation:  ex.Explanation,
			WrongCount:   w.WrongCount,
			Mastered:     w.Mastered,
			LastWrongAt:  w.LastWrongAt,
			ReviewedAt:   w.ReviewedAt,
		})
	}

	return items, nil
}

func (s *WrongExerciseService) MarkMastered(userID, exerciseID uint) error {
	return s.repo.MarkWrongExerciseMastered(userID, exerciseID)
}

func (s *WrongExerciseService) CountWrong(userID uint) (int64, error) {
	return s.repo.CountWrongExercises(userID)
}
