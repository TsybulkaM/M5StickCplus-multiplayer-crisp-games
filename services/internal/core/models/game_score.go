package models

import "time"

type GameScore struct {
	ID        int64     `db:"id"`
	DeviceID  string    `db:"device_id"`
	GameCode  string    `db:"game_code"`
	Score     int       `db:"score"`
	CreatedAt time.Time `db:"created_at"`
}
