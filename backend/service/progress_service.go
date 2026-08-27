package service

import (
	"strings"
	"time"

	"codelearn/model"
	"codelearn/repository"
)

type ProgressService struct {
	repo       *repository.Repository
	xpLesson   int
	xpExercise int
	maxHearts  int
}

func NewProgressService(repo *repository.Repository, xpLesson, xpExercise, maxHearts int) *ProgressService {
	return &ProgressService{
		repo:       repo,
		xpLesson:   xpLesson,
		xpExercise: xpExercise,
		maxHearts:  maxHearts,
	}
}

type UserStats struct {
	XP            int   `json:"xp"`
	StreakDays    int   `json:"streak_days"`
	Hearts        int   `json:"hearts"`
	MaxHearts     int   `json:"max_hearts"`
	DailyGoal     int   `json:"daily_goal"`
	TodayXP       int   `json:"today_xp"`
	CompletedToday int  `json:"completed_today"`
}

func (s *ProgressService) GetStats(userID uint) (*UserStats, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	completed, _ := s.repo.CountCompletedLessons(userID)
	submissions, _ := s.repo.CountTodaySubmissions(userID)
	todayXP := int(submissions) * s.xpExercise

	return &UserStats{
		XP:            user.XP,
		StreakDays:    user.StreakDays,
		Hearts:        user.Hearts,
		MaxHearts:     user.MaxHearts,
		DailyGoal:     user.DailyGoal,
		TodayXP:       todayXP,
		CompletedToday: int(completed),
	}, nil
}

// CompleteLesson 标记课程完成，发放 XP 并更新连续打卡
func (s *ProgressService) CompleteLesson(userID, lessonID uint, score int) (int, error) {
	existing, err := s.repo.GetProgress(userID, lessonID)
	if err == nil && existing.Completed {
		// 已完成过，不重复发 XP
		return 0, nil
	}

	xpEarned := s.xpLesson
	progress := &model.UserProgress{
		UserID:   userID,
		LessonID: lessonID,
		Completed: true,
		Score:    score,
		XPEarned: xpEarned,
	}
	if err := s.repo.UpsertProgress(progress); err != nil {
		return 0, err
	}

	// 更新用户 XP 和连续打卡
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return xpEarned, nil
	}
	user.XP += xpEarned
	updateStreak(user)
	s.repo.UpdateUser(user)

	return xpEarned, nil
}

// LoseHeart 答错时扣心数
func (s *ProgressService) LoseHeart(userID uint) (int, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return 0, err
	}
	if user.Hearts > 0 {
		user.Hearts--
		s.repo.UpdateUser(user)
	}
	return user.Hearts, nil
}

// RestoreHeart 恢复一个心数（如完成课程后）
func (s *ProgressService) RestoreHeart(userID uint) (int, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return 0, err
	}
	if user.Hearts < user.MaxHearts {
		user.Hearts++
		s.repo.UpdateUser(user)
	}
	return user.Hearts, nil
}

func (s *ProgressService) ListProgress(userID uint) ([]model.UserProgress, error) {
	return s.repo.ListProgressByUser(userID)
}

// updateStreak 更新连续打卡天数
func updateStreak(user *model.User) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if user.LastStreakAt != nil {
		last := time.Date(
			user.LastStreakAt.Year(),
			user.LastStreakAt.Month(),
			user.LastStreakAt.Day(),
			0, 0, 0, 0, user.LastStreakAt.Location(),
		)
		diff := today.Sub(last).Hours() / 24
		if diff < 1 {
			// 今天已打卡
			return
		}
		if diff < 2 {
			// 连续打卡
			user.StreakDays++
		} else {
			// 断了
			user.StreakDays = 1
		}
	} else {
		user.StreakDays = 1
	}
	user.LastStreakAt = &today
}

func normalize(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}
