package repository

import (
	"context"
	"net/http"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OosaUserRepository struct{}

func (r OosaUserRepository) Read(c *gin.Context) {
	userIdVal := c.Param("id")
	userId, _ := primitive.ObjectIDFromHex(userIdVal)
	userDetail := helpers.GetAuthUser(c)

	var User models.Users
	err := r.ReadUserById(userId, &User)

	if err == mongo.ErrNoDocuments {
		helpers.ResponseNoData(c, "")
		return
	}

	var UserFollowing models.UserFollowings
	UserFollowingFilter := bson.D{
		{Key: "user_followings_user", Value: userDetail.UsersId},
		{Key: "user_followings_following", Value: User.UsersId},
	}
	UserFollowingErr := config.DB.Collection("UserFollowings").FindOne(context.TODO(), UserFollowingFilter).Decode(&UserFollowing)
	if UserFollowingErr == nil {
		User.UsersFollowings = &UserFollowing
	}

	c.JSON(http.StatusOK, User)
}

func (r OosaUserRepository) ReadUserById(userId primitive.ObjectID, User *models.Users) error {
	err := config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "_id", Value: userId}}).Decode(&User)
	return err
}
