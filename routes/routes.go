package routes

import (
	"net/http"
	_ "oosa/docs"
	"os"

	"database/sql"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

func RegisterRoutes() *gin.Engine {
	r := gin.Default()
	initSession(r)

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
	sessionStore := os.Getenv("SESSION_STORE_TYPE")
	if sessionStore == "redis" {
		initRedisSession(r)
	} else if sessionStore == "postgres" {
		initPostgresSession(r)
	}
}

func initRedisSession(r *gin.Engine) {
	redisHost := os.Getenv("SESSION_REDIS_HOST")
	if redisHost == "" {
		panic("SESSION_REDIS_HOST not set")
	}
	redisSessionDB := os.Getenv("SESSION_REDIS_DB")
	if redisSessionDB == "" {
		panic("SESSION_REDIS_DB not set")
	}
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		panic("SESSION_SECRET not set")
	}
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		panic("SESSION_KEY not set")
	}

	store, err := redis.NewStoreWithDB(10, "tcp", redisHost, "", redisSessionDB, []byte(sessionSecret))
	if err != nil {
		panic(err)
	}
	store.Options(sessions.Options{Secure: true, HttpOnly: true, MaxAge: 86400, SameSite: http.SameSiteLaxMode})
	r.Use(sessions.Sessions(sessionKey, store))
}

func initPostgresSession(r *gin.Engine) {
	connectionStr := os.Getenv("SESSION_PSQL_CONNECTION_STRING")
	if connectionStr == "" {
		panic("SESSION_PSQL_CONNECTION_STRING not set")
	}
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		panic("SESSION_SECRET not set")
	}
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		panic("SESSION_KEY not set")
	}

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	store, err := postgres.NewStore(db, []byte(sessionSecret))
	if err != nil {
		panic("failed new store: " + err.Error())
	}
	store.Options(sessions.Options{Secure: true, HttpOnly: true, MaxAge: 86400, SameSite: http.SameSiteLaxMode})
	r.Use(sessions.Sessions(sessionKey, store))
}
