package handler

import (
	"log"
	"net/http"

	"codelearn/eino"
	"codelearn/sandbox"

	"github.com/gin-gonic/gin"
)

type TutorHandler struct {
	agent *eino.TutorAgent
}

func NewTutorHandler(agent *eino.TutorAgent) *TutorHandler {
	return &TutorHandler{agent: agent}
}

// POST /api/tutor/debug
func (h *TutorHandler) Debug(c *gin.Context) {
	var req eino.TutorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	log.Printf("[TutorAPI] debug: language=%s code_len=%d", req.Language, len(req.Code))

	result, err := h.agent.Debug(c.Request.Context(), req)
	if err != nil {
		log.Printf("[TutorAPI] debug 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// POST /api/tutor/chat
func (h *TutorHandler) Chat(c *gin.Context) {
	var req struct {
		Messages []eino.TutorChatMessage `json:"messages"`
		Code     string                  `json:"code"`
		Language string                  `json:"language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	log.Printf("[TutorAPI] chat: %d messages", len(req.Messages))

	reply, err := h.agent.Chat(c.Request.Context(), req.Messages, req.Code, req.Language)
	if err != nil {
		log.Printf("[TutorAPI] chat 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

// POST /api/tutor/review
func (h *TutorHandler) Review(c *gin.Context) {
	var req struct {
		Code     string `json:"code"`
		Language string `json:"language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	log.Printf("[TutorAPI] review: language=%s", req.Language)

	result, err := h.agent.AnalyzeCode(c.Request.Context(), req.Code, req.Language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": result})
}

// POST /api/tutor/run
func (h *TutorHandler) Run(c *gin.Context) {
	var req struct {
		Language string `json:"language"`
		Code     string `json:"code"`
		Input    string `json:"input"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	log.Printf("[TutorAPI] run: language=%s", req.Language)
	result := h.agent.RunCodeForTutor(req.Language, req.Code, req.Input)
	c.JSON(http.StatusOK, result)
}

// GET /api/tutor/sandbox-languages
func (h *TutorHandler) SandboxLanguages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"languages": []string{"go", "python"},
		"info":     "沙箱支持的语言",
	})
	_ = sandbox.RunResult{} // ensure package usage
}
