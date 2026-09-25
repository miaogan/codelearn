package handler

import (
	"net/http"
	"strconv"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

type SkillHandler struct {
	svc *service.SkillMapService
}

func NewSkillHandler(svc *service.SkillMapService) *SkillHandler {
	return &SkillHandler{svc: svc}
}

// SkillMap 课程能力图谱
func (h *SkillHandler) SkillMap(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程 ID 无效"})
		return
	}
	userID := middleware.GetUserID(c)

	skill, err := h.svc.CourseSkillMap(uint(courseID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取能力图谱失败"})
		return
	}
	c.JSON(http.StatusOK, skill)
}

// ReviewRecommendations 薄弱章节复习推荐
func (h *SkillHandler) ReviewRecommendations(c *gin.Context) {
	userID := middleware.GetUserID(c)

	recs, err := h.svc.ReviewRecommendations(userID, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取复习推荐失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recommendations": recs})
}
