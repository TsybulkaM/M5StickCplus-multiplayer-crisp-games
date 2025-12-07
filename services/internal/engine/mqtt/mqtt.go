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

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	log.Println("MQTT worker started")
    
	client.Subscribe("devices/+/score", 0, func(client mqtt.Client, msg mqtt.Message) {
		handleScoreMessage(db, msg)
	})

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
