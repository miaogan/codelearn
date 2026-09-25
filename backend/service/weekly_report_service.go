package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"codelearn/repository"
)

// WeeklyReportService 学习周报（U13）：过去 7 天学习总结 + 薄弱点 + 建议（模板生成，无 LLM 依赖可降级）
type WeeklyReportService struct {
	repo *repository.Repository
}

func NewWeeklyReportService(repo *repository.Repository) *WeeklyReportService {
	return &WeeklyReportService{repo: repo}
}

// WeakTopic 薄弱知识点（按课时聚合本周错题）
type WeakTopic struct {
	LessonID   uint   `json:"lesson_id"`
	Title      string `json:"title"`
	WrongCount int    `json:"wrong_count"`
}

// WeeklyReportDTO 周报
type WeeklyReportDTO struct {
	WeekStart     string      `json:"week_start"`
	WeekEnd       string      `json:"week_end"`
	XPTotal       int         `json:"xp_total"`
	ActiveDays    int         `json:"active_days"`
	LessonsDone   int64       `json:"lessons_completed"`
	ExercisesDone int64       `json:"exercises_done"`
	ExamsPassed   int64       `json:"exams_passed"`
	WrongAdded    int64       `json:"wrong_added"`
	StreakDays    int         `json:"streak_days"`
	WeakTopics    []WeakTopic `json:"weak_topics"`
	Summary       string      `json:"summary"`
	Advice        string      `json:"advice"`
}

// Report 生成过去 7 天的学习周报
func (s *WeeklyReportService) Report(userID uint) (*WeeklyReportDTO, error) {
	now := time.Now()
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	start := end.AddDate(0, 0, -7)

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	xpTotal, _ := s.repo.SumXPBetween(userID, start, end)
	days, _ := s.repo.CalendarDays(userID, start, end)
	lessons, _ := s.repo.CountCompletedLessonsBetween(userID, start, end)
	exercises, _ := s.repo.CountSubmissionsBetween(userID, start, end)
	examsPassed, _ := s.repo.CountExamPassedBetween(userID, start, end)
	wrongAdded, _ := s.repo.CountWrongAddedBetween(userID, start, end)
	weakTopics := s.weakTopics(userID)

	report := &WeeklyReportDTO{
		WeekStart:     start.Format("01-02"),
		WeekEnd:       end.AddDate(0, 0, -1).Format("01-02"),
		XPTotal:       xpTotal,
		ActiveDays:    len(days),
		LessonsDone:   lessons,
		ExercisesDone: exercises,
		ExamsPassed:   examsPassed,
		WrongAdded:    wrongAdded,
		StreakDays:    user.StreakDays,
		WeakTopics:    weakTopics,
	}
	report.Summary = s.buildSummary(report)
	report.Advice = s.buildAdvice(report)
	return report, nil
}

// weakTopics 按课时聚合错题（TOP3）
func (s *WeeklyReportService) weakTopics(userID uint) []WeakTopic {
	wrongs, err := s.repo.ListWrongExercises(userID, false)
	if err != nil || len(wrongs) == 0 {
		return []WeakTopic{}
	}
	exIDs := make([]uint, 0, len(wrongs))
	for _, w := range wrongs {
		exIDs = append(exIDs, w.ExerciseID)
	}
	exercises, err := s.repo.ListExercisesByIDs(exIDs)
	if err != nil {
		return []WeakTopic{}
	}
	lessonCount := map[uint]int{}
	for _, e := range exercises {
		lessonCount[e.LessonID]++
	}
	lessonIDs := make([]uint, 0, len(lessonCount))
	for id := range lessonCount {
		lessonIDs = append(lessonIDs, id)
	}
	lessons, err := s.repo.ListLessonsByIDs(lessonIDs)
	if err != nil {
		return []WeakTopic{}
	}
	lessonTitle := map[uint]string{}
	for _, l := range lessons {
		lessonTitle[l.ID] = l.Title
	}

	topics := make([]WeakTopic, 0, len(lessonCount))
	for id, cnt := range lessonCount {
		topics = append(topics, WeakTopic{LessonID: id, Title: lessonTitle[id], WrongCount: cnt})
	}
	sort.Slice(topics, func(i, j int) bool { return topics[i].WrongCount > topics[j].WrongCount })
	if len(topics) > 3 {
		topics = topics[:3]
	}
	return topics
}

func (s *WeeklyReportService) buildSummary(r *WeeklyReportDTO) string {
	return fmt.Sprintf(
		"本周学习 %d 天，累计获得 %d XP：完成 %d 个课时、练习 %d 题、通过 %d 场考试，新增错题 %d 道。当前连续打卡 %d 天。",
		r.ActiveDays, r.XPTotal, r.LessonsDone, r.ExercisesDone, r.ExamsPassed, r.WrongAdded, r.StreakDays,
	)
}

func (s *WeeklyReportService) buildAdvice(r *WeeklyReportDTO) string {
	if r.ActiveDays == 0 {
		return "本周还没有学习记录。从今天开始，完成一个课时，迈出第一步吧！"
	}
	if len(r.WeakTopics) > 0 {
		names := make([]string, 0, len(r.WeakTopics))
		for _, t := range r.WeakTopics {
			name := t.Title
			if name == "" {
				name = fmt.Sprintf("课时 %d", t.LessonID)
			}
			names = append(names, name)
		}
		return "本周薄弱知识点：" + strings.Join(names, "、") + "。建议优先复习这些课时，再做 1 次单元考试巩固。"
	}
	return "本周表现稳定！建议尝试一次单元考试或挑战课程认证考试，检验自己的掌握度。"
}
