package models

import "time"

type Firmware struct {
	ID          int64     `db:"id"`
	Version     string    `db:"version"`
	FilePath    string    `db:"file_path"`
	Description string    `db:"description"`
	FileSize    int64     `db:"file_size"`
	Checksum    string    `db:"checksum"`
	CreatedAt   time.Time `db:"created_at"`
	IsActive    bool      `db:"is_active"`
}
