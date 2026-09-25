package service

import (
	"errors"
	"math/rand"
	"time"

	"codelearn/model"
	"codelearn/repository"
	"codelearn/sandbox"
)

var (
	ErrUnitNotCompleted = errors.New("请先完成该单元的全部课时")
	ErrNoQuestions      = errors.New("该单元暂无足够的题目组成试卷")
	ErrAlreadyPassed    = errors.New("该考试已通过，无需重考")
	ErrRetryCoolDown    = errors.New("24 小时内不可重考，请稍后再试")
)

const (
	ExamTypeUnit = "unit"
	ExamTypeCert = "cert"

	// ExamPassScore 单元考试及格线
	ExamPassScore = 60
	// ExamXP 考试通过奖励 XP
	ExamXP = 20
	// ExamRetryCoolDownHours 未通过后的重考冷却时长
	ExamRetryCoolDownHours = 24
)

// ExamQuestionDTO 下发给前端的考题（不含答案）
type ExamQuestionDTO struct {
	ID           uint   `json:"id"` // ExamQuestion ID
	ExerciseID   uint   `json:"exercise_id"`
	Order        int    `json:"order"`
	Type         string `json:"type"`
	Question     string `json:"question"`
	Options      string `json:"options"`
	CodeTemplate string `json:"code_template"`
	Difficulty   string `json:"difficulty"`
}

// ExamDTO 考试信息与试卷
type ExamDTO struct {
	ID          uint              `json:"id"`
	Title       string            `json:"title"`
	ExamType    string            `json:"exam_type"`
	DurationMin int               `json:"duration_min"`
	PassScore   int               `json:"pass_score"`
	Questions   []ExamQuestionDTO `json:"questions"`
}

// ExamAnswerItem 考试提交的一道答案
type ExamAnswerItem struct {
	QuestionID uint   `json:"question_id"`
	ExerciseID uint   `json:"exercise_id"`
	Answer     string `json:"answer"` // 代码题为用户提交的代码
}

// ExamResultItemDTO 考试结果单题
type ExamResultItemDTO struct {
	QuestionID    uint   `json:"question_id"`
	ExerciseID    uint   `json:"exercise_id"`
	Correct       bool   `json:"correct"`
	UserAnswer    string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	Explanation   string `json:"explanation"`
}

// TypeAccuracy 分题型正确率
type TypeAccuracy struct {
	Type     string `json:"type"`
	Total    int    `json:"total"`
	Correct  int    `json:"correct"`
	Accuracy int    `json:"accuracy"` // 百分比 0-100
}

// KnowledgePoint 知识点掌握度（以课时为知识点粒度）
type KnowledgePoint struct {
	LessonID uint   `json:"lesson_id"`
	Title    string `json:"title"`
	Total    int    `json:"total"`
	Correct  int    `json:"correct"`
	Mastery  int    `json:"mastery"` // 百分比 0-100
}

// ExamReportDTO 考试报告
type ExamReportDTO struct {
	ExamID       uint                `json:"exam_id"`
	Score        int                 `json:"score"`
	CorrectCount int                 `json:"correct_count"`
	TotalCount   int                 `json:"total_count"`
	Passed       bool                `json:"passed"`
	PassScore    int                 `json:"pass_score"`
	DurationSec  int                 `json:"duration_sec"`
	TabSwitches  int                 `json:"tab_switches"`
	Attempts     int                 `json:"attempts"`
	BestScore    int                 `json:"best_score"`
	ByType       []TypeAccuracy      `json:"by_type"`
	Knowledge    []KnowledgePoint    `json:"knowledge"`
	Results      []ExamResultItemDTO `json:"results"`
}

type ExamService struct {
	repo *repository.Repository
}

func NewExamService(repo *repository.Repository) *ExamService {
	return &ExamService{repo: repo}
}

// AssembleUnitExam 组装单元考试：校验课时完成 → 按单元习题组卷 → 创建考试快照
func (s *ExamService) AssembleUnitExam(unitID, userID uint) (*ExamDTO, error) {
	unit, err := s.repo.GetUnit(unitID)
	if err != nil {
		return nil, err
	}

	lessons, err := s.repo.ListLessonsByUnit(unitID)
	if err != nil {
		return nil, err
	}
	if len(lessons) == 0 {
		return nil, ErrNoQuestions
	}
	// 前置校验：单元内所有课时必须已完成
	for _, l := range lessons {
		p, err := s.repo.GetProgress(userID, l.ID)
		if err != nil || !p.Completed {
			return nil, ErrUnitNotCompleted
		}
	}

	exercises, err := s.repo.ListExercisesByUnit(unitID)
	if err != nil {
		return nil, err
	}
	if len(exercises) == 0 {
		return nil, ErrNoQuestions
	}

	selected := pickExamQuestions(exercises, unit.ExamQuestionCount)

	exam := &model.Exam{
		CourseID:      unit.CourseID,
		UnitID:        unitID,
		Title:         unit.Title + " 单元考试",
		ExamType:      ExamTypeUnit,
		DurationMin:   unit.ExamDurationMin,
		QuestionCount: len(selected),
		PassScore:     ExamPassScore,
	}
	if err := s.repo.CreateExam(exam); err != nil {
		return nil, err
	}

	qs := make([]model.ExamQuestion, 0, len(selected))
	for i, ex := range selected {
		qs = append(qs, model.ExamQuestion{ExamID: exam.ID, ExerciseID: ex.ID, Order: i + 1})
	}
	if err := s.repo.CreateExamQuestions(qs); err != nil {
		return nil, err
	}

	dto := &ExamDTO{
		ID:          exam.ID,
		Title:       exam.Title,
		ExamType:    exam.ExamType,
		DurationMin: exam.DurationMin,
		PassScore:   exam.PassScore,
		Questions:   make([]ExamQuestionDTO, 0, len(selected)),
	}
	for i, ex := range selected {
		dto.Questions = append(dto.Questions, ExamQuestionDTO{
			ID:           qs[i].ID,
			ExerciseID:   ex.ID,
			Order:        qs[i].Order,
			Type:         ex.Type,
			Question:     ex.Question,
			Options:      ex.Options,
			CodeTemplate: ex.CodeTemplate,
			Difficulty:   ex.Difficulty,
		})
	}
	return dto, nil
}

// SubmitExam 提交考试并判分，生成考试报告，错题自动进入错题本。
// 重考规则：已通过不可重考；未通过需等待 ExamRetryCoolDownHours 冷却。
func (s *ExamService) SubmitExam(userID, examID uint, answers []ExamAnswerItem, durationSec, tabSwitches int) (*ExamReportDTO, error) {
	exam, err := s.repo.GetExam(examID)
	if err != nil {
		return nil, err
	}

	// 重考冷却与已通过拦截
	latest, err := s.repo.GetLatestExamSubmission(userID, examID)
	if err == nil {
		if latest.Passed {
			return nil, ErrAlreadyPassed
		}
		if time.Since(latest.CreatedAt) < ExamRetryCoolDownHours*time.Hour {
			return nil, ErrRetryCoolDown
		}
	}

	qs, err := s.repo.ListExamQuestions(examID)
	if err != nil {
		return nil, err
	}

	exIDs := make([]uint, 0, len(qs))
	for _, q := range qs {
		exIDs = append(exIDs, q.ExerciseID)
	}
	exList, err := s.repo.ListExercisesByIDs(exIDs)
	if err != nil {
		return nil, err
	}
	exMap := make(map[uint]model.Exercise, len(exList))
	for _, e := range exList {
		exMap[e.ID] = e
	}

	courseLang, _ := s.repo.GetCourseLanguage(exam.CourseID)

	answerMap := make(map[uint]string, len(answers))
	for _, a := range answers {
		answerMap[a.QuestionID] = a.Answer
	}

	report := &ExamReportDTO{
		ExamID:      exam.ID,
		TotalCount:  len(qs),
		PassScore:   exam.PassScore,
		DurationSec: durationSec,
		TabSwitches: tabSwitches,
		Results:     make([]ExamResultItemDTO, 0, len(qs)),
	}

	for _, q := range qs {
		ex, ok := exMap[q.ExerciseID]
		if !ok {
			continue
		}
		userAnswer := answerMap[q.ID]
		correct := false
		if ex.Type == "code" {
			// 代码题：用沙箱跑测试用例
			judge := sandbox.JudgeCodeJSON(courseLang, userAnswer, ex.TestCases)
			correct = judge.AllPass && judge.TotalCount > 0
		} else {
			correct = normalize(userAnswer) == normalize(ex.Answer)
		}
		if correct {
			report.CorrectCount++
		}
		report.Results = append(report.Results, ExamResultItemDTO{
			QuestionID:    q.ID,
			ExerciseID:    q.ExerciseID,
			Correct:       correct,
			UserAnswer:    userAnswer,
			CorrectAnswer: ex.Answer,
			Explanation:   ex.Explanation,
		})
	}

	if report.TotalCount > 0 {
		report.Score = report.CorrectCount * 100 / report.TotalCount
	}
	report.Passed = report.Score >= exam.PassScore
	report.ByType = computeTypeAccuracy(report.Results, exMap)
	report.Knowledge = s.computeKnowledgePoints(report.Results, exMap)

	// 错题自动进入错题本（标记来源为考试）
	for _, r := range report.Results {
		if !r.Correct {
			_ = s.repo.UpsertWrongExercise(userID, r.ExerciseID, r.UserAnswer, "exam")
		}
	}

	submission := &model.ExamSubmission{
		UserID:       userID,
		ExamID:       examID,
		Score:        report.Score,
		CorrectCount: report.CorrectCount,
		TotalCount:   report.TotalCount,
		DurationSec:  durationSec,
		TabSwitches:  tabSwitches,
		Passed:       report.Passed,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.CreateExamSubmission(submission); err != nil {
		return nil, err
	}

	report.Attempts, report.BestScore = s.examSummary(userID, examID)
	return report, nil
}

// examSummary 统计考试尝试次数与最高分
func (s *ExamService) examSummary(userID, examID uint) (attempts, bestScore int) {
	subs, err := s.repo.ListExamSubmissions(userID, examID)
	if err != nil {
		return 0, 0
	}
	for _, sub := range subs {
		if sub.Score > bestScore {
			bestScore = sub.Score
		}
	}
	return len(subs), bestScore
}

// computeTypeAccuracy 统计分题型正确率
func computeTypeAccuracy(results []ExamResultItemDTO, exMap map[uint]model.Exercise) []TypeAccuracy {
	order := []string{"choice", "fillblank", "code"}
	counts := map[string]*TypeAccuracy{}
	for _, t := range order {
		counts[t] = &TypeAccuracy{Type: t}
	}
	for _, r := range results {
		ex, ok := exMap[r.ExerciseID]
		if !ok {
			continue
		}
		item, ok := counts[ex.Type]
		if !ok {
			item = &TypeAccuracy{Type: ex.Type}
			counts[ex.Type] = item
		}
		item.Total++
		if r.Correct {
			item.Correct++
		}
	}
	out := make([]TypeAccuracy, 0, len(counts))
	for _, t := range order {
		if item, ok := counts[t]; ok && item.Total > 0 {
			item.Accuracy = item.Correct * 100 / item.Total
			out = append(out, *item)
		}
	}
	return out
}

// computeKnowledgePoints 按课时（知识点）统计掌握度
func (s *ExamService) computeKnowledgePoints(results []ExamResultItemDTO, exMap map[uint]model.Exercise) []KnowledgePoint {
	type agg struct {
		lessonID uint
		title    string
		total    int
		correct  int
	}
	aggs := map[uint]*agg{}
	for _, r := range results {
		ex, ok := exMap[r.ExerciseID]
		if !ok {
			continue
		}
		a, ok := aggs[ex.LessonID]
		if !ok {
			a = &agg{lessonID: ex.LessonID}
			aggs[ex.LessonID] = a
		}
		a.total++
		if r.Correct {
			a.correct++
		}
	}

	lessonIDs := make([]uint, 0, len(aggs))
	for id := range aggs {
		lessonIDs = append(lessonIDs, id)
	}
	lessons, _ := s.repo.ListLessonsByIDs(lessonIDs)
	titleMap := make(map[uint]string, len(lessons))
	for _, l := range lessons {
		titleMap[l.ID] = l.Title
	}

	points := make([]KnowledgePoint, 0, len(aggs))
	for _, a := range aggs {
		mastery := 0
		if a.total > 0 {
			mastery = a.correct * 100 / a.total
		}
		points = append(points, KnowledgePoint{
			LessonID: a.lessonID,
			Title:    titleMap[a.lessonID],
			Total:    a.total,
			Correct:  a.correct,
			Mastery:  mastery,
		})
	}
	return points
}

// GetReport 获取用户最近一次考试的报告概要
func (s *ExamService) GetReport(userID, examID uint) (*ExamReportDTO, error) {
	exam, err := s.repo.GetExam(examID)
	if err != nil {
		return nil, err
	}
	sub, err := s.repo.GetLatestExamSubmission(userID, examID)
	if err != nil {
		return nil, err
	}
	attempts, best := s.examSummary(userID, examID)
	return &ExamReportDTO{
		ExamID:       exam.ID,
		Score:        sub.Score,
		CorrectCount: sub.CorrectCount,
		TotalCount:   sub.TotalCount,
		Passed:       sub.Passed,
		PassScore:    exam.PassScore,
		DurationSec:  sub.DurationSec,
		TabSwitches:  sub.TabSwitches,
		Attempts:     attempts,
		BestScore:    best,
	}, nil
}

// pickExamQuestions 按题型比例组卷：客观题优先，代码题不超过 30%
func pickExamQuestions(exercises []model.Exercise, targetCount int) []model.Exercise {
	if targetCount <= 0 {
		targetCount = 10
	}
	if len(exercises) <= targetCount {
		shuffled := make([]model.Exercise, len(exercises))
		copy(shuffled, exercises)
		rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		return shuffled
	}

	var objective, code []model.Exercise
	for _, ex := range exercises {
		if ex.Type == "code" {
			code = append(code, ex)
		} else {
			objective = append(objective, ex)
		}
	}
	rand.Shuffle(len(objective), func(i, j int) { objective[i], objective[j] = objective[j], objective[i] })
	rand.Shuffle(len(code), func(i, j int) { code[i], code[j] = code[j], code[i] })

	codeCount := targetCount * 30 / 100
	if codeCount > len(code) {
		codeCount = len(code)
	}
	objCount := targetCount - codeCount
	if objCount > len(objective) {
		objCount = len(objective)
	}

	selected := make([]model.Exercise, 0, objCount+codeCount)
	selected = append(selected, objective[:objCount]...)
	selected = append(selected, code[:codeCount]...)
	rand.Shuffle(len(selected), func(i, j int) { selected[i], selected[j] = selected[j], selected[i] })
	return selected
}
