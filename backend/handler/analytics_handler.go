package handler

import (
	"net/http"
	"strconv"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	svc       *service.AnalyticsService
	adminToken string
}

func NewAnalyticsHandler(svc *service.AnalyticsService, adminToken string) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, adminToken: adminToken}
}

// Funnel 学习漏斗汇总（管理端，需 X-Admin-Token）
func (h *AnalyticsHandler) Funnel(c *gin.Context) {
	if h.adminToken == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "未配置管理端令牌"})
		return
	}
	if c.GetHeader("X-Admin-Token") != h.adminToken {
		c.JSON(http.StatusForbidden, gin.H{"error": "管理端令牌无效"})
		return
	}

	days := 30
	if v := c.Query("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			days = n
		}
	}

	c.JSON(http.StatusOK, gin.H{"funnel": h.svc.Funnel(days)})
}

type feedbackReq struct {
	Category string `json:"category"`
	Content  string `json:"content" binding:"required"`
	Contact  string `json:"contact"`
}

// SubmitFeedback 提交用户反馈（问题上报）
func (h *AnalyticsHandler) SubmitFeedback(c *gin.Context) {
	var req feedbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	if req.Category == "" {
		req.Category = "other"
	}
	userID := middleware.GetUserID(c)

	if err := h.svc.SubmitFeedback(userID, req.Category, req.Content, req.Contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "感谢反馈，我们会尽快处理"})
}
