package repository

import (
	"net/http"
	"oosa/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OosaUserBadgesRepository struct{}

func (r OosaUserBadgesRepository) Retrieve(c *gin.Context) {
	userIdVal := c.Param("id")
	userId, _ := primitive.ObjectIDFromHex(userIdVal)
	var Badges []models.Badges
	AuthRepository{}.RetrieveBadgesByUserId(c, userId, &Badges)
	c.JSON(http.StatusOK, Badges)
}
