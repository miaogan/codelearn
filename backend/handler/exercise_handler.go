package handler

import (
	"log"
	"net/http"
	"strconv"

	"codelearn/eino"
	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

type ExerciseHandler struct {
	courseSvc   *service.CourseService
	progressSvc *service.ProgressService
	generator   *eino.ExerciseGenerator
}

func NewExerciseHandler(courseSvc *service.CourseService, progressSvc *service.ProgressService, generator *eino.ExerciseGenerator) *ExerciseHandler {
	return &ExerciseHandler{courseSvc: courseSvc, progressSvc: progressSvc, generator: generator}
}

type submitReq struct {
	Answer string `json:"answer" binding:"required"`
}

func (h *ExerciseHandler) Submit(c *gin.Context) {
	exerciseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "习题 ID 无效"})
		return
	}
	userID := middleware.GetUserID(c)

	var req submitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供答案"})
		return
	}

	correct, explanation, err := h.courseSvc.SubmitAnswer(userID, uint(exerciseID), req.Answer)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "习题不存在"})
		return
	}

	// 答错扣心数
	hearts := 5
	if !correct {
		hearts, _ = h.progressSvc.LoseHeart(userID)
	}

	c.JSON(http.StatusOK, gin.H{
		"correct":     correct,
		"explanation": explanation,
		"hearts":      hearts,
	})
}

type generateReq struct {
	Language   string `json:"language" binding:"required"`
	Topic      string `json:"topic" binding:"required"`
	Count      int    `json:"count"`
	Type       string `json:"type"`
	Difficulty string `json:"difficulty"`
}

func (h *ExerciseHandler) Generate(c *gin.Context) {
	lessonID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Printf("[Generate] 课程ID无效: %s", c.Param("id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程 ID 无效"})
		return
	}

	var req generateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Generate] 参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效: " + err.Error()})
		return
	}
	if req.Count <= 0 || req.Count > 10 {
		req.Count = 3
	}
	if req.Type == "" {
		req.Type = "choice"
	}
	if req.Difficulty == "" {
		req.Difficulty = "easy"
	}

	log.Printf("[Generate] 收到请求: lessonID=%d language=%s topic=%s count=%d type=%s difficulty=%s",
		lessonID, req.Language, req.Topic, req.Count, req.Type, req.Difficulty)

	// 调用 Eino 框架的 ChatModel 生成习题
	exercises, err := h.generator.Generate(c.Request.Context(), eino.GenRequest{
		Language:   req.Language,
		Topic:      req.Topic,
		Count:      req.Count,
		Type:       req.Type,
		Difficulty: req.Difficulty,
	})
	if err != nil {
		log.Printf("[Generate] AI生成失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 生成习题失败: " + err.Error()})
		return
	}

	log.Printf("[Generate] AI生成成功, %d 道习题, 开始保存到数据库", len(exercises))

	// 保存到数据库
	for i := range exercises {
		exercises[i].LessonID = uint(lessonID)
	}
	if err := h.courseSvc.SaveAIGenExercises(uint(lessonID), exercises); err != nil {
		log.Printf("[Generate] 保存到数据库失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存习题失败: " + err.Error()})
		return
	}

	log.Printf("[Generate] 保存成功, 返回结果")

	// 不暴露答案
	for i := range exercises {
		exercises[i].Answer = ""
		exercises[i].TestCases = ""
	}

	c.JSON(http.StatusOK, gin.H{"exercises": exercises})
}

type hintReq struct {
	Question    string `json:"question" binding:"required"`
	UserAnswer  string `json:"user_answer" binding:"required"`
	Language    string `json:"language" binding:"required"`
}

func (h *ExerciseHandler) Hint(c *gin.Context) {
	var req hintReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	hint, err := h.generator.GenerateHint(c.Request.Context(), req.Question, req.UserAnswer, req.Language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成提示失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hint": hint})
}

// GetTestCase 获取代码题的测试用例（供前端展示，不含答案）
func (h *ExerciseHandler) GetTestCase(c *gin.Context) {
	exerciseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "习题 ID 无效"})
		return
	}

	ex, err := h.courseSvc.GetExercise(uint(exerciseID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "习题不存在"})
		return
	}

	// 只返回代码模板，不返回测试用例
	c.JSON(http.StatusOK, gin.H{
		"code_template": ex.CodeTemplate,
		"type":          ex.Type,
		"question":      ex.Question,
		"difficulty":    ex.Difficulty,
	})
}

// ExamSubmit 考试模式批量提交：不扣心数，统一打分
type examAnswerItem struct {
	ExerciseID uint   `json:"exercise_id"`
	Answer     string `json:"answer"`
}

type examSubmitReq struct {
	Answers []examAnswerItem `json:"answers" binding:"required"`
}

type examResultItem struct {
	ExerciseID  uint   `json:"exercise_id"`
	Correct     bool   `json:"correct"`
	UserAnswer  string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	Explanation string `json:"explanation"`
	Feedback    string `json:"feedback,omitempty"`
}

func (h *ExerciseHandler) ExamSubmit(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req examSubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ExamSubmit] 参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	log.Printf("[ExamSubmit] userID=%d 提交 %d 道题", userID, len(req.Answers))

	results := make([]examResultItem, 0, len(req.Answers))
	correctCount := 0

	for _, item := range req.Answers {
		ex, err := h.courseSvc.GetExercise(item.ExerciseID)
		if err != nil {
			log.Printf("[ExamSubmit] 习题不存在: %d", item.ExerciseID)
			continue
		}

		result := examResultItem{
			ExerciseID:   item.ExerciseID,
			UserAnswer:   item.Answer,
			CorrectAnswer: ex.Answer,
			Explanation:  ex.Explanation,
		}

		// 非标准客观题类型（subjective, short_answer, essay 等）都用 AI 判定
		isObjective := ex.Type == "choice" || ex.Type == "fillblank" || ex.Type == "order"
		if isObjective {
			// 客观题直接比对
			correct, _, _ := h.courseSvc.SubmitAnswer(userID, item.ExerciseID, item.Answer)
			result.Correct = correct
		} else {
			// 主观题/其他类型调用 AI 判定（只传题目和学生作答，不传参考答案）
			log.Printf("[ExamSubmit] AI 判定: exerciseID=%d type=%s", item.ExerciseID, ex.Type)
			correct, feedback, err := h.generator.JudgeSubjective(
				c.Request.Context(), ex.Question, item.Answer,
			)
			if err != nil {
				log.Printf("[ExamSubmit] AI 判定失败: %v", err)
				result.Correct = false
				result.Feedback = "AI 判定失败: " + err.Error()
			} else {
				result.Correct = correct
				result.Feedback = feedback
			}
		}

		if result.Correct {
			correctCount++
		}
		results = append(results, result)
	}

	log.Printf("[ExamSubmit] 完成: %d/%d 正确", correctCount, len(results))

	score := 0
	if len(results) > 0 {
		score = correctCount * 100 / len(results)
	}

	c.JSON(http.StatusOK, gin.H{
		"results":       results,
		"correct_count": correctCount,
		"total_count":   len(results),
		"score":         score,
	})
}


