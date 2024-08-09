package routes

import (
	"oosa/internal/middleware"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func OosaDailyRoutes(r *gin.Engine) *gin.Engine {
	repository := repository.OosaDailyRepository{}
	main := r.Group("/oosa-daily", middleware.AuthMiddleware())
	{
		main.POST("", repository.Watched)
	}
	return r
}
