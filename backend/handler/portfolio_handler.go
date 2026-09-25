package handler

import (
	"net/http"

	"codelearn/service"

	"github.com/gin-gonic/gin"
)

// PortfolioHandler 数字能力档案（U10，公开路由）
type PortfolioHandler struct {
	svc *service.PortfolioService
}

func NewPortfolioHandler(svc *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{svc: svc}
}

// Get 公开能力档案页数据
func (h *PortfolioHandler) Get(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名为空"})
		return
	}
	pf, err := h.svc.Get(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, pf)
}
