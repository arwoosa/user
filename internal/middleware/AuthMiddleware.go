package middleware

import (
	"context"
	"fmt"
	"net/http"
	"oosa/internal/auth"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"oosa/internal/structs"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ssoAuth(c) {
			return
		}
		reqToken := c.Request.Header.Get("Authorization")
		if reqToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "AUTH01-USER: You are not authorized to access this resource"})
			c.Abort()
			return
		}

		splitToken := strings.Split(reqToken, "Bearer ")
		reqToken = splitToken[1]

		user, err := auth.VerifyToken(reqToken)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "AUTH02-USER: You are not authorized to access this resource"})
			c.Abort()
			return
		}

		if helpers.MongoZeroID(user.UsersId) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "AUTH03-USER: You are not authorized to access this resource"})
			c.Abort()
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}

func ssoAuth(c *gin.Context) bool {
	headerUserId := c.GetHeader("X-User-Id")
	if headerUserId == "" {
		return false
	}
	if headerUserId == "guest" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "AUTH01-USER: You are not authorized to access this resource"})
		c.Abort()
		return true
	}

	var headerUser structs.UserBindByHeader
	err := c.BindHeader(&headerUser)
	if err != nil || headerUser.Id == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "AUTH01-USER: You are not authorized to access this resource"})
		c.Abort()
		return true
	}

	var user models.Users
	err = config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "users_source_id", Value: headerUser.Id}}).Decode(&user)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "AUTH01-USER: You are not authorized to access this resource"})
		c.Abort()
		return true
	}
	for k, h := range c.Request.Header {
		fmt.Println(k, h)
	}
	needUpdate := false
	if user.UsersAvatar == "" {
		user.UsersAvatar = headerUser.Avatar
		needUpdate = true
	}
	if user.UsersUsername == "" {
		user.UsersUsername = headerUser.User
		needUpdate = true
	}
	if needUpdate {
		config.DB.Collection("Users").UpdateByID(c, user.UsersId, bson.D{{Key: "$set", Value: bson.D{{Key: "users_avatar", Value: user.UsersAvatar}, {Key: "users_username", Value: user.UsersUsername}}}})
	}
	c.Set("user", &user)
	c.Next()
	return true
}
