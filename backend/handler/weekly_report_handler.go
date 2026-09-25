package handler

import (
	"net/http"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

// WeeklyReportHandler 学习周报（U13）
type WeeklyReportHandler struct {
	svc *service.WeeklyReportService
}

func NewWeeklyReportHandler(svc *service.WeeklyReportService) *WeeklyReportHandler {
	return &WeeklyReportHandler{svc: svc}
}

// Get 生成过去 7 天学习周报
func (h *WeeklyReportHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)
	report, err := h.svc.Report(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成周报失败"})
		return
	}
	c.JSON(http.StatusOK, report)
}
