package handler

import (
	"net/http"
	"strconv"

	"codelearn/middleware"
	"codelearn/service"

	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	courseSvc *service.CourseService
}

func NewCourseHandler(courseSvc *service.CourseService) *CourseHandler {
	return &CourseHandler{courseSvc: courseSvc}
}

func (h *CourseHandler) ListCourses(c *gin.Context) {
	courses, err := h.courseSvc.ListCourses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取课程列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"courses": courses})
}

func (h *CourseHandler) GetLearningPath(c *gin.Context) {
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程 ID 无效"})
		return
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	path, err := h.courseSvc.GetLearningPath(uint(courseID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}
	c.JSON(http.StatusOK, path)
}

func (h *CourseHandler) GetLesson(c *gin.Context) {
	lessonID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程 ID 无效"})
		return
	}
	lesson, err := h.courseSvc.GetLesson(uint(lessonID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "课程不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"lesson": lesson})
}

func (h *CourseHandler) GetExercises(c *gin.Context) {
	lessonID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "课程 ID 无效"})
		return
	}
	exercises, err := h.courseSvc.GetExercises(uint(lessonID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "获取习题失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exercises": exercises})
}
