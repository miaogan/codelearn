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
	Username   string             `json:"username"`
	XP         int                `json:"xp"`
	StreakDays int                `json:"streak_days"`
	TitleName  string             `json:"title_name"`
	TitleIcon  string             `json:"title_icon"`
	TitleLevel int                `json:"title_level"`
	BadgeCount int                `json:"badge_count"`
	Courses    []PortfolioCourse  `json:"courses"`
	Projects   []PortfolioProject `json:"projects"`
}

// Get 生成用户公开档案（不暴露隐私信息）。一次性批量加载单元/考试/提交/证书，避免按课程逐次查询（N+1）。
func (s *PortfolioService) Get(username string) (*Portfolio, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	courses, _ := s.repo.ListCourses()
	badges, _ := s.repo.ListAchievements(user.ID)
	projects, _ := s.repo.ListProjectsByUser(user.ID)
	certs, _ := s.repo.ListCertificatesByUser(user.ID)
	units, _ := s.repo.ListAllUnits()
	allSubs, _ := s.repo.ListExamSubmissionsByUser(user.ID)

	// 组装内存索引
	unitsByCourse := map[uint][]model.Unit{}
	allUnitIDs := make([]uint, 0, len(units))
	for _, u := range units {
		unitsByCourse[u.CourseID] = append(unitsByCourse[u.CourseID], u)
		allUnitIDs = append(allUnitIDs, u.ID)
	}
	exams, _ := s.repo.ListExamsByUnitIDs(allUnitIDs)
	examsByUnit := map[uint][]model.Exam{}
	for _, e := range exams {
		examsByUnit[e.UnitID] = append(examsByUnit[e.UnitID], e)
	}
	latestByExam := map[uint]*model.ExamSubmission{}
	for i := range allSubs {
		latestByExam[allSubs[i].ExamID] = &allSubs[i] // 按时间正序，后写覆盖即为最近一次
	}
	certByCourse := map[uint]*model.Certificate{}
	for i := range certs {
		certByCourse[certs[i].CourseID] = &certs[i]
	}

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
		cunits := unitsByCourse[c.ID]
		item.UnitsTotal = len(cunits)
		passed := 0
		examTotal := 0
		examCorrect := 0
		for _, u := range cunits {
			examTotal++
			if s.courseExamPassed(examsByUnit[u.ID], latestByExam) {
				passed++
			}
			if acc, ok := s.unitExamAccuracy(examsByUnit[u.ID], latestByExam); ok {
				examCorrect += acc
			}
		}
		item.UnitsPassed = passed
		// 掌握度：单元通过率 70% + 考试正确率 30%（无考试记录按 0）
		mastery := 0
		if examTotal > 0 {
			unitRate := passed * 100 / examTotal
			avgAcc := examCorrect / examTotal
			mastery = unitRate*70/100 + avgAcc*30/100
		}
		item.Mastery = mastery
		if cert, ok := certByCourse[c.ID]; ok {
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

// courseExamPassed 判断某用户是否通过某单元考试（基于预加载索引）
func (s *PortfolioService) courseExamPassed(exams []model.Exam, latestByExam map[uint]*model.ExamSubmission) bool {
	for _, e := range exams {
		if sub, ok := latestByExam[e.ID]; ok && sub.Passed {
			return true
		}
	}
	return false
}

// unitExamAccuracy 单元考试平均正确率（该单元各考试实例中最近一次提交，基于预加载索引）
func (s *PortfolioService) unitExamAccuracy(exams []model.Exam, latestByExam map[uint]*model.ExamSubmission) (int, bool) {
	var last *model.ExamSubmission
	for _, e := range exams {
		if sub, ok := latestByExam[e.ID]; ok {
			if last == nil || sub.CreatedAt.After(last.CreatedAt) {
				last = sub
			}
		}
	}
	if last == nil || last.TotalCount == 0 {
		return 0, false
	}
	return last.CorrectCount * 100 / last.TotalCount, true
}
