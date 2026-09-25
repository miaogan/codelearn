package service

import (
	"errors"
	"strings"
	"time"

	"codelearn/model"
	"codelearn/repository"
)

const (
	EventRegister       = "register"
	EventLessonComplete = "lesson_complete"
	EventUnitExamPassed = "unit_exam_passed"
	EventCertIssued     = "cert_issued"
)

var ErrFeedbackEmpty = errors.New("反馈内容不能为空")

// AnalyticsService 埋点与反馈服务
type AnalyticsService struct {
	repo *repository.Repository
}

func NewAnalyticsService(repo *repository.Repository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

// Record 记录一条行为埋点
func (s *AnalyticsService) Record(eventType string, userID, courseID uint) {
	if eventType == "" || userID == 0 {
		return
	}
	_ = s.repo.CreateAnalyticsEvent(&model.AnalyticsEvent{
		UserID:    userID,
		EventType: eventType,
		CourseID:  courseID,
		CreatedAt: time.Now(),
	})
}

// FunnelRow 漏斗输出
type FunnelRow struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Users     int64  `json:"users"`
}

// Funnel 学习漏斗：注册 → 完成首课 → 通过单元考试 → 获得认证（去重用户数）
func (s *AnalyticsService) Funnel(days int) []FunnelRow {
	if days <= 0 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	rows, err := s.repo.AnalyticsFunnel(since)
	if err != nil {
		return []FunnelRow{}
	}

	counts := map[string]int64{}
	for _, r := range rows {
		counts[r.EventType] = r.Users
	}

	order := []struct {
		key   string
		label string
	}{
		{EventRegister, "注册"},
		{EventLessonComplete, "完成首课"},
		{EventUnitExamPassed, "通过单元考试"},
		{EventCertIssued, "获得认证证书"},
	}
	out := make([]FunnelRow, 0, len(order))
	for _, o := range order {
		out = append(out, FunnelRow{Key: o.key, Label: o.label, Users: counts[o.key]})
	}
	return out
}

// SubmitFeedback 提交用户反馈
func (s *AnalyticsService) SubmitFeedback(userID uint, category, content, contact string) error {
	if strings.TrimSpace(content) == "" {
		return ErrFeedbackEmpty
	}
	return s.repo.CreateFeedback(&model.UserFeedback{
		UserID:    userID,
		Category:  category,
		Content:   content,
		Contact:   contact,
		Status:    "open",
		CreatedAt: time.Now(),
	})
}
