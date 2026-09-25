package service

import (
	"fmt"
	"sort"

	"codelearn/model"
	"codelearn/repository"
)

const (
	SkillStatusNotStarted = "未学"
	SkillStatusLearning   = "学习中"
	SkillStatusMastered   = "已掌握"
	SkillStatusWeak       = "薄弱"

	// 能力图谱权重
	skillWeightCompleted = 50 // 完成课时
	skillWeightPractice  = 40 // 练习正确率（最高）
	skillWeightExam      = 10 // 通过单元考试
	skillWrongPenalty    = 25 // 每条未掌握错题
)

// SkillNode 技能节点（以课时为知识点粒度）
type SkillNode struct {
	LessonID    uint   `json:"lesson_id"`
	LessonTitle string `json:"lesson_title"`
	Icon        string `json:"icon"`
	UnitID      uint   `json:"unit_id"`
	UnitTitle   string `json:"unit_title"`
	Mastery     int    `json:"mastery"` // 0-100
	Status      string `json:"status"`  // 未学 / 学习中 / 已掌握 / 薄弱
	Completed   bool   `json:"completed"`
}

// SkillMap 课程能力图谱
type SkillMap struct {
	Course  model.Course `json:"course"`
	Nodes   []SkillNode  `json:"nodes"`
	Overall int          `json:"overall"` // 已开始知识点平均掌握度
}

// ReviewRecommendation 复习推荐项（薄弱章节）
type ReviewRecommendation struct {
	LessonID    uint   `json:"lesson_id"`
	LessonTitle string `json:"lesson_title"`
	CourseID    uint   `json:"course_id"`
	CourseTitle string `json:"course_title"`
	UnitID      uint   `json:"unit_id"`
	WeakCount   int    `json:"weak_count"`
	Mastery     int    `json:"mastery"`
	Reason      string `json:"reason"`
}

type SkillMapService struct {
	repo *repository.Repository
}

func NewSkillMapService(repo *repository.Repository) *SkillMapService {
	return &SkillMapService{repo: repo}
}

// CourseSkillMap 计算课程能力图谱
func (s *SkillMapService) CourseSkillMap(courseID, userID uint) (*SkillMap, error) {
	course, err := s.repo.GetCourse(courseID)
	if err != nil {
		return nil, err
	}
	units, err := s.repo.ListUnitsByCourse(courseID)
	if err != nil {
		return nil, err
	}

	progress, _ := s.repo.ListProgressByUser(userID)
	completed := make(map[uint]bool)
	for _, p := range progress {
		completed[p.LessonID] = p.Completed
	}

	unmastered, _ := s.repo.ListWrongExercises(userID, true)
	weakSet := make(map[uint]bool)
	for _, w := range unmastered {
		weakSet[w.ExerciseID] = true
	}

	passedUnits := s.passedExamUnits(userID, units)

	skill := &SkillMap{Course: *course, Nodes: make([]SkillNode, 0)}
	totalMastery := 0
	started := 0
	for _, u := range units {
		lessons, err := s.repo.ListLessonsByUnit(u.ID)
		if err != nil {
			continue
		}
		for _, l := range lessons {
			exs, _ := s.repo.ListExercisesByLesson(l.ID)
			correct, total, weakCount := s.lessonPracticeStats(userID, exs, weakSet)
			node := s.buildNode(l, u, completed[l.ID], total, correct, weakCount, passedUnits[u.ID])
			if node.Status != SkillStatusNotStarted {
				totalMastery += node.Mastery
				started++
			}
			skill.Nodes = append(skill.Nodes, node)
		}
	}
	if started > 0 {
		skill.Overall = totalMastery / started
	}
	return skill, nil
}

// ReviewRecommendations 薄弱章节复习推荐（按未掌握错题数排序）
func (s *SkillMapService) ReviewRecommendations(userID uint, limit int) ([]ReviewRecommendation, error) {
	if limit <= 0 {
		limit = 5
	}
	unmastered, _ := s.repo.ListWrongExercises(userID, true)
	if len(unmastered) == 0 {
		return []ReviewRecommendation{}, nil
	}

	exIDs := make([]uint, 0, len(unmastered))
	for _, w := range unmastered {
		exIDs = append(exIDs, w.ExerciseID)
	}
	exs, err := s.repo.ListExercisesByIDs(exIDs)
	if err != nil {
		return nil, err
	}
	exMap := make(map[uint]model.Exercise, len(exs))
	for _, e := range exs {
		exMap[e.ID] = e
	}

	lessonCount := map[uint]int{}
	for _, w := range unmastered {
		if ex, ok := exMap[w.ExerciseID]; ok {
			lessonCount[ex.LessonID]++
		}
	}

	lessonIDs := make([]uint, 0, len(lessonCount))
	for id := range lessonCount {
		lessonIDs = append(lessonIDs, id)
	}
	lessons, _ := s.repo.ListLessonsByIDs(lessonIDs)
	lessonMap := make(map[uint]model.Lesson, len(lessons))
	for _, l := range lessons {
		lessonMap[l.ID] = l
	}

	recs := make([]ReviewRecommendation, 0, len(lessonCount))
	for lessonID, count := range lessonCount {
		l, ok := lessonMap[lessonID]
		if !ok {
			continue
		}
		u, err := s.repo.GetUnit(l.UnitID)
		if err != nil {
			continue
		}
		c, err := s.repo.GetCourse(u.CourseID)
		if err != nil {
			continue
		}
		exs2, _ := s.repo.ListExercisesByLesson(lessonID)
		correct, total, _ := s.lessonPracticeStats(userID, exs2, nil)
		mastery := 0
		if total > 0 {
			mastery += correct * skillWeightPractice / total
		}
		mastery -= skillWrongPenalty * count
		if mastery < 0 {
			mastery = 0
		}
		recs = append(recs, ReviewRecommendation{
			LessonID:    lessonID,
			LessonTitle: l.Title,
			CourseID:    u.CourseID,
			CourseTitle: c.Title,
			UnitID:      u.ID,
			WeakCount:   count,
			Mastery:     mastery,
			Reason:      fmt.Sprintf("有 %d 道错题未掌握", count),
		})
	}
	sort.Slice(recs, func(i, j int) bool { return recs[i].WeakCount > recs[j].WeakCount })
	if len(recs) > limit {
		recs = recs[:limit]
	}
	return recs, nil
}

// lessonPracticeStats 统计课时练习情况（正确数/总数/错题数）
func (s *SkillMapService) lessonPracticeStats(userID uint, exercises []model.Exercise, weakSet map[uint]bool) (correct, total, weakCount int) {
	exIDs := make([]uint, 0, len(exercises))
	for _, e := range exercises {
		exIDs = append(exIDs, e.ID)
		if weakSet != nil && weakSet[e.ID] {
			weakCount++
		}
	}
	subs, _ := s.repo.ListSubmissionsForExercises(userID, exIDs)
	for _, sub := range subs {
		total++
		if sub.Correct {
			correct++
		}
	}
	return correct, total, weakCount
}

// buildNode 计算单课时掌握度与状态
func (s *SkillMapService) buildNode(l model.Lesson, u model.Unit, isCompleted bool, total, correct, weakCount int, examPassed bool) SkillNode {
	mastery := 0
	if isCompleted {
		mastery += skillWeightCompleted
	}
	if total > 0 {
		mastery += correct * skillWeightPractice / total
	}
	if examPassed {
		mastery += skillWeightExam
	}
	mastery -= skillWrongPenalty * weakCount
	if mastery < 0 {
		mastery = 0
	}
	if mastery > 100 {
		mastery = 100
	}

	status := SkillStatusNotStarted
	if isCompleted || total > 0 || weakCount > 0 {
		switch {
		case mastery >= 70:
			status = SkillStatusMastered
		case mastery < 40:
			status = SkillStatusWeak
		default:
			status = SkillStatusLearning
		}
	}
	return SkillNode{
		LessonID:    l.ID,
		LessonTitle: l.Title,
		Icon:        l.Icon,
		UnitID:      u.ID,
		UnitTitle:   u.Title,
		Mastery:     mastery,
		Status:      status,
		Completed:   isCompleted,
	}
}

// passedExamUnits 找出用户已通过单元考试的单元
func (s *SkillMapService) passedExamUnits(userID uint, units []model.Unit) map[uint]bool {
	passed := make(map[uint]bool)
	for _, u := range units {
		exams, _ := s.repo.ListExamsByUnit(u.ID)
		for _, exam := range exams {
			sub, err := s.repo.GetLatestExamSubmission(userID, exam.ID)
			if err == nil && sub.Passed {
				passed[u.ID] = true
			}
		}
	}
	return passed
}
