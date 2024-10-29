package main

import (
	"flag"
	"fmt"
	"oosa/internal/config"
	"oosa/routes"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	dev = flag.Bool("dev", false, "isDev")
)

var db *mongo.Database

// @title		OOSA API
// @version 	1.0
// @description This is OOSA's API. Generate the swagger documentation by running `swag init --parseDependency`
// @securityDefinitions.apikey 		ApiKeyAuth
// @in								header
// @name 							Authorization
func main() {
	flag.Parse()

	config.InitialiseConfig()
	db = config.ConnectDatabase()

	appPort := config.APP.AppPort
	if *dev {
		fmt.Println("Starting dev mode app on port: ", appPort)
	} else {
		fmt.Println("Starting app on port: ", appPort)
	}

	engine := routes.RegisterRoutes(*dev)
	engine.Run(":" + appPort)
}
