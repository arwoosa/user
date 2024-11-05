package routes

import (
	"oosa/internal/middleware"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func OosaUserRoutes(r *gin.Engine) *gin.Engine {
	repoUser := repository.OosaUserRepository{}
	repoUserFriend := repository.UserFriendRepository{}
	repoUserBadges := repository.OosaUserBadgesRepository{}

	me := r.Group("/user/:id", middleware.AuthMiddleware())
	{
		me.GET("", repoUser.Read)
		me.GET("badges", repoUserBadges.Retrieve)
		me.GET("friends", repoUserFriend.RetrieveOther)
	}

	return r
}
