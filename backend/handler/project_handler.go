package handler

import (
	"log"
	"net/http"
	"strconv"

	"codelearn/middleware"
	"codelearn/model"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

// ProjectHandler 项目实战工坊（U9）
type ProjectHandler struct {
	svc         *service.ProjectService
	achievement *service.AchievementService
}

func NewProjectHandler(svc *service.ProjectService, achievement *service.AchievementService) *ProjectHandler {
	return &ProjectHandler{svc: svc, achievement: achievement}
}

// Templates 项目模板列表
func (h *ProjectHandler) Templates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"templates": h.svc.ListTemplates()})
}

type createProjectReq struct {
	CourseID    uint   `json:"course_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Language    string `json:"language"`
	MainFile    string `json:"main_file"`
}

// Create 创建项目
func (h *ProjectHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req createProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	p, err := h.svc.Create(userID, req.CourseID, req.Title, req.Description, req.Language, req.MainFile)
	if err != nil {
		log.Printf("[ProjectCreate] 创建失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建项目失败"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// List 项目列表
func (h *ProjectHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ps, err := h.svc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询项目失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"projects": ps})
}

// Get 项目详情（含文件）
func (h *ProjectHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目 ID 无效"})
		return
	}
	p, files, err := h.svc.Get(userID, uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"project": p, "files": files})
}

type saveFilesReq struct {
	MainFile string              `json:"main_file"`
	Files    []model.ProjectFile `json:"files"`
}

// SaveFiles 保存项目文件
func (h *ProjectHandler) SaveFiles(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目 ID 无效"})
		return
	}
	var req saveFilesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	if err := h.svc.SaveFiles(userID, uint(projectID), req.MainFile, req.Files); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已保存"})
}

// Run 运行项目
func (h *ProjectHandler) Run(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目 ID 无效"})
		return
	}
	res, err := h.svc.Run(userID, uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Complete 标记项目完成
func (h *ProjectHandler) Complete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目 ID 无效"})
		return
	}
	p, err := h.svc.Complete(userID, uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	h.achievement.OnProjectCompleted(userID)
	c.JSON(http.StatusOK, gin.H{"project": p, "message": "项目已完成"})
}

// Delete 删除项目
func (h *ProjectHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目 ID 无效"})
		return
	}
	if err := h.svc.Delete(userID, uint(projectID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
