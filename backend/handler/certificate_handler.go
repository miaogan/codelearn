package handler

import (
	"net/http"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

type CertificateHandler struct {
	svc *service.CertificateService
}

func NewCertificateHandler(svc *service.CertificateService) *CertificateHandler {
	return &CertificateHandler{svc: svc}
}

// List 我的证书列表
func (h *CertificateHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	list, err := h.svc.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取证书失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"certificates": list})
}

type verifyReq struct {
	CertNo string `json:"cert_no" binding:"required"`
}

// Verify 证书在线验证（公开接口，无需登录）
func (h *CertificateHandler) Verify(c *gin.Context) {
	var req verifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	cert, err := h.svc.Verify(req.CertNo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": false, "message": "未找到该证书"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": true, "certificate": cert})
}
