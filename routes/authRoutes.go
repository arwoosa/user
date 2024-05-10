package routes

import (
	"oosa/internal/middleware"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) *gin.Engine {
	repository := repository.AuthRepository{}

	register := r.Group("/register")
	{
		register.POST("", repository.RegisterEmail)
	}

	main := r.Group("/auth")
	{
		main.GET("/", middleware.AuthMiddleware(), repository.Auth)
		main.POST("/google", repository.AuthGoogle)
		main.GET("/line", repository.AuthLine)
		main.POST("/facebook", repository.AuthFacebook)
		main.POST("/email", repository.AuthEmail)
	}

	return r
}
