package routes

import (
	"oosa/pkg/repository"

	"github.com/gin-gonic/gin"
)

func WorldRoutes(r gin.IRouter) gin.IRouter {
	repository := repository.StatisticsRepository{}
	main := r.Group("/world")
	{
		main.GET("statistics", repository.Retrieve)
		main.GET("ranking-feelings", repository.RetrieveRankingFeelings)
		main.GET("ranking-rewilding", repository.RetrieveRankingRewilding)
	}

	return r
}
