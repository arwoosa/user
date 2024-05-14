package routes

import (
	"oosa/internal/middleware"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) *gin.Engine {
	repository := repository.UserRepository{}
	main := r.Group("/user-following", middleware.AuthMiddleware())
	{
		main.GET("/", repository.UserFollowingRetrieve)
		main.POST("/", repository.UserFollowingCreate)
	}

	detail := main.Group("/:id", middleware.AuthMiddleware())
	{
		detail.GET("", repository.UserFollowingRead)
		detail.PUT("", repository.UserFollowingUpdate)
		detail.DELETE("", repository.UserFollowingDelete)
	}

	me := r.Group("/user", middleware.AuthMiddleware())
	{
		me.GET("/friends", repository.RetrieveUserFriends)
	}

	usersetting := me.Group("/setting", middleware.AuthMiddleware())
	{
		usersetting.GET("/:id", repository.RetrieveUserSettings)
		usersetting.PUT("/:id", repository.UpdateUserSettings)
	}
	return r
}
