package routes

import (
	"oosa/internal/middleware"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r gin.IRouter) gin.IRouter {
	userRepository := repository.UserRepository{}
	userFriendRepository := repository.UserFriendRepository{}
	userStatisticsRepository := repository.UserStatisticsRepository{}

	main := r.Group("/user-following", middleware.AuthMiddleware())
	{
		main.GET("", userRepository.UserFollowingRetrieve)
		main.POST("", userRepository.UserFollowingCreate)
	}

	detail := main.Group("/:id", middleware.AuthMiddleware())
	{
		detail.GET("", userRepository.UserFollowingRead)
		detail.PUT("", userRepository.UserFollowingUpdate)
		detail.DELETE("", userRepository.UserFollowingDelete)

	}

	users := r.Group("/users", middleware.AuthMiddleware())
	{
		users.GET("", userRepository.RetrieveUsers)
		users.GET("/statistics", userStatisticsRepository.Retrieve)
	}

	friends := r.Group("user/friends", middleware.AuthMiddleware())
	{
		friends.GET("", userFriendRepository.Retrieve)
		friends.POST("", userFriendRepository.Create)
		friends.PUT(":userFriendId", userFriendRepository.Update)
		friends.DELETE(":userFriendId", userFriendRepository.Delete)
		friends.GET("recommended", userFriendRepository.Recommended)
		//friends.GET("/friends", userRepository.RetrieveUserFriends)
		//friends.GET("/statistics", userStatisticsRepository.Retrieve)
	}

	return r
}
