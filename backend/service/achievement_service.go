package service

import (
	"time"

	"codelearn/model"
	"codelearn/repository"
)

// AchievementService 成就徽章与段位称号（U12）
type AchievementService struct {
	repo *repository.Repository
}

func NewAchievementService(repo *repository.Repository) *AchievementService {
	return &AchievementService{repo: repo}
}

type badgeDef struct {
	code        string
	title       string
	icon        string
	description string
}

var badgeDefs = map[string]badgeDef{
	"first_lesson":    {code: "first_lesson", title: "初出茅庐", icon: "🎓", description: "完成第一个课时"},
	"streak_7":        {code: "streak_7", title: "七日之约", icon: "🔥", description: "连续打卡 7 天"},
	"streak_30":       {code: "streak_30", title: "月度战士", icon: "👑", description: "连续打卡 30 天"},
	"first_exam_pass": {code: "first_exam_pass", title: "首战告捷", icon: "🏅", description: "首次通过单元考试"},
	"perfect_exam":    {code: "perfect_exam", title: "满分学霸", icon: "💯", description: "单场考试满分"},
	"first_cert":      {code: "first_cert", title: "能力认证", icon: "🎖️", description: "首次获得能力认证证书"},
	"practice_50":     {code: "practice_50", title: "勤学苦练", icon: "✏️", description: "累计练习答对 50 题"},
	"practice_200":    {code: "practice_200", title: "千锤百炼", icon: "🚀", description: "累计练习答对 200 题"},
}

// titleLevels 段位称号（按 XP 从低到高）
var titleLevels = []struct {
	minXP int
	name  string
	icon  string
}{
	{0, "代码萌芽", "🌱"},
	{100, "代码学徒", "🧑‍🎓"},
	{300, "代码骑士", "⚔️"},
	{600, "代码剑士", "🗡️"},
	{1000, "代码大师", "🧙"},
	{2000, "代码宗师", "🏆"},
}

// TitleForXP 根据 XP 返回段位称号
func TitleForXP(xp int) (name, icon string, level int) {
	name = titleLevels[0].name
	icon = titleLevels[0].icon
	for i, tl := range titleLevels {
		if xp >= tl.minXP {
			level = i
			name = tl.name
			icon = tl.icon
		}
	}
	return name, icon, level
}

// Unlock 解锁徽章（幂等：已解锁则跳过），返回是否新解锁
func (s *AchievementService) Unlock(userID uint, code string) bool {
	def, ok := badgeDefs[code]
	if !ok {
		return false
	}
	if _, err := s.repo.GetAchievement(userID, code); err == nil {
		return false
	}
	return s.repo.CreateAchievement(&model.Achievement{
		UserID:      userID,
		Code:        def.code,
		Title:       def.title,
		Icon:        def.icon,
		Description: def.description,
		UnlockedAt:  time.Now(),
	}) == nil
}

// Evaluate 全量评估用户当前应得徽章（幂等）
func (s *AchievementService) Evaluate(userID uint) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return
	}
	if user.StreakDays >= 7 {
		s.Unlock(userID, "streak_7")
	}
	if user.StreakDays >= 30 {
		s.Unlock(userID, "streak_30")
	}
	if n, err := s.repo.CountXPEventsByReason(userID, XPReasonExercise); err == nil {
		if n >= 50 {
			s.Unlock(userID, "practice_50")
		}
		if n >= 200 {
			s.Unlock(userID, "practice_200")
		}
	}
}

// OnLessonCompleted 完成课时事件（首课 + streak + 练习数）
func (s *AchievementService) OnLessonCompleted(userID uint) {
	s.Unlock(userID, "first_lesson")
	s.Evaluate(userID)
}

// OnExamPassed 考试通过事件（首次通过 / 满分）
func (s *AchievementService) OnExamPassed(userID uint, score int) {
	s.Unlock(userID, "first_exam_pass")
	if score >= 100 {
		s.Unlock(userID, "perfect_exam")
	}
	s.Evaluate(userID)
}

// OnCertIssued 发证事件
func (s *AchievementService) OnCertIssued(userID uint) {
	s.Unlock(userID, "first_cert")
	s.Evaluate(userID)
}

// OnPracticeCorrect 练习答对事件
func (s *AchievementService) OnPracticeCorrect(userID uint) {
	s.Evaluate(userID)
}

// AchievementDTO 徽章输出
type AchievementDTO struct {
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	UnlockedAt  time.Time `json:"unlocked_at"`
}

// TitleDTO 段位输出
type TitleDTO struct {
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	Level      int    `json:"level"`
	CurrentXP  int    `json:"current_xp"`
	XPToNext   int    `json:"xp_to_next"` // -1 表示已达最高段位
	NextName   string `json:"next_name,omitempty"`
}

// AchievementSummary 成就墙汇总
type AchievementSummary struct {
	Badges []AchievementDTO `json:"badges"`
	Total  int              `json:"total"`
	Title  TitleDTO         `json:"title"`
}

// Summary 查询用户成就墙与段位
func (s *AchievementService) Summary(userID uint) (*AchievementSummary, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	badges, _ := s.repo.ListAchievements(userID)

	name, icon, level := TitleForXP(user.XP)
	nextName := ""
	xpToNext := -1
	if level+1 < len(titleLevels) {
		nextName = titleLevels[level+1].name
		xpToNext = titleLevels[level+1].minXP - user.XP
		if xpToNext < 0 {
			xpToNext = 0
		}
	}

	dto := make([]AchievementDTO, 0, len(badges))
	for _, b := range badges {
		dto = append(dto, AchievementDTO{
			Code:        b.Code,
			Title:       b.Title,
			Icon:        b.Icon,
			Description: b.Description,
			UnlockedAt:  b.UnlockedAt,
		})
	}
	return &AchievementSummary{
		Badges: dto,
		Total:  len(dto),
		Title: TitleDTO{
			Name:      name,
			Icon:      icon,
			Level:     level,
			CurrentXP: user.XP,
			XPToNext:  xpToNext,
			NextName:  nextName,
		},
	}, nil
}
