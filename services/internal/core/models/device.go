package models

import "time"

type Device struct {
    ID          string    `gorm:"primaryKey"`
    UserID      uint
    FirmwareVer string
    LastSeen    time.Time
    IsBanned    bool
}