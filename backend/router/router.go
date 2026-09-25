package router

import (
	"time"

	"codelearn/config"
	"codelearn/handler"
	"codelearn/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, auth *handler.AuthHandler, course *handler.CourseHandler, exercise *handler.ExerciseHandler, code *handler.CodeHandler, progress *handler.ProgressHandler, wrong *handler.WrongExerciseHandler, adaptive *handler.AdaptiveHandler, tutor *handler.TutorHandler, knowledge *handler.KnowledgeHandler, exam *handler.ExamHandler, leaderboard *handler.LeaderboardHandler, cert *handler.CertificateHandler, skill *handler.SkillHandler, analytics *handler.AnalyticsHandler, achievement *handler.AchievementHandler, weekly *handler.WeeklyReportHandler, project *handler.ProjectHandler, portfolio *handler.PortfolioHandler, prediction *handler.PredictionHandler, studyPlan *handler.StudyPlanHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Admin-Token")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		// 认证接口按 IP 限流，防撞库/批量注册
		api.POST("/auth/register", middleware.RateLimit(5, time.Minute), auth.Register)
		api.POST("/auth/login", middleware.RateLimit(10, time.Minute), auth.Login)
		api.POST("/certificates/verify", cert.Verify)

		// 公开能力档案（U10）
		api.GET("/me/:username", portfolio.Get)

		// 管理端：学习漏斗（需 X-Admin-Token）
		api.GET("/admin/funnel", analytics.Funnel)

		authed := api.Group("")
		authed.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			authed.GET("/courses", course.ListCourses)
			authed.GET("/courses/:id", course.GetLearningPath)
			authed.GET("/lessons/:id", course.GetLesson)
			authed.GET("/lessons/:id/exercises", course.GetExercises)

			authed.POST("/exercises/:id/submit", exercise.Submit)
			authed.POST("/exercises/exam-submit", exercise.ExamSubmit)
			authed.POST("/lessons/:id/generate", exercise.Generate)
			authed.POST("/exercises/hint", exercise.Hint)
			authed.GET("/exercises/:id/template", exercise.GetTestCase)

			authed.POST("/code/run", code.Run)
			authed.POST("/code/judge", code.Judge)
			authed.POST("/lessons/:id/complete", code.CompleteLesson)

			// 用户反馈
			authed.POST("/feedback", analytics.SubmitFeedback)

			authed.GET("/users/me/stats", progress.Stats)
			authed.GET("/users/me/progress", progress.ListProgress)
			authed.PUT("/users/me/daily-goal", progress.UpdateDailyGoal)
			authed.GET("/users/me/calendar", leaderboard.Calendar)

			authed.GET("/wrong-exercises", wrong.List)
			authed.POST("/wrong-exercises/:id/master", wrong.MarkMastered)
			authed.GET("/wrong-exercises/count", wrong.Count)

			// SRS 间隔重复复习（U3）
			authed.GET("/users/me/srs/reviews", wrong.TodayReviews)
			authed.POST("/wrong-exercises/:id/srs-review", wrong.SubmitReview)

			// 成就徽章与段位（U12）
			authed.GET("/users/me/achievements", achievement.Summary)

			// 学习周报（U13）
			authed.GET("/users/me/weekly-report", weekly.Get)

			// 项目实战工坊（U9）
			authed.GET("/projects/templates", project.Templates)
			authed.GET("/projects", project.List)
			authed.POST("/projects", project.Create)
			authed.GET("/projects/:id", project.Get)
			authed.PUT("/projects/:id/files", project.SaveFiles)
			authed.POST("/projects/:id/run", project.Run)
			authed.POST("/projects/:id/complete", project.Complete)
			authed.DELETE("/projects/:id", project.Delete)

			// 考试（单元考试 / 认证考试）
			authed.POST("/units/:id/exam", exam.StartUnitExam)
			authed.POST("/exams/:id/submit", exam.SubmitExam)
			authed.GET("/exams/:id/report", exam.GetReport)
			authed.POST("/courses/:id/cert-exam", exam.StartCertExam)
			authed.GET("/courses/:id/cert-status", exam.CertStatus)

			// 分数预测 + 考试日历（U7）
			authed.GET("/courses/:id/prediction", prediction.Predict)

			// AI 学习计划（U5）
			authed.POST("/study-plan/generate", studyPlan.Generate)

			// 排行榜
			authed.GET("/leaderboard", leaderboard.Weekly)

			// 证书
			authed.GET("/certificates", cert.List)

			// 能力图谱 / 复习推荐
			authed.GET("/courses/:id/skill-map", skill.SkillMap)
			authed.GET("/users/me/review-recommendations", skill.ReviewRecommendations)

			// Eino Chain 模式：自适应学习路径
			authed.GET("/adaptive/recommend", adaptive.Recommend)

			// Eino Agent 模式：AI 编程导师
			authed.POST("/tutor/debug", tutor.Debug)
			authed.POST("/tutor/chat", tutor.Chat)
			authed.POST("/tutor/review", tutor.Review)
			authed.POST("/tutor/run", tutor.Run)

			// Eino RAG 模式：知识点问答
			authed.POST("/knowledge/ask", knowledge.Ask)
		}
	}

	r.Static("/static", cfg.FrontendDir)
	r.Static("/assets", cfg.FrontendDir+"/assets")
	r.NoRoute(func(c *gin.Context) {
		c.File(cfg.FrontendDir + "/index.html")
	})

	return r
}
