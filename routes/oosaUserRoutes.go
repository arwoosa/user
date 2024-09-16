package routes

import (
	"oosa/internal/middleware"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func OosaUserRoutes(r *gin.Engine) *gin.Engine {
	repoUser := repository.OosaUserRepository{}
	repoUserBadges := repository.OosaUserBadgesRepository{}

	me := r.Group("/user/:id", middleware.AuthMiddleware())
	{
		me.GET("", repoUser.Read)
		me.GET("/badges", repoUserBadges.Retrieve)
	}

	return r
}
