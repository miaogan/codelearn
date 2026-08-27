package service

import (
	"time"

	"codelearn/model"
	"codelearn/repository"
)

type CourseService struct {
	repo *repository.Repository
}

func NewCourseService(repo *repository.Repository) *CourseService {
	return &CourseService{repo: repo}
}

// SkillTreeLesson 技能树中的课程节点
type SkillTreeLesson struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Icon        string `json:"icon"`
	Order       int    `json:"order"`
	Unlocked    bool   `json:"unlocked"`
	Completed   bool   `json:"completed"`
	Score       int    `json:"score"`
}

type SkillTreeUnit struct {
	ID          uint              `json:"id"`
	Title       string `json:"title"`
	Description string            `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Order       int               `json:"order"`
	Lessons     []SkillTreeLesson `json:"lessons"`
}

type LearningPath struct {
	Course model.Course  `json:"course"`
	Units  []SkillTreeUnit `json:"units"`
}

func (s *CourseService) ListCourses() ([]model.Course, error) {
	return s.repo.ListCourses()
}

func (s *CourseService) GetCourse(courseID uint) (*model.Course, error) {
	return s.repo.GetCourse(courseID)
}

// GetLearningPath 获取技能树（含解锁状态）
func (s *CourseService) GetLearningPath(courseID, userID uint) (*LearningPath, error) {
	course, err := s.repo.GetCourse(courseID)
	if err != nil {
		return nil, err
	}

	units, err := s.repo.ListUnitsByCourse(courseID)
	if err != nil {
		return nil, err
	}

	progress, err := s.repo.ListProgressByUser(userID)
	if err != nil {
		return nil, err
	}
	completed := make(map[uint]bool)
	for _, p := range progress {
		completed[p.LessonID] = p.Completed
	}

	// 按顺序遍历，前一课完成才解锁后一课
	prevCompleted := true
	treeUnits := make([]SkillTreeUnit, 0, len(units))
	for _, u := range units {
		lessons, err := s.repo.ListLessonsByUnit(u.ID)
		if err != nil {
			return nil, err
		}
		treeLessons := make([]SkillTreeLesson, 0, len(lessons))
		for _, l := range lessons {
			isDone := completed[l.ID]
			treeLessons = append(treeLessons, SkillTreeLesson{
				ID:        l.ID,
				Title:     l.Title,
				Icon:      l.Icon,
				Order:     l.Order,
				Unlocked:  prevCompleted,
				Completed: isDone,
			})
			if !isDone {
				prevCompleted = false
			}
		}
		treeUnits = append(treeUnits, SkillTreeUnit{
			ID:          u.ID,
			Title:       u.Title,
			Description: u.Description,
			Icon:        u.Icon,
			Color:       u.Color,
			Order:       u.Order,
			Lessons:     treeLessons,
		})
	}

	return &LearningPath{Course: *course, Units: treeUnits}, nil
}

func (s *CourseService) GetLesson(lessonID uint) (*model.Lesson, error) {
	return s.repo.GetLesson(lessonID)
}

func (s *CourseService) GetExercises(lessonID uint) ([]model.Exercise, error) {
	exercises, err := s.repo.ListExercisesByLesson(lessonID)
	if err != nil {
		return nil, err
	}
	// 不暴露答案给前端
	for i := range exercises {
		exercises[i].Answer = ""
		exercises[i].TestCases = ""
	}
	return exercises, nil
}

// SubmitAnswer 提交答案并判定对错（内部使用，不暴露正确答案）
func (s *CourseService) SubmitAnswer(userID, exerciseID uint, userAnswer string) (correct bool, explanation string, err error) {
	ex, err := s.repo.GetExercise(exerciseID)
	if err != nil {
		return false, "", err
	}

	correct = checkAnswer(ex, userAnswer)

	submission := &model.Submission{
		UserID:     userID,
		ExerciseID: exerciseID,
		Answer:     userAnswer,
		Correct:    correct,
		CreatedAt:  time.Now(),
	}
	_ = s.repo.CreateSubmission(submission)

	return correct, ex.Explanation, nil
}

func (s *CourseService) GetExercise(exerciseID uint) (*model.Exercise, error) {
	return s.repo.GetExercise(exerciseID)
}

// SaveAIGenExercises 保存 AI 生成的习题到数据库
func (s *CourseService) SaveAIGenExercises(lessonID uint, exercises []model.Exercise) error {
	for i := range exercises {
		exercises[i].LessonID = lessonID
	}
	return s.repo.CreateExercises(exercises)
}

func checkAnswer(ex *model.Exercise, userAnswer string) bool {
	switch ex.Type {
	case "choice", "fillblank", "order":
		return normalize(userAnswer) == normalize(ex.Answer)
	case "code":
		return false // 代码题由沙箱评判
	default:
		return normalize(userAnswer) == normalize(ex.Answer)
	}
}
