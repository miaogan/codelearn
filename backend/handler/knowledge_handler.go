package handler

import (
	"log"
	"net/http"

	"codelearn/eino"

	"github.com/gin-gonic/gin"
)

type KnowledgeHandler struct {
	rag *eino.KnowledgeRAG
}

func NewKnowledgeHandler(rag *eino.KnowledgeRAG) *KnowledgeHandler {
	return &KnowledgeHandler{rag: rag}
}

// POST /api/knowledge/ask
func (h *KnowledgeHandler) Ask(c *gin.Context) {
	var req struct {
		Question string `json:"question"`
		Language string `json:"language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "问题不能为空"})
		return
	}

	log.Printf("[KnowledgeAPI] ask: question=%s language=%s", req.Question[:min(50, len(req.Question))], req.Language)

	result, err := h.rag.Ask(c.Request.Context(), req.Question, req.Language)
	if err != nil {
		log.Printf("[KnowledgeAPI] ask 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
