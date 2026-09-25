package service

import (
	"time"

	"codelearn/model"
	"codelearn/repository"
)

// PortfolioService 数字能力档案（U10）：聚合证书/项目/课程掌握度生成公开能力页
type PortfolioService struct {
	repo *repository.Repository
}

func NewPortfolioService(repo *repository.Repository) *PortfolioService {
	return &PortfolioService{repo: repo}
}

// PortfolioItem 单门课程能力条目
type PortfolioCourse struct {
	CourseID     uint   `json:"course_id"`
	CourseTitle  string `json:"course_title"`
	Language     string `json:"language"`
	Emoji        string `json:"emoji"`
	UnitsTotal   int    `json:"units_total"`
	UnitsPassed  int    `json:"units_passed"`
	Mastery      int    `json:"mastery"` // 0-100
	Certified    bool   `json:"certified"`
	CertLevel    string `json:"cert_level,omitempty"`
	CertScore    int    `json:"cert_score,omitempty"`
	CertNo       string `json:"cert_no,omitempty"`
	ProjectCount int    `json:"project_count"`
}

// PortfolioProject 公开档案中的项目作品（仅暴露公开字段）
type PortfolioProject struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CourseTitle string     `json:"course_title"`
	Language    string     `json:"language"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	RunCount    int        `json:"run_count"`
}

// Portfolio 公开能力档案
type Portfolio struct {
	Username     string              `json:"username"`
	XP           int                 `json:"xp"`
	StreakDays   int                 `json:"streak_days"`
	TitleName    string              `json:"title_name"`
	TitleIcon    string              `json:"title_icon"`
	TitleLevel   int                 `json:"title_level"`
	BadgeCount   int                 `json:"badge_count"`
	Courses      []PortfolioCourse   `json:"courses"`
	Projects     []PortfolioProject  `json:"projects"`
}

// Get 生成用户公开档案（不暴露隐私信息）
func (s *PortfolioService) Get(username string) (*Portfolio, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	courses, _ := s.repo.ListCourses()
	badges, _ := s.repo.ListAchievements(user.ID)
	projects, _ := s.repo.ListProjectsByUser(user.ID)
	name, icon, level := TitleForXP(user.XP)

	pf := &Portfolio{
		Username:   user.Username,
		XP:         user.XP,
		StreakDays: user.StreakDays,
		TitleName:  name,
		TitleIcon:  icon,
		TitleLevel: level,
		BadgeCount: len(badges),
		Courses:    make([]PortfolioCourse, 0, len(courses)),
		Projects:   trimProjects(projects),
	}

	for _, c := range courses {
		item := PortfolioCourse{
			CourseID:    c.ID,
			CourseTitle: c.Title,
			Language:    c.Language,
			Emoji:       c.Emoji,
		}
		units, _ := s.repo.ListUnitsByCourse(c.ID)
		item.UnitsTotal = len(units)
		passed := 0
		examTotal := 0
		examCorrect := 0
		for _, u := range units {
			examTotal++
			if ok := s.courseExamPassed(user.ID, u.ID); ok {
				passed++
			}
			if acc, err := s.unitExamAccuracy(user.ID, u.ID); err == nil {
				examCorrect += acc
			}
		}
		item.UnitsPassed = passed
		// 掌握度：单元通过率 70% + 考试正确率 30%（无考试记录按 0）
		mastery := 0
		if examTotal > 0 {
			unitRate := passed * 100 / examTotal
			avgAcc := 0
			if examTotal > 0 {
				avgAcc = examCorrect / examTotal
			}
			mastery = unitRate*70/100 + avgAcc*30/100
		}
		item.Mastery = mastery
		if cert, err := s.repo.GetCertificate(user.ID, c.ID); err == nil {
			item.Certified = true
			item.CertLevel = cert.Level
			item.CertScore = cert.Score
			item.CertNo = cert.CertNo
		}
		for _, p := range projects {
			if p.CourseID == c.ID && p.Status == "completed" {
				item.ProjectCount++
			}
		}
		pf.Courses = append(pf.Courses, item)
	}
	return pf, nil
}

// trimProjects 仅保留项目作品的公开字段
func trimProjects(projects []model.Project) []PortfolioProject {
	out := make([]PortfolioProject, 0, len(projects))
	for _, p := range projects {
		out = append(out, PortfolioProject{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			CourseTitle: p.CourseTitle,
			Language:    p.Language,
			Status:      p.Status,
			CompletedAt: p.CompletedAt,
			RunCount:    p.RunCount,
		})
	}
	return out
}

// courseExamPassed 判断某用户是否通过某单元考试
func (s *PortfolioService) courseExamPassed(userID, unitID uint) bool {
	exams, err := s.repo.ListExamsByUnit(unitID)
	if err != nil {
		return false
	}
	for _, e := range exams {
		if subs, err := s.repo.ListExamSubmissions(userID, e.ID); err == nil {
			for _, sub := range subs {
				if sub.Passed {
					return true
				}
			}
		}
	}
	return false
}

// unitExamAccuracy 单元考试平均正确率（最近一次提交）
func (s *PortfolioService) unitExamAccuracy(userID, unitID uint) (int, error) {
	exams, err := s.repo.ListExamsByUnit(unitID)
	if err != nil || len(exams) == 0 {
		return 0, err
	}
	last, err := s.repo.GetLatestExamSubmission(userID, exams[0].ID)
	if err != nil {
		return 0, err
	}
	if last.TotalCount == 0 {
		return 0, nil
	}
	return last.CorrectCount * 100 / last.TotalCount, nil
}
