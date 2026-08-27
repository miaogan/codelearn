package handler

import (
	"log"
	"net/http"
	"strconv"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

type WrongExerciseHandler struct {
	wrongSvc *service.WrongExerciseService
}

func NewWrongExerciseHandler(wrongSvc *service.WrongExerciseService) *WrongExerciseHandler {
	return &WrongExerciseHandler{wrongSvc: wrongSvc}
}

// List 获取错题列表
func (h *WrongExerciseHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	onlyUnmastered := c.Query("unmastered") == "1"
	log.Printf("[WrongList] userID=%d unmastered=%v", userID, onlyUnmastered)

	items, err := h.wrongSvc.ListWrongExercises(userID, onlyUnmastered)
	if err != nil {
		log.Printf("[WrongList] 查询失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询错题失败"})
		return
	}

	log.Printf("[WrongList] 返回 %d 条错题", len(items))
	c.JSON(http.StatusOK, gin.H{"wrong_exercises": items, "total": len(items)})
}

// MarkMastered 标记错题已掌握
func (h *WrongExerciseHandler) MarkMastered(c *gin.Context) {
	userID := middleware.GetUserID(c)
	exerciseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "习题 ID 无效"})
		return
	}

	log.Printf("[WrongMastered] userID=%d exerciseID=%d", userID, exerciseID)

	if err := h.wrongSvc.MarkMastered(userID, uint(exerciseID)); err != nil {
		log.Printf("[WrongMastered] 标记失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "标记失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已标记为掌握"})
}

// Count 获取未掌握错题数量
func (h *WrongExerciseHandler) Count(c *gin.Context) {
	userID := middleware.GetUserID(c)

	count, err := h.wrongSvc.CountWrong(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
