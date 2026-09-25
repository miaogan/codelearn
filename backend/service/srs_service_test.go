package service

import (
	"testing"

	"codelearn/model"
)

func seedExercise(db interface {
	Create(v interface{}) *gormDB
}, lessonID uint) *model.Exercise {
	ex := &model.Exercise{
		LessonID:    lessonID,
		Type:        "choice",
		Question:    "测试题目",
		Options:     "A|B|C|D",
		Answer:      "A",
		Explanation: "解析",
		Difficulty:  "easy",
	}
	createExercise(db, ex)
	return ex
}
