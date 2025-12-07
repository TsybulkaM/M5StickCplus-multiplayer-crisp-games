package core

import (
    "database/sql"
    "log"
    
    _ "github.com/lib/pq"
)

func ConnectDB() *sql.DB {
	config := LoadConfig()
    connStr := config.DatabaseURL
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    
    if err = db.Ping(); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Database connected")
    return db
}