package handler

import (
	"net/http"
	"time"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

type LeaderboardHandler struct {
	svc *service.LeaderboardService
}

func NewLeaderboardHandler(svc *service.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{svc: svc}
}

// Weekly 本周排行榜 + 我的排名
func (h *LeaderboardHandler) Weekly(c *gin.Context) {
	userID := middleware.GetUserID(c)

	entries, err := h.svc.WeeklyLeaderboard(20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排行榜失败"})
		return
	}
	me, _ := h.svc.MyRank(userID)

	c.JSON(http.StatusOK, gin.H{"entries": entries, "me": me})
}

// Calendar 学习日历（按月查询每天 XP）
func (h *LeaderboardHandler) Calendar(c *gin.Context) {
	userID := middleware.GetUserID(c)

	month := c.Query("month") // YYYY-MM
	now := time.Now()
	year, m := now.Year(), int(now.Month())
	if month != "" {
		if t, err := time.Parse("2006-01", month); err == nil {
			year, m = t.Year(), int(t.Month())
		}
	}

	days, err := h.svc.Calendar(userID, year, m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学习日历失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"year": year, "month": m, "days": days})
}
