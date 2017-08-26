package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
)

const apiPath = "/api/v1"

func main() {
	configfile := pflag.StringP("config", "c", "/tmp/checkup.config", "Configuration file")
	httpService := pflag.String("http", ":8801", "HTTP service address")

	pflag.Parse()

	if err := loadCheckupConfiguration(*configfile); err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}

	initializeOutput()

	// Load data
	if err := loadInitialData(); err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}

	router := gin.Default()

	// Routes
	//router.POST(apiPath + "/:a/device", apiAB)

	router.GET(apiPath+"/check", apiCheck)
	router.GET(apiPath+"/checkup", apiCheckup)
	router.GET(apiPath+"/status", apiStatus)
	router.GET(apiPath+"/status/:site", apiStatusSite)
	router.GET(apiPath+"/timeline", apiTimeline)
	router.GET(apiPath+"/timeline/:site", apiTimeline)
	router.GET(apiPath+"/stats", apiStats)
	router.GET(apiPath+"/stats/:site", apiStats)

	// Periodic update
	startUpdateTimer()

	// Let's go!
	router.Run(*httpService)
}
