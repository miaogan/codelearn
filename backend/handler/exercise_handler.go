package handler

import (
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程 ID 无效"})
		return
	}

	var req generateReq
	if err := c.ShouldBindJSON(&req); err != nil {
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

	// 调用 Eino 框架的 ChatModel 生成习题
	exercises, err := h.generator.Generate(c.Request.Context(), eino.GenRequest{
		Language:   req.Language,
		Topic:      req.Topic,
		Count:      req.Count,
		Type:       req.Type,
		Difficulty: req.Difficulty,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 生成习题失败: " + err.Error()})
		return
	}

	// 保存到数据库
	for i := range exercises {
		exercises[i].LessonID = uint(lessonID)
	}
	if err := h.courseSvc.SaveAIGenExercises(uint(lessonID), exercises); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存习题失败"})
		return
	}

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


