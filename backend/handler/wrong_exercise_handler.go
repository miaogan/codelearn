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
	srsSvc   *service.SRSReviewService
}

func NewWrongExerciseHandler(wrongSvc *service.WrongExerciseService, srsSvc *service.SRSReviewService) *WrongExerciseHandler {
	return &WrongExerciseHandler{wrongSvc: wrongSvc, srsSvc: srsSvc}
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

// TodayReviews 今日到期的 SRS 复习题
func (h *WrongExerciseHandler) TodayReviews(c *gin.Context) {
	userID := middleware.GetUserID(c)

	items, err := h.srsSvc.TodayReviews(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询复习题失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reviews": items, "due_count": len(items)})
}

type srsReviewReq struct {
	Correct bool `json:"correct"`
}

// SubmitReview 提交一次 SRS 复习结果（记得/忘了）
func (h *WrongExerciseHandler) SubmitReview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	wrongID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "复习题 ID 无效"})
		return
	}

	var req srsReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	if err := h.srsSvc.SubmitReview(userID, uint(wrongID), req.Correct); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "复习题不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "复习已记录"})
}
