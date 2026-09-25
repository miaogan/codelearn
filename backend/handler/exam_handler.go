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
	certSvc     *service.CertificateService
	analytics   *service.AnalyticsService
}

func NewExamHandler(examSvc *service.ExamService, progressSvc *service.ProgressService, certSvc *service.CertificateService, analytics *service.AnalyticsService) *ExamHandler {
	return &ExamHandler{examSvc: examSvc, progressSvc: progressSvc, certSvc: certSvc, analytics: analytics}
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

// StartCertExam 组装并开始课程认证考试（前置：课程全部单元考试通过）
func (h *ExamHandler) StartCertExam(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程 ID 无效"})
		return
	}
	userID := middleware.GetUserID(c)

	exam, err := h.examSvc.AssembleCertExam(uint(courseID), userID)
	if err != nil {
		switch err {
		case service.ErrCertUnitsNotPassed:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case service.ErrNoQuestions:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			log.Printf("[StartCertExam] courseID=%d err=%v", courseID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建认证考试失败"})
		}
		return
	}
	c.JSON(http.StatusOK, exam)
}

// CertStatus 查询课程认证状态（是否具备考试资格、是否已持证）
func (h *ExamHandler) CertStatus(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程 ID 无效"})
		return
	}
	userID := middleware.GetUserID(c)

	status, err := h.examSvc.CertStatus(uint(courseID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询认证状态失败"})
		return
	}
	c.JSON(http.StatusOK, status)
}

type unitExamSubmitReq struct {
	Answers     []service.ExamAnswerItem `json:"answers" binding:"required"`
	DurationSec int                      `json:"duration_sec"`
	TabSwitches int                      `json:"tab_switches"` // 防作弊：考试期间切屏次数
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

	report, err := h.examSvc.SubmitExam(userID, uint(examID), req.Answers, req.DurationSec, req.TabSwitches)
	if err != nil {
		switch err {
		case service.ErrAlreadyPassed, service.ErrRetryCoolDown:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			log.Printf("[SubmitExam] examID=%d err=%v", examID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "提交考试失败"})
		}
		return
	}

	if report.Passed {
		h.progressSvc.RecordExamXP(userID, service.ExamXP)
		// 认证考试通过 → 颁发能力认证证书（已持有则忽略）
		if report.ExamType == service.ExamTypeCert {
			if cert, err := h.certSvc.IssueCertificate(userID, report.CourseID, report.Score); err == nil {
				report.Certificate = cert
				h.analytics.Record(service.EventCertIssued, userID, report.CourseID)
			} else if err != service.ErrAlreadyCertified {
				log.Printf("[SubmitExam] issue certificate err=%v", err)
			}
		} else {
			h.analytics.Record(service.EventUnitExamPassed, userID, report.CourseID)
		}
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
