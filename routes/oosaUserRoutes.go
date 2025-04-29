package routes

import (
	"oosa/internal/middleware"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func OosaUserRoutes(r gin.IRouter) gin.IRouter {
	repoUser := repository.OosaUserRepository{}
	repoUserFriend := repository.UserFriendRepository{}
	repoUserBadges := repository.OosaUserBadgesRepository{}

	r.GET("/user/:id", repoUser.Read)

	me := r.Group("/user/:id", middleware.AuthMiddleware())
	{
		me.GET("badges", repoUserBadges.Retrieve)
		me.GET("friends", repoUserFriend.RetrieveOther)
	}

	return r
}
