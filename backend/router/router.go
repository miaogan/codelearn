package router

import (
	"codelearn/config"
	"codelearn/handler"
	"codelearn/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, auth *handler.AuthHandler, course *handler.CourseHandler, exercise *handler.ExerciseHandler, code *handler.CodeHandler, progress *handler.ProgressHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		api.POST("/auth/register", auth.Register)
		api.POST("/auth/login", auth.Login)

		authed := api.Group("")
		authed.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			authed.GET("/courses", course.ListCourses)
			authed.GET("/courses/:id", course.GetLearningPath)
			authed.GET("/lessons/:id", course.GetLesson)
			authed.GET("/lessons/:id/exercises", course.GetExercises)

			authed.POST("/exercises/:id/submit", exercise.Submit)
			authed.POST("/lessons/:id/generate", exercise.Generate)
			authed.POST("/exercises/hint", exercise.Hint)
			authed.GET("/exercises/:id/template", exercise.GetTestCase)

			authed.POST("/code/run", code.Run)
			authed.POST("/code/judge", code.Judge)
			authed.POST("/lessons/:id/complete", code.CompleteLesson)

			authed.GET("/users/me/stats", progress.Stats)
			authed.GET("/users/me/progress", progress.ListProgress)
		}
	}

	r.Static("/static", "../frontend/dist")
	r.NoRoute(func(c *gin.Context) {
		c.File("../frontend/dist/index.html")
	})

	return r
}
