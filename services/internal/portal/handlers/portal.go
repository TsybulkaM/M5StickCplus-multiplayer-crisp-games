package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *sql.DB
}

type Device struct {
	ID          string
	LastSeen    *time.Time
	FirmwareVer *string
	TotalScores int
	GameCount   int
}

type GameScore struct {
	ID        int64
	DeviceID  string
	GameCode  string
	Score     int
	CreatedAt time.Time
}

type DeviceDetail struct {
	Device
	RecentScores []GameScore
	BestScores   map[string]int // game_code -> best score
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// Главная страница со списком всех устройств
func (h *Handler) ShowDevices(c *gin.Context) {
	search := c.Query("search")

	query := `
		SELECT 
			d.id,
			d.last_seen,
			d.firmware_ver,
			COALESCE(SUM(gs.score), 0) as total_scores,
			COALESCE(COUNT(gs.id), 0) as game_count
		FROM devices d
		LEFT JOIN game_scores gs ON d.id = gs.device_id
		WHERE ($1 = '' OR d.id ILIKE '%' || $1 || '%')
		GROUP BY d.id, d.last_seen, d.firmware_ver
		ORDER BY d.last_seen DESC NULLS LAST
	`

	rows, err := h.db.Query(query, search)
	if err != nil {
		log.Printf("Failed to query devices: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		err := rows.Scan(&d.ID, &d.LastSeen, &d.FirmwareVer, &d.TotalScores, &d.GameCount)
		if err != nil {
			log.Printf("Failed to scan device: %v", err)
			continue
		}
		devices = append(devices, d)
	}

	c.HTML(http.StatusOK, "devices.html", gin.H{
		"devices": devices,
		"search":  search,
	})
}

func (h *Handler) ShowDeviceDetail(c *gin.Context) {
	deviceID := c.Param("id")

	var device DeviceDetail
	err := h.db.QueryRow(`
		SELECT 
			d.id,
			d.last_seen,
			d.firmware_ver,
			COALESCE(SUM(gs.score), 0) as total_scores,
			COALESCE(COUNT(gs.id), 0) as game_count
		FROM devices d
		LEFT JOIN game_scores gs ON d.id = gs.device_id
		WHERE d.id = $1
		GROUP BY d.id, d.last_seen, d.firmware_ver
	`, deviceID).Scan(&device.ID, &device.LastSeen, &device.FirmwareVer, &device.TotalScores, &device.GameCount)

	if err == sql.ErrNoRows {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Device not found"})
		return
	}
	if err != nil {
		log.Printf("Failed to query device: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Получаем последние 50 результатов
	rows, err := h.db.Query(`
		SELECT id, device_id, game_code, score, created_at
		FROM game_scores
		WHERE device_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, deviceID)
	if err != nil {
		log.Printf("Failed to query scores: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var score GameScore
			err := rows.Scan(&score.ID, &score.DeviceID, &score.GameCode, &score.Score, &score.CreatedAt)
			if err != nil {
				log.Printf("Failed to scan score: %v", err)
				continue
			}
			device.RecentScores = append(device.RecentScores, score)
		}
	}

	// Получаем лучшие результаты по каждой игре
	device.BestScores = make(map[string]int)
	bestRows, err := h.db.Query(`
		SELECT game_code, MAX(score) as best_score
		FROM game_scores
		WHERE device_id = $1
		GROUP BY game_code
	`, deviceID)
	if err != nil {
		log.Printf("Failed to query best scores: %v", err)
	} else {
		defer bestRows.Close()
		for bestRows.Next() {
			var gameCode string
			var bestScore int
			err := bestRows.Scan(&gameCode, &bestScore)
			if err != nil {
				log.Printf("Failed to scan best score: %v", err)
				continue
			}
			device.BestScores[gameCode] = bestScore
		}
	}

	c.HTML(http.StatusOK, "device_detail.html", gin.H{
		"device": device,
	})
}

// Лидерборд по игре
func (h *Handler) ShowLeaderboard(c *gin.Context) {
	gameCode := c.Param("game")

	type LeaderboardEntry struct {
		Rank      int
		DeviceID  string
		Score     int
		CreatedAt time.Time
	}

	rows, err := h.db.Query(`
		SELECT device_id, score, created_at
		FROM game_scores
		WHERE game_code = $1
		ORDER BY score DESC, created_at ASC
		LIMIT 100
	`, gameCode)
	if err != nil {
		log.Printf("Failed to query leaderboard: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	rank := 1
	for rows.Next() {
		var entry LeaderboardEntry
		err := rows.Scan(&entry.DeviceID, &entry.Score, &entry.CreatedAt)
		if err != nil {
			log.Printf("Failed to scan leaderboard entry: %v", err)
			continue
		}
		entry.Rank = rank
		entries = append(entries, entry)
		rank++
	}

	c.HTML(http.StatusOK, "leaderboard.html", gin.H{
		"game":    gameCode,
		"entries": entries,
	})
}

// Список всех игр
func (h *Handler) ShowGames(c *gin.Context) {
	type GameInfo struct {
		GameCode   string
		PlayCount  int
		TopScore   int
		LastPlayed *time.Time
	}

	rows, err := h.db.Query(`
		SELECT 
			game_code,
			COUNT(*) as play_count,
			MAX(score) as top_score,
			MAX(created_at) as last_played
		FROM game_scores
		GROUP BY game_code
		ORDER BY play_count DESC
	`)
	if err != nil {
		log.Printf("Failed to query games: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var games []GameInfo
	for rows.Next() {
		var game GameInfo
		err := rows.Scan(&game.GameCode, &game.PlayCount, &game.TopScore, &game.LastPlayed)
		if err != nil {
			log.Printf("Failed to scan game: %v", err)
			continue
		}
		games = append(games, game)
	}

	c.HTML(http.StatusOK, "games.html", gin.H{
		"games": games,
	})
}
