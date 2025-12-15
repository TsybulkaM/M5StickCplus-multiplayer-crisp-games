package main

import (
	"context"
	"log"
	"net/http"

	"embed"
	"os"

	"github.com/pressly/goose/v3"

	"github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/internal/core"
	"github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/internal/engine/fota"
	"github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/internal/engine/mqtt"
	"github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/internal/storage"
)

//go:embed migrations
var embedMigrations embed.FS

func main() {
	config := core.LoadConfig()
	db := core.ConnectDB()

	// Handle database migrations
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		goose.SetBaseFS(embedMigrations)
		if err := goose.Up(db, "migrations"); err != nil {
			panic(err)
		}
		return
	}

	go mqtt.StartWorker(db, config.MQTTBroker)

	// Initialize storage
	stor, err := storage.NewStorage(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer stor.Close()

	fotaHandler, err := fota.NewHandler(db, stor)
	if err != nil {
		log.Fatalf("Failed to create FOTA handler: %v", err)
	}

	http.HandleFunc("/api/fota/check", fotaHandler.CheckUpdate)
	http.HandleFunc("/api/fota/download", fotaHandler.DownloadBin)
	http.HandleFunc("/api/fota/upload", fotaHandler.UploadBin)

	port := config.Port
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	log.Printf("Engine starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
