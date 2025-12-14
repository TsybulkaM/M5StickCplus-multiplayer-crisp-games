package mqtt

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ScoreMessage struct {
	Game  string `json:"game"`
	Score int    `json:"score"`
}

func StartWorker(db *sql.DB, broker string) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("engine-worker")
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(false)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	log.Println("MQTT worker started")

	// Shared subscription for load balancing across multiple workers
	// Format: $share/group_name/topic
	sharedTopic := "$share/engine-workers/devices/+/score"
	if token := client.Subscribe(sharedTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		handleScoreMessage(db, msg)
	}); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to subscribe to %s: %v", sharedTopic, token.Error())
	}

	log.Printf("Subscribed to shared topic: %s", sharedTopic)

	select {} // Keep worker running
}

func handleScoreMessage(db *sql.DB, msg mqtt.Message) {
	parts := strings.Split(msg.Topic(), "/")
	if len(parts) != 3 {
		log.Printf("Invalid topic format: %s", msg.Topic())
		return
	}

	deviceID := parts[1]

	var scoreMsg ScoreMessage
	if err := json.Unmarshal(msg.Payload(), &scoreMsg); err != nil {
		log.Printf("Failed to parse message from %s: %v", deviceID, err)
		return
	}

	log.Printf("Device %s: game=%s, score=%d", deviceID, scoreMsg.Game, scoreMsg.Score)

	deviceQuery := `
		INSERT INTO devices (id, last_seen)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET last_seen = $2
	`
	_, err := db.Exec(deviceQuery, deviceID, time.Now())
	if err != nil {
		log.Printf("Failed to upsert device %s: %v", deviceID, err)
		return
	}

	scoreQuery := `
        INSERT INTO game_scores (device_id, game_code, score, created_at)
        VALUES ($1, $2, $3, $4)
    `

	_, err = db.Exec(scoreQuery, deviceID, scoreMsg.Game, scoreMsg.Score, time.Now())
	if err != nil {
		log.Printf("Failed to save score for device %s: %v", deviceID, err)
		return
	}

	log.Printf("Score saved successfully for device %s", deviceID)
}
