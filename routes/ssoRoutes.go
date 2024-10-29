package routes

import (
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func SsoRoutes(r *gin.Engine) *gin.Engine {
	ssoRepo := repository.SsoRepository{}
	sso := r.Group("/sso")
	{
		sso.GET("/auth", ssoRepo.ResponseHeader)
		sso.GET("/no-auth", ssoRepo.ResponseHeader)
		sso.GET("/error", ssoRepo.Error)
	}

	register := sso.Group("/register")
	{
		register.GET("", ssoRepo.Register)
		register.POST("/finish", ssoRepo.CallbackAndSaveUser)

	}

	login := sso.Group("/login")
	{
		login.POST("/email", ssoRepo.LoginByEmail)
		login.GET("/:provider", ssoRepo.LoginByProvider)
		login.GET("/:provider/callback", ssoRepo.ProviderCallback)
		login.POST("/register", ssoRepo.OcidRegister)
	}

	recover := sso.Group("/recover")
	{
		recover.POST("", ssoRepo.Recover)
		recover.POST("code", ssoRepo.RecoverCode)
	}

	sso.PUT("/test", ssoRepo.PutTest)
	sso.POST("/test", ssoRepo.PutTest)
	return r
}
