package main

import (
	"log"

	"github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/internal/core"
	"github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/internal/portal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	config := core.LoadConfig()
	db := core.ConnectDB()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.LoadHTMLGlob("internal/portal/templates/*.html")

	h := handlers.NewHandler(db)

	r.GET("/", h.ShowDevices)
	r.GET("/device/:id", h.ShowDeviceDetail)
	r.GET("/games", h.ShowGames)
	r.GET("/leaderboard/:game", h.ShowLeaderboard)

	log.Printf("Portal starting on port %s", config.Port)
	if err := r.Run(":" + config.Port); err != nil {
		log.Fatal(err)
	}
}
