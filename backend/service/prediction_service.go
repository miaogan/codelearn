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
	Score  int    `json:"score"`  // 0-100
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
	Level       string             `json:"level"`       // low / medium / high
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

	// 一次批量加载课程单元考试与用户全部考试提交，避免逐单元 N+1
	unitIDs := make([]uint, 0, len(units))
	for _, u := range units {
		unitIDs = append(unitIDs, u.ID)
	}
	exams, _ := s.repo.ListExamsByUnitIDs(unitIDs)
	examsByUnit := map[uint][]model.Exam{}
	for _, e := range exams {
		examsByUnit[e.UnitID] = append(examsByUnit[e.UnitID], e)
	}
	allSubs, _ := s.repo.ListExamSubmissionsByUser(userID)
	latestByExam := map[uint]*model.ExamSubmission{}
	for i := range allSubs {
		latestByExam[allSubs[i].ExamID] = &allSubs[i] // CreatedAt 升序，后写覆盖即最近一次
	}

	passed := 0
	totalLessons := 0
	completedLessons := 0
	examScoreSum := 0
	examUnits := 0
	passedByUnit := map[uint]bool{}
	for _, u := range units {
		lessons, _ := s.repo.ListLessonsByUnit(u.ID)
		for _, l := range lessons {
			totalLessons++
			if p, err := s.repo.GetProgress(userID, l.ID); err == nil && p.Completed {
				completedLessons++
			}
		}
		ok, score, attempted := unitExamStatus(examsByUnit[u.ID], latestByExam)
		passedByUnit[u.ID] = ok
		if attempted {
			examScoreSum += score
			examUnits++
		}
		if ok {
			passed++
		}
	}

	// 练习正确率
	correctSub, totalSub := s.practiceAccuracy(userID)
	accuracy := 0
	if totalSub > 0 {
		accuracy = correctSub * 100 / totalSub
	}

	// 错题掌握率（错题本按习题唯一，含练习与考试来源；练习正确率已由考试作答不写流水保证纯净）
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

	// 单元考试平均分（含未通过尝试的最近一次成绩，避免只看通过场次的乐观偏差）
	avgExam := 0
	if examUnits > 0 {
		avgExam = examScoreSum / examUnits
	}

	coverage := 0
	if totalLessons > 0 {
		coverage = completedLessons * 100 / totalLessons
	}

	factors := []PredictionFactor{
		{Label: "练习正确率", Score: accuracy, Weight: 35, Detail: fmt.Sprintf("%d 次作答，正确率 %d%%", totalSub, accuracy)},
		{Label: "错题掌握率", Score: masteryRate, Weight: 20, Detail: fmt.Sprintf("%d 道错题，已掌握 %d%%", len(wrongs), masteryRate)},
		{Label: "单元考试成绩", Score: avgExam, Weight: 30, Detail: fmt.Sprintf("%d 个单元有考试记录，平均 %d 分", examUnits, avgExam)},
		{Label: "课时完成度", Score: coverage, Weight: 15, Detail: fmt.Sprintf("%d/%d 课时已完成", completedLessons, totalLessons)},
	}

	prob := 0
	for _, f := range factors {
		prob += f.Score * f.Weight / 100
	}
	hasData := totalSub > 0 || len(wrongs) > 0 || examUnits > 0 || completedLessons > 0
	if prob == 0 && !hasData {
		prob = 50 // 完全无数据时中性默认；有数据但概率为 0 则如实反映
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

	prepTasks := s.buildPrepTasks(units, passedByUnit)

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

// unitExamStatus 基于批量预加载的单元考试与最近提交判断最近表现：
// 通过（任一考试实例最近一次提交通过，与认证资格判定一致）、最近分数、是否有过考试尝试。
// 未考由 attempted=false 区分，避免虚增平均分。
func unitExamStatus(exams []model.Exam, latestByExam map[uint]*model.ExamSubmission) (passed bool, score int, attempted bool) {
	var latest *model.ExamSubmission
	for _, e := range exams {
		sub, ok := latestByExam[e.ID]
		if !ok {
			continue
		}
		if sub.Passed {
			passed = true
		}
		if latest == nil || sub.CreatedAt.After(latest.CreatedAt) {
			latest = sub
		}
	}
	if latest == nil {
		return false, 0, false
	}
	return passed, latest.Score, true
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

// buildPrepTasks 生成备考任务（基于缺口，复用 Predict 已算出的各单元通过状态）
func (s *PredictionService) buildPrepTasks(units []model.Unit, passedByUnit map[uint]bool) []PrepTask {
	tasks := []PrepTask{}
	for _, u := range units {
		if !passedByUnit[u.ID] {
			tasks = append(tasks, PrepTask{
				Title:    "通过单元考试：" + u.Title,
				Type:     "exam",
				TargetID: u.ID,
				Reason:   "完成全部单元考试是参加认证考试的前提",
			})
		}
	}
	return tasks
}
