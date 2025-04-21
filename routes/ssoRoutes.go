package routes

import (
	"net/url"
	"oosa/internal/config"
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func SsoRoutes(r gin.IRouter) gin.IRouter {
	var err error
	registerUrl := config.APP.SSORegisterUrl
	if registerUrl == "" {
		panic("SSO_REGISTER_URL not set")
	}
	ssoRegisterUrl, err := url.Parse(registerUrl)
	if err != nil {
		panic("failed to parse SSO_REGISTER_URL: " + err.Error())
	}
	ssoRepo := repository.NewSSoRepository(ssoRegisterUrl)
	register := r.Group("/sso/register")
	{
		initSession(register)
		register.GET("", ssoRepo.Register)
		register.GET("/finish",
			ssoRepo.CallbackAndSaveUser)

	}
	return r
}
