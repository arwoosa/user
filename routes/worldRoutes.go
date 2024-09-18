package routes

import (
	"oosa/internal/middleware"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func WorldRoutes(r *gin.Engine) *gin.Engine {
	repository := repository.StatisticsRepository{}
	main := r.Group("/world", middleware.AuthMiddleware())
	{
		main.GET("statistics", repository.Retrieve)
	}

	return r
}
