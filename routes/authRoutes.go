package routes

import (
	"oosa/internal/middleware"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) *gin.Engine {
	authRepo := repository.AuthRepository{}
	forgetPasswordRepo := repository.ForgetPasswordRepository{}

	register := r.Group("/register")
	{
		register.POST("", authRepo.RegisterEmail)
	}

	main := r.Group("/auth")
	{
		main.POST("/google", authRepo.AuthGoogle)
		main.GET("/line", authRepo.AuthLine)
		main.POST("/facebook", authRepo.AuthFacebook)
		main.POST("/email", authRepo.AuthEmail)
	}

	auth := main.Group("/", middleware.AuthMiddleware())
	{
		auth.GET("/", authRepo.Auth)
		auth.PUT("/", authRepo.AuthUpdate)
		auth.PUT("/change-password", authRepo.AuthUpdatePassword)
		auth.POST("/update-profile-picture", authRepo.AuthUpdateProfilePicture)
	}

	forgetPassword := r.Group("/forget-password")
	{
		forgetPassword.POST("", forgetPasswordRepo.Create)
		forgetPassword.POST("/:token", forgetPasswordRepo.Update)
	}

	return r
}
