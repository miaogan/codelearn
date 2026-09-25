package handler

import (
	"net/http"
	"strconv"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

// PredictionHandler 分数预测 + 考试日历（U7）
type PredictionHandler struct {
	svc *service.PredictionService
}

func NewPredictionHandler(svc *service.PredictionService) *PredictionHandler {
	return &PredictionHandler{svc: svc}
}

// Predict 预测指定课程认证考试通过概率
func (h *PredictionHandler) Predict(c *gin.Context) {
	userID := middleware.GetUserID(c)
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程 ID 无效"})
		return
	}
	pred, err := h.svc.Predict(userID, uint(courseID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}
	c.JSON(http.StatusOK, pred)
}
