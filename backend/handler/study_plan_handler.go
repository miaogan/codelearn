package handler

import (
	"net/http"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

// StudyPlanHandler AI 学习计划（U5）
type StudyPlanHandler struct {
	svc *service.StudyPlanService
}

func NewStudyPlanHandler(svc *service.StudyPlanService) *StudyPlanHandler {
	return &StudyPlanHandler{svc: svc}
}

type generatePlanReq struct {
	Goal     string `json:"goal"`
	Language string `json:"language"`
	Weeks    int    `json:"weeks"`
}

// Generate 生成学习计划
func (h *StudyPlanHandler) Generate(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req generatePlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	plan, err := h.svc.Generate(c.Request.Context(), userID, req.Goal, req.Language, req.Weeks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成计划失败"})
		return
	}
	c.JSON(http.StatusOK, plan)
}
