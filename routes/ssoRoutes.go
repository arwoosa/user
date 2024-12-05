package routes

import (
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func SsoRoutes(r *gin.Engine) *gin.Engine {
	ssoRepo := repository.SsoRepository{}
	sso := r.Group("/sso")
	register := sso.Group("/register")
	{
		register.GET("", ssoRepo.Register)
		register.GET("/finish",
			ssoRepo.CallbackAndSaveUser)

	}
	return r
}
