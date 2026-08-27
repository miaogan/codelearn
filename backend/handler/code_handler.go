package handler

import (
	"net/http"
	"strconv"

	"codelearn/middleware"
	"codelearn/sandbox"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

type CodeHandler struct {
	courseSvc   *service.CourseService
	progressSvc *service.ProgressService
}

func NewCodeHandler(courseSvc *service.CourseService, progressSvc *service.ProgressService) *CodeHandler {
	return &CodeHandler{courseSvc: courseSvc, progressSvc: progressSvc}
}

type runReq struct {
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// Run 直接运行用户代码，返回输出
func (h *CodeHandler) Run(c *gin.Context) {
	var req runReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	result := sandbox.RunCode(req.Language, req.Code)
	c.JSON(http.StatusOK, result)
}

type judgeReq struct {
	ExerciseID uint   `json:"exercise_id"`
	Language   string `json:"language" binding:"required"`
	Code       string `json:"code" binding:"required"`
}

// Judge 评判用户提交的代码题
func (h *CodeHandler) Judge(c *gin.Context) {
	var req judgeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	if req.ExerciseID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少习题 ID"})
		return
	}

	ex, err := h.courseSvc.GetExercise(req.ExerciseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "习题不存在"})
		return
	}

	// 用测试用例评判代码
	result := sandbox.JudgeCodeJSON(req.Language, req.Code, ex.TestCases)

	// 全部通过则记录正确提交
	userID := middleware.GetUserID(c)
	if userID > 0 && result.AllPass {
		h.progressSvc.RestoreHeart(userID)
	}

	c.JSON(http.StatusOK, result)
}

// CompleteLesson 完成课程，发放 XP
func (h *CodeHandler) CompleteLesson(c *gin.Context) {
	lessonID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程 ID 无效"})
		return
	}

	userID := middleware.GetUserID(c)
	score := 100 // 默认满分

	xp, err := h.progressSvc.CompleteLesson(userID, uint(lessonID), score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存进度失败"})
		return
	}

	hearts, _ := h.progressSvc.RestoreHeart(userID)

	c.JSON(http.StatusOK, gin.H{
		"xp_earned": xp,
		"hearts":    hearts,
		"message":   "课程完成！",
	})
}
