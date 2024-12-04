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

	binding := main.Group("/bind", middleware.AuthMiddleware())
	{
		binding.POST("/facebook", authRepo.BindFacebook)
	}

	auth := main.Group("", middleware.AuthMiddleware())
	{
		auth.GET("", authRepo.Auth)
		auth.PUT("", authRepo.AuthUpdate)
		auth.PUT("/take-me", authRepo.AuthUpdateTakeMe)
		auth.PUT("/change-password", authRepo.AuthUpdatePassword)
		auth.POST("/profile-picture", authRepo.AuthUpdateProfilePicture)
		auth.POST("/avatar", authRepo.AuthUpdateAvatar)
	}

	usersetting := main.Group("/setting", middleware.AuthMiddleware())
	{
		usersetting.GET("", authRepo.RetrieveUserSettings)
		usersetting.PUT("", authRepo.UpdateUserSettings)
	}

	badges := main.Group("/badges", middleware.AuthMiddleware())
	{
		badges.GET("", authRepo.RetrieveBadges)
	}

	notifications := main.Group("/notifications", middleware.AuthMiddleware())
	{
		notifications.GET("", authRepo.RetrieveNotifications)
	}

	forgetPassword := r.Group("/forget-password")
	{
		forgetPassword.POST("", forgetPasswordRepo.Create)
		forgetPassword.POST("/:token", forgetPasswordRepo.Update)
	}

	return r
}
