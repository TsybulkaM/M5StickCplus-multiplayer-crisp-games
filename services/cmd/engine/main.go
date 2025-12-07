package main

import (
    "log"
    "net/http"

    "github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/internal/core"
    "github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/internal/engine/mqtt"
    "github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/internal/engine/fota"
)

func main() {
    config := core.LoadConfig()
    db := core.ConnectDB()

    go mqtt.StartWorker(db, config.MQTTBroker)

    fotaHandler := fota.NewHandler(db, config.FilesDir)
    
    http.HandleFunc("/api/fota/check", fotaHandler.CheckUpdate)
    http.HandleFunc("/api/fota/download", fotaHandler.DownloadBin)

    log.Printf("Starting server on port %s", config.Port)
    if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
        log.Fatal(err)
    }
}