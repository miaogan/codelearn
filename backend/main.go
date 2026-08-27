package main

import (
	"fmt"
	"log"

	"codelearn/config"
	"codelearn/eino"
	"codelearn/handler"
	"codelearn/model"
	"codelearn/repository"
	"codelearn/router"
	"codelearn/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	// 初始化数据库（SQLite，纯 Go 驱动，无需 CGO）
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(
		&model.User{}, &model.Course{}, &model.Unit{},
		&model.Lesson{}, &model.Exercise{}, &model.UserProgress{},
		&model.Submission{}, &model.WrongExercise{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化各层
	repo := repository.New(db)
	courseSvc := service.NewCourseService(repo)
	progressSvc := service.NewProgressService(repo, cfg.XPPerLesson, cfg.XPPerExercise, cfg.MaxHearts)
	wrongSvc := service.NewWrongExerciseService(repo)
	generator := eino.NewExerciseGenerator(cfg)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(repo, cfg.JWTSecret, cfg.MaxHearts)
	courseHandler := handler.NewCourseHandler(courseSvc)
	exerciseHandler := handler.NewExerciseHandler(courseSvc, progressSvc, generator)
	codeHandler := handler.NewCodeHandler(courseSvc, progressSvc)
	progressHandler := handler.NewProgressHandler(progressSvc)
	wrongHandler := handler.NewWrongExerciseHandler(wrongSvc)

	// 初始化路由
	r := router.Setup(cfg, authHandler, courseHandler, exerciseHandler, codeHandler, progressHandler, wrongHandler)

	// 种子数据
	if err := seedData(repo); err != nil {
		log.Printf("种子数据初始化警告: %v", err)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 CodeLearn 服务启动: http://localhost:%s", cfg.Port)
	log.Printf("   API 文档: http://localhost:%s/api/courses", cfg.Port)
	if cfg.LLMEnabled() {
		log.Printf("   AI 习题生成: 已启用 (模型: %s)", cfg.LLMModel)
	} else {
		log.Printf("   AI 习题生成: 未启用 (请设置 LLM_API_KEY 环境变量)")
	}

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
