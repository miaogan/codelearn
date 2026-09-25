package handler

import (
	"net/http"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

type ProgressHandler struct {
	progressSvc *service.ProgressService
}

func NewProgressHandler(progressSvc *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{progressSvc: progressSvc}
}

// Stats 获取用户游戏化统计数据（XP、连续打卡、心数、每日目标）
func (h *ProgressHandler) Stats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stats, err := h.progressSvc.GetStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计数据失败"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ListProgress 获取用户所有课程进度
func (h *ProgressHandler) ListProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	progress, err := h.progressSvc.ListProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取进度失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"progress": progress})
}

type dailyGoalReq struct {
	Goal int `json:"goal" binding:"required"`
}

// UpdateDailyGoal 更新每日目标
func (h *ProgressHandler) UpdateDailyGoal(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dailyGoalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	goal, err := h.progressSvc.UpdateDailyGoal(userID, req.Goal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新每日目标失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"daily_goal": goal})
}
