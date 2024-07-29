package routes

import (
	"oosa/internal/middleware"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func StatisticsRoutes(r *gin.Engine) *gin.Engine {
	repository := repository.StatisticsRepository{}
	main := r.Group("/statistics", middleware.AuthMiddleware())
	{
		main.GET("", repository.Retrieve)
	}

	return r
}
