package repository

import (
	"context"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ForgetPasswordRepository struct{}
type ForgetPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}
type ForgetPasswordUpdateRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (t ForgetPasswordRepository) Create(c *gin.Context) {
	var payload ForgetPasswordRequest
	validateError := helpers.Validate(c, &payload)
	if validateError != nil {
		return
	}

	var User models.Users
	config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "users_email", Value: payload.Email}}).Decode(&User)

	if !helpers.MongoZeroID(User.UsersId) {
		token := helpers.RandomString(32)
		insert := models.PasswordResets{
			PasswordResetsUserId:    User.UsersId,
			PasswordResetsEmail:     User.UsersEmail,
			PasswordResetsToken:     token,
			PasswordResetsIsActive:  1,
			PasswordResetsCreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		}

		_, err := config.DB.Collection("PasswordResets").InsertOne(context.TODO(), insert)
		if err != nil {
			helpers.ResponseBadRequestError(c, err.Error())
			return
		}

		c.JSON(200, gin.H{"token": token})
		return
	}
	helpers.ResponseNoData(c, "")
}

func (t ForgetPasswordRepository) Update(c *gin.Context) {
	var payload ForgetPasswordUpdateRequest
	token := c.Param("token")
	validateError := helpers.Validate(c, &payload)
	if validateError != nil {
		return
	}

	var PasswordResets models.PasswordResets
	config.DB.Collection("PasswordResets").FindOne(context.TODO(), bson.D{
		{Key: "password_resets_token", Value: token},
		{Key: "password_resets_email", Value: payload.Email},
	}).Decode(&PasswordResets)

	if helpers.MongoZeroID(PasswordResets.PasswordResetsId) {
		helpers.ResponseBadRequestError(c, "Reset token not found or incorrect email")
		return
	}

	hashedPassword, err := helpers.HashPassword(payload.Password)
	if err != nil {
		helpers.ResponseError(c, "Failed to hash password")
		return
	}

	if PasswordResets.PasswordResetsIsActive == 0 {
		helpers.ResponseError(c, "Reset token used")
		return
	}

	filters := bson.D{{Key: "_id", Value: PasswordResets.PasswordResetsUserId}}
	upd := bson.D{{Key: "$set", Value: models.Users{
		UsersPassword: hashedPassword,
	}}}
	config.DB.Collection("Users").UpdateOne(context.TODO(), filters, upd)

	updPasswordReset := bson.D{
		{Key: "$set", Value: bson.D{{Key: "password_resets_is_active", Value: 0}}},
	}
	config.DB.Collection("PasswordResets").UpdateOne(context.TODO(), bson.D{{Key: "_id", Value: PasswordResets.PasswordResetsId}}, updPasswordReset)

	c.JSON(200, gin.H{"message": "Password updated successfully"})

}
