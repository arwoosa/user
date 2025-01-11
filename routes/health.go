package routes

import (
	"oosa/internal/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func healthRoutes(r *gin.Engine) *gin.Engine {
	main := r.Group("/health")
	{
		main.GET("alive", aliveHandler)
		main.GET("ready", readyHandler)
	}
	return r
}

func aliveHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "I am alive",
	})
}

func readyHandler(c *gin.Context) {
	client, err := mongo.Connect(c, options.Client().ApplyURI(config.APP.DbConnection))
	if err != nil {
		c.JSON(500, gin.H{"message": "Database is not ready, connect error: " + err.Error()})
		return
	}
	err = client.Ping(c, nil)
	if err != nil {
		c.JSON(500, gin.H{"message": "Database is not ready, ping error: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"message": "service app is ready",
	})
}
