package main

import (
	"github.com/SkycoinPro/skywire-services-uptime/src/app"
	"github.com/SkycoinPro/skywire-services-uptime/src/config"
	"github.com/SkycoinPro/skywire-services-uptime/src/database/postgres"
	node_checker "github.com/SkycoinPro/skywire-services-uptime/src/node-checker"
	"github.com/spf13/viper"
)

// @title Skywire User System API
// @version 1.0
// @description This is a Skywire User System service.

// @host localhost:8080
// @BasePath /api/v1
func main() {
	config.Init("node-checker-config")

	tearDown := postgres.Init()
	defer tearDown()

	uc := node_checker.DefaultController()
	if viper.GetBool("config.make-uptimes-for-previous-months") {
		go uc.GetUptimesForPreviousMonths()
	}
	go uc.RunningRoutine()
	// register all of the controllers here
	app.NewServer(
		uc,
	).Run()
}
