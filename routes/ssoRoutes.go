package routes

import (
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func SsoRoutes(r *gin.Engine) *gin.Engine {
	ssoRepo := repository.SsoRepository{}

	register := r.Group("/sso/register")
	{
		register.GET("", ssoRepo.Register)
		register.POST("/finish", ssoRepo.CallbackAndSaveUser)

	}

	return r
}
