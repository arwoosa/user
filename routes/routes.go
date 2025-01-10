package routes

import (
	"net/http"
	_ "oosa/docs"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

func RegisterRoutes() *gin.Engine {
	r := gin.Default()
	AuthRoutes(r)
	UserRoutes(r)
	OosaUserRoutes(r)
	PaymentRoutes(r)
	ContactUsRoutes(r)
	OosaDailyRoutes(r)
	WorldRoutes(r)
	SsoRoutes(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
}

func initSession(r *gin.Engine) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		panic("REDIS_HOST not set")
	}
	redisSecret := os.Getenv("REDIS_SECRET")
	if redisSecret == "" {
		panic("REDIS_SECRET not set")
	}
	redisSessionDB := os.Getenv("REDIS_SESSION_DB")
	if redisSessionDB == "" {
		panic("REDIS_SESSION_DB not set")
	}

	store, err := redis.NewStoreWithDB(10, "tcp", redisHost, "", redisSessionDB, []byte(redisSecret))
	if err != nil {
		panic(err)
	}
	store.Options(sessions.Options{Secure: true, HttpOnly: true, MaxAge: 86400, SameSite: http.SameSiteLaxMode})
	r.Use(sessions.Sessions("oosa_user_session", store))
}
