package routes

import (
	"oosa/internal/middleware"
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
			middleware.AuthMiddleware(),
			ssoRepo.CallbackAndSaveUser)

	}
	return r
}
