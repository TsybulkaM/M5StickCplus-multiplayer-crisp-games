package core

import (
	"os"
)

type Config struct {
    DatabaseURL         string
    MQTTBroker          string
    FilesDir            string
    Port                string
    AzureStorageAccount string
    AzureStorageKey     string
    AzureBlobContainer  string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/crisp_db?sslmode=disable"),
		MQTTBroker:          getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		FilesDir:            getEnv("FILES_DIR", "/mnt/azure_files"),
		Port:                getEnv("PORT", "8080"),
		AzureStorageAccount: getEnv("AZURE_STORAGE_ACCOUNT", ""),
		AzureStorageKey:     getEnv("AZURE_STORAGE_KEY", ""),
		AzureBlobContainer:  getEnv("AZURE_BLOB_CONTAINER", "firmware"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
