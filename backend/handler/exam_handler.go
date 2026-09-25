package handler

import (
	"log"
	"net/http"
	"strconv"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

type ExamHandler struct {
	examSvc     *service.ExamService
	progressSvc *service.ProgressService
}

func NewExamHandler(examSvc *service.ExamService, progressSvc *service.ProgressService) *ExamHandler {
	return &ExamHandler{examSvc: examSvc, progressSvc: progressSvc}
}

// StartUnitExam 组装并开始单元考试（前置：单元内课时全部完成）
func (h *ExamHandler) StartUnitExam(c *gin.Context) {
	unitID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "单元 ID 无效"})
		return
	}
	userID := middleware.GetUserID(c)

	exam, err := h.examSvc.AssembleUnitExam(uint(unitID), userID)
	if err != nil {
		switch err {
		case service.ErrUnitNotCompleted:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case service.ErrNoQuestions:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			log.Printf("[StartUnitExam] unitID=%d err=%v", unitID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建考试失败"})
		}
		return
	}
	c.JSON(http.StatusOK, exam)
}

type unitExamSubmitReq struct {
	Answers     []service.ExamAnswerItem `json:"answers" binding:"required"`
	DurationSec int                      `json:"duration_sec"`
}

// SubmitExam 提交考试，返回考试报告；通过后发放考试 XP
func (h *ExamHandler) SubmitExam(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "考试 ID 无效"})
		return
	}
	var req unitExamSubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	userID := middleware.GetUserID(c)

	report, err := h.examSvc.SubmitExam(userID, uint(examID), req.Answers, req.DurationSec)
	if err != nil {
		log.Printf("[SubmitExam] examID=%d err=%v", examID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交考试失败"})
		return
	}

	if report.Passed {
		h.progressSvc.RecordExamXP(userID, service.ExamXP)
	}
	c.JSON(http.StatusOK, report)
}

// GetReport 获取用户最近一次考试报告
func (h *ExamHandler) GetReport(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "考试 ID 无效"})
		return
	}
	userID := middleware.GetUserID(c)

	report, err := h.examSvc.GetReport(userID, uint(examID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "暂无考试记录"})
		return
	}
	c.JSON(http.StatusOK, report)
}
