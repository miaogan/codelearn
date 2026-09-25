package handler

import (
	"net/http"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

// AchievementHandler 成就徽章与段位（U12）
type AchievementHandler struct {
	svc *service.AchievementService
}

func NewAchievementHandler(svc *service.AchievementService) *AchievementHandler {
	return &AchievementHandler{svc: svc}
}

// Summary 查询成就墙与当前段位
func (h *AchievementHandler) Summary(c *gin.Context) {
	userID := middleware.GetUserID(c)
	s, err := h.svc.Summary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询成就失败"})
		return
	}
	c.JSON(http.StatusOK, s)
}
