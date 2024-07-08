package repository

import (
	"context"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContactUsRepository struct{}
type ContactUsRequest struct {
	Email   string `json:"email" binding:"required"`
	Message string `json:"message" binding:"required"`
}

func (t ContactUsRepository) Create(c *gin.Context) {
	var payload ContactUsRequest
	validateError := helpers.Validate(c, &payload)
	if validateError != nil {
		return
	}

	insert := models.ContactUs{
		ContactUsEmail:     payload.Email,
		ContactUsMessage:   payload.Message,
		ContactUsCreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	_, err := config.DB.Collection("ContactUs").InsertOne(context.TODO(), insert)

	if err != nil {
		helpers.ResponseBadRequestError(c, err.Error())
	}

	helpers.ResponseSuccessMessage(c, "Thank you for contacting us")
}
