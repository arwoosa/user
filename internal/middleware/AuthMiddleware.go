package middleware

import (
	"net/http"
	"oosa/internal/auth"
	"oosa/internal/helpers"
	"strings"

	"github.com/gin-gonic/gin"
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
	if _, ok := c.Get("user"); ok {
		c.Next()
		return true
	}
	return false
}
