package routes

import (
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func ContactUsRoutes(r *gin.Engine) *gin.Engine {
	repository := repository.ContactUsRepository{}
	main := r.Group("/contact-us")
	{
		main.POST("", repository.Create)
	}
	return r
}
