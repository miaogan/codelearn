package server

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

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// New 初始化数据库、服务与路由，返回可运行的 Gin 引擎。
// Web 版与桌面版共用此入口，保证两种形态行为一致。
func New(cfg *config.Config) (*gin.Engine, error) {
	// 初始化数据库（SQLite，纯 Go 驱动，无需 CGO）
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(
		&model.User{}, &model.Course{}, &model.Unit{},
		&model.Lesson{}, &model.Exercise{}, &model.UserProgress{},
		&model.Submission{}, &model.WrongExercise{},
		&model.XPEvent{}, &model.Exam{}, &model.ExamQuestion{},
		&model.ExamSubmission{}, &model.Certificate{},
		&model.AnalyticsEvent{}, &model.UserFeedback{},
		&model.Achievement{},
		&model.Project{}, &model.ProjectFile{},
	); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	repo := repository.New(db)

	// 从 XP 流水对账用户总 XP（保证顶栏与今日进度同口径）
	if err := repo.RebuildUserXP(); err != nil {
		log.Printf("XP 对账警告: %v", err)
	}

	courseSvc := service.NewCourseService(repo)
	progressSvc := service.NewProgressService(repo, cfg.XPPerLesson, cfg.XPPerExercise, cfg.MaxHearts)
	wrongSvc := service.NewWrongExerciseService(repo)
	srsSvc := service.NewSRSReviewService(repo)
	achievementSvc := service.NewAchievementService(repo)
	weeklySvc := service.NewWeeklyReportService(repo)
	projectSvc := service.NewProjectService(repo)
	portfolioSvc := service.NewPortfolioService(repo)
	predictionSvc := service.NewPredictionService(repo)
	planAdvisor := eino.NewPlanAdvisor(cfg)
	studyPlanSvc := service.NewStudyPlanService(repo, planAdvisor)
	examSvc := service.NewExamService(repo)
	leaderboardSvc := service.NewLeaderboardService(repo)
	certSvc := service.NewCertificateService(repo)
	skillSvc := service.NewSkillMapService(repo)
	analyticsSvc := service.NewAnalyticsService(repo)
	generator := eino.NewExerciseGenerator(cfg)

	// Eino 组件初始化
	adaptiveAdvisor := eino.NewAdaptiveAdvisor(cfg)
	tutorAgent := eino.NewTutorAgent(cfg)
	knowledgeRAG := eino.NewKnowledgeRAG(cfg, repo)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(repo, cfg.JWTSecret, cfg.MaxHearts, analyticsSvc)
	courseHandler := handler.NewCourseHandler(courseSvc)
	exerciseHandler := handler.NewExerciseHandler(courseSvc, progressSvc, generator, achievementSvc)
	codeHandler := handler.NewCodeHandler(courseSvc, progressSvc, analyticsSvc, achievementSvc)
	progressHandler := handler.NewProgressHandler(progressSvc)
	wrongHandler := handler.NewWrongExerciseHandler(wrongSvc, srsSvc)
	adaptiveHandler := handler.NewAdaptiveHandler(adaptiveAdvisor, repo)
	tutorHandler := handler.NewTutorHandler(tutorAgent)
	knowledgeHandler := handler.NewKnowledgeHandler(knowledgeRAG)
	examHandler := handler.NewExamHandler(examSvc, progressSvc, certSvc, analyticsSvc, achievementSvc)
	leaderboardHandler := handler.NewLeaderboardHandler(leaderboardSvc)
	certHandler := handler.NewCertificateHandler(certSvc)
	skillHandler := handler.NewSkillHandler(skillSvc)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc, cfg.AdminToken)
	achievementHandler := handler.NewAchievementHandler(achievementSvc)
	weeklyHandler := handler.NewWeeklyReportHandler(weeklySvc)
	projectHandler := handler.NewProjectHandler(projectSvc, achievementSvc)
	portfolioHandler := handler.NewPortfolioHandler(portfolioSvc)
	predictionHandler := handler.NewPredictionHandler(predictionSvc)
	studyPlanHandler := handler.NewStudyPlanHandler(studyPlanSvc)

	// 初始化路由
	r := router.Setup(cfg, authHandler, courseHandler, exerciseHandler, codeHandler, progressHandler, wrongHandler, adaptiveHandler, tutorHandler, knowledgeHandler, examHandler, leaderboardHandler, certHandler, skillHandler, analyticsHandler, achievementHandler, weeklyHandler, projectHandler, portfolioHandler, predictionHandler, studyPlanHandler)

	// 种子数据
	if err := seedData(repo); err != nil {
		log.Printf("种子数据初始化警告: %v", err)
	}

	return r, nil
}
