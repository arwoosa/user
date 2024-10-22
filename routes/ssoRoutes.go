package routes

import (
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func SsoRoutes(r *gin.Engine) *gin.Engine {
	ssoRepo := repository.SsoRepository{}

	register := r.Group("/sso")
	{
		register.GET("/register", ssoRepo.Register)
		register.POST("/register/finish", ssoRepo.Register)
	}

	return r
}
