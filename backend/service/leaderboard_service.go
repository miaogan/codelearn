package service

import (
	"time"

	"codelearn/repository"
)

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	XP       int    `json:"xp"`
}

type LeaderboardService struct {
	repo *repository.Repository
}

func NewLeaderboardService(repo *repository.Repository) *LeaderboardService {
	return &LeaderboardService{repo: repo}
}

// WeeklyLeaderboard 本周排行榜（自然周，周一为起点）
func (s *LeaderboardService) WeeklyLeaderboard(limit int) ([]LeaderboardEntry, error) {
	rows, err := s.repo.WeeklyLeaderboard(weekStart(), limit)
	if err != nil {
		return nil, err
	}
	entries := make([]LeaderboardEntry, 0, len(rows))
	for i, r := range rows {
		entries = append(entries, LeaderboardEntry{Rank: i + 1, UserID: r.UserID, Username: r.Username, XP: r.XP})
	}
	return entries, nil
}

// MyRank 当前用户本周排名与 XP
func (s *LeaderboardService) MyRank(userID uint) (*LeaderboardEntry, error) {
	since := weekStart()
	xp, err := s.repo.SumXPBetween(userID, since, time.Now().Add(time.Hour))
	if err != nil {
		return nil, err
	}

	// 用 Top 1000 估算排在自己前面的用户数
	rows, err := s.repo.WeeklyLeaderboard(since, 1000)
	if err != nil {
		return nil, err
	}
	greater := 0
	for _, r := range rows {
		if r.XP > xp {
			greater++
		}
	}
	return &LeaderboardEntry{Rank: greater + 1, UserID: userID, Username: "", XP: xp}, nil
}

// Calendar 学习日历（某月每天的 XP）
func (s *LeaderboardService) Calendar(userID uint, year, month int) ([]repository.CalendarDay, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	return s.repo.CalendarDays(userID, start, end)
}

// weekStart 返回本周一零点
func weekStart() time.Time {
	now := time.Now()
	wd := int(now.Weekday())
	if wd == 0 {
		wd = 7
	}
	daysSinceMonday := wd - 1
	return time.Date(now.Year(), now.Month(), now.Day()-daysSinceMonday, 0, 0, 0, 0, now.Location())
}
