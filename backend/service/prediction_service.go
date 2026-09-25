package service

import (
	"fmt"

	"codelearn/model"
	"codelearn/repository"
)

// PredictionService 分数预测 + 考试日历（U7）：基于练习/错题/考试数据预测认证考试通过概率并生成备考任务
type PredictionService struct {
	repo *repository.Repository
}

func NewPredictionService(repo *repository.Repository) *PredictionService {
	return &PredictionService{repo: repo}
}

// PredictionFactor 预测因子
type PredictionFactor struct {
	Label  string `json:"label"`
	Score  int    `json:"score"` // 0-100
	Weight int    `json:"weight"` // 权重（百分数）
	Detail string `json:"detail"`
}

// PrepTask 备考任务
type PrepTask struct {
	Title    string `json:"title"`
	Type     string `json:"type"` // lesson / exam / review / project
	TargetID uint   `json:"target_id,omitempty"`
	Reason   string `json:"reason"`
}

// ExamPrediction 认证考试预测结果
type ExamPrediction struct {
	CourseID    uint               `json:"course_id"`
	CourseTitle string             `json:"course_title"`
	Probability int                `json:"probability"` // 0-100
	Level       string             `json:"level"`      // low / medium / high
	Eligible    bool               `json:"eligible"`
	UnitsPassed int                `json:"units_passed"`
	UnitsTotal  int                `json:"units_total"`
	Factors     []PredictionFactor `json:"factors"`
	Advice      string             `json:"advice"`
	PrepTasks   []PrepTask         `json:"prep_tasks"`
}

// Predict 预测用户对某课程认证考试的通过概率
func (s *PredictionService) Predict(userID, courseID uint) (*ExamPrediction, error) {
	course, err := s.repo.GetCourse(courseID)
	if err != nil {
		return nil, err
	}
	units, _ := s.repo.ListUnitsByCourse(courseID)
	passed := 0
	totalLessons := 0
	completedLessons := 0
	var examScoreSum, examCount int
	for _, u := range units {
		lessons, _ := s.repo.ListLessonsByUnit(u.ID)
		for _, l := range lessons {
			totalLessons++
			if p, err := s.repo.GetProgress(userID, l.ID); err == nil && p.Completed {
				completedLessons++
			}
		}
		if ok, score := s.unitExamPassedAndScore(userID, u.ID); ok {
			passed++
			examScoreSum += score
			examCount++
		}
	}

	// 练习正确率
	correctSub, totalSub := s.practiceAccuracy(userID)
	accuracy := 0
	if totalSub > 0 {
		accuracy = correctSub * 100 / totalSub
	}

	// 错题掌握率
	wrongs, _ := s.repo.ListWrongExercises(userID, false)
	masteredWrong := 0
	for _, w := range wrongs {
		if w.Mastered {
			masteredWrong++
		}
	}
	masteryRate := 0
	if len(wrongs) > 0 {
		masteryRate = masteredWrong * 100 / len(wrongs)
	}

	// 单元考试平均分
	avgExam := 0
	if examCount > 0 {
		avgExam = examScoreSum / examCount
	}

	coverage := 0
	if totalLessons > 0 {
		coverage = completedLessons * 100 / totalLessons
	}

	factors := []PredictionFactor{
		{Label: "练习正确率", Score: accuracy, Weight: 35, Detail: fmt.Sprintf("%d 次作答，正确率 %d%%", totalSub, accuracy)},
		{Label: "错题掌握率", Score: masteryRate, Weight: 20, Detail: fmt.Sprintf("%d 道错题，已掌握 %d%%", len(wrongs), masteryRate)},
		{Label: "单元考试成绩", Score: avgExam, Weight: 30, Detail: fmt.Sprintf("%d 场通过考试，平均 %d 分", examCount, avgExam)},
		{Label: "课时完成度", Score: coverage, Weight: 15, Detail: fmt.Sprintf("%d/%d 课时已完成", completedLessons, totalLessons)},
	}

	prob := 0
	for _, f := range factors {
		prob += f.Score * f.Weight / 100
	}
	if prob == 0 {
		prob = 50 // 无数据时中性默认
	}
	level := "medium"
	if prob >= 75 {
		level = "high"
	} else if prob < 50 {
		level = "low"
	}

	advice := ""
	switch level {
	case "high":
		advice = "备考状态良好，通过概率较高。建议做 1 次模拟认证考试并复习薄弱知识点，保持手感。"
	case "low":
		advice = "当前通过概率偏低。建议先完成剩余课时、稳定练习正确率并掌握错题，再挑战认证考试。"
	default:
		advice = "状态一般。建议优先补齐未完成课时、针对薄弱知识点加强练习后再考试。"
	}

	prepTasks := s.buildPrepTasks(userID, units, passed, len(units))

	return &ExamPrediction{
		CourseID:    courseID,
		CourseTitle: course.Title,
		Probability: prob,
		Level:       level,
		Eligible:    passed >= len(units) && len(units) > 0,
		UnitsPassed: passed,
		UnitsTotal:  len(units),
		Factors:     factors,
		Advice:      advice,
		PrepTasks:   prepTasks,
	}, nil
}

// unitExamPassedAndScore 返回用户是否通过某单元考试及其最近成绩
func (s *PredictionService) unitExamPassedAndScore(userID, unitID uint) (bool, int) {
	exams, err := s.repo.ListExamsByUnit(unitID)
	if err != nil {
		return false, 0
	}
	for _, e := range exams {
		if subs, err := s.repo.ListExamSubmissions(userID, e.ID); err == nil {
			for _, sub := range subs {
				if sub.Passed {
					return true, sub.Score
				}
			}
		}
	}
	return false, 0
}

// practiceAccuracy 练习正确率
func (s *PredictionService) practiceAccuracy(userID uint) (correct, total int) {
	subs, err := s.repo.AllSubmissionsByUser(userID)
	if err != nil {
		return 0, 0
	}
	for _, sub := range subs {
		total++
		if sub.Correct {
			correct++
		}
	}
	return correct, total
}

// buildPrepTasks 生成备考任务（基于缺口）
func (s *PredictionService) buildPrepTasks(userID uint, units []model.Unit, passed, total int) []PrepTask {
	tasks := []PrepTask{}
	if passed < total {
		for _, u := range units {
			if ok, _ := s.unitExamPassedAndScore(userID, u.ID); !ok {
				tasks = append(tasks, PrepTask{
					Title:    "通过单元考试：" + u.Title,
					Type:     "exam",
					TargetID: u.ID,
					Reason:   "完成全部单元考试是参加认证考试的前提",
				})
			}
		}
	}
	return tasks
}
