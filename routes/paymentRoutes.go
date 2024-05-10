package routes

import (
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func PaymentRoutes(r *gin.Engine) *gin.Engine {
	repository := repository.PaymentRepository{}
	main := r.Group("/line/pay")
	{
		main.GET("", repository.MakeLinePayment)
		main.POST("/confirm", repository.ConfirmLinePayment)

		// main.POST("/", repository.UserFollowingCreate)

		// main.PUT("/:id", repository.UserFollowingUpdate)

		// main.DELETE("/:id", repository.UserFollowingDelete)
	}
	return r
}
