package handler

import (
	"log"
	"net/http"
	"strconv"

	domainmodel "codelearn/model"
	"codelearn/eino"

	"github.com/gin-gonic/gin"
)

type AdaptiveHandler struct {
	advisor *eino.AdaptiveAdvisor
	repo    interface {
		ListWrongExercises(userID uint, onlyUnmastered bool) ([]domainmodel.WrongExercise, error)
		ListExercisesByIDs(ids []uint) ([]domainmodel.Exercise, error)
		GetCourse(id uint) (*domainmodel.Course, error)
		CreateExercises(exercises []domainmodel.Exercise) error
	}
}

func NewAdaptiveHandler(advisor *eino.AdaptiveAdvisor, repo interface {
	ListWrongExercises(userID uint, onlyUnmastered bool) ([]domainmodel.WrongExercise, error)
	ListExercisesByIDs(ids []uint) ([]domainmodel.Exercise, error)
	GetCourse(id uint) (*domainmodel.Course, error)
	CreateExercises(exercises []domainmodel.Exercise) error
}) *AdaptiveHandler {
	return &AdaptiveHandler{advisor: advisor, repo: repo}
}

// GET /api/adaptive/recommend?course_id=1
func (h *AdaptiveHandler) Recommend(c *gin.Context) {
	userID := c.GetUint("userID")
	courseID, _ := strconv.Atoi(c.Query("course_id"))
	language := c.DefaultQuery("language", "Go")

	log.Printf("[AdaptiveAPI] recommend: userID=%d courseID=%d", userID, courseID)

	// 1. 获取用户未掌握的错题
	wrongList, err := h.repo.ListWrongExercises(userID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取错题失败"})
		return
	}

	if len(wrongList) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"weak_points": []interface{}{},
			"summary":     "暂无错题数据，继续学习以积累数据",
			"exercises":   []interface{}{},
		})
		return
	}

	// 2. 获取错题详情
	exIDs := make([]uint, len(wrongList))
	for i, w := range wrongList {
		exIDs[i] = w.ExerciseID
	}
	exercises, err := h.repo.ListExercisesByIDs(exIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取错题详情失败"})
		return
	}

	// 3. Eino Chain: 分析 → 生成
	rec, err := h.advisor.Recommend(c.Request.Context(), exercises, language)
	if err != nil {
		log.Printf("[AdaptiveAPI] AI 分析失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 保存生成的练习题
	if len(rec.Exercises) > 0 {
		for i := range rec.Exercises {
			rec.Exercises[i].LessonID = 0 // 自适应练习不属于特定课程
		}
		_ = h.repo.CreateExercises(rec.Exercises)
		// 清除 answer 字段不返回给前端
		for i := range rec.Exercises {
			rec.Exercises[i].Answer = ""
		}
	}

	c.JSON(http.StatusOK, rec)
}
