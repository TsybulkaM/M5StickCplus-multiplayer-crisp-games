package fota

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

type Handler struct {
	db            *sql.DB
	blobClient    *azblob.Client
	containerName string
}

type CheckUpdateResponse struct {
	Status      string `json:"status"`
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
	FileSize    int64  `json:"file_size"`
	Checksum    string `json:"checksum"`
	Description string `json:"description"`
}

func NewHandler(db *sql.DB, storageAccount, storageKey, containerName string) (*Handler, error) {
	// Создаем клиент Azure Blob Storage
	var blobClient *azblob.Client

	if storageAccount != "" && storageKey != "" {
		credential, err := azblob.NewSharedKeyCredential(storageAccount, storageKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create credential: %w", err)
		}

		serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccount)
		blobClient, err = azblob.NewClientWithSharedKeyCredential(serviceURL, credential, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create blob client: %w", err)
		}

		log.Printf("Azure Blob Storage client initialized: account=%s, container=%s", storageAccount, containerName)
	} else {
		log.Println("Azure Storage credentials not provided, using local storage fallback")
	}

	return &Handler{
		db:            db,
		blobClient:    blobClient,
		containerName: containerName,
	}, nil
}

func (h *Handler) CheckUpdate(w http.ResponseWriter, r *http.Request) {
	currentVersion := r.URL.Query().Get("current_version")
	deviceID := r.URL.Query().Get("device_id")

	log.Printf("FOTA check: device=%s, current_version=%s", deviceID, currentVersion)

	query := `
		SELECT version, blob_url, description, file_size, checksum
		FROM firmwares
		WHERE is_active = TRUE
		ORDER BY created_at DESC
		LIMIT 1
	`

	var firmware struct {
		Version     string
		BlobURL     string
		Description string
		FileSize    int64
		Checksum    string
	}

	err := h.db.QueryRow(query).Scan(
		&firmware.Version,
		&firmware.BlobURL,
		&firmware.Description,
		&firmware.FileSize,
		&firmware.Checksum,
	)

	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CheckUpdateResponse{
			Status: "no_update",
		})
		return
	}

	if err != nil {
		log.Printf("Failed to query firmware: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if currentVersion == firmware.Version {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CheckUpdateResponse{
			Status: "no_update",
		})
		return
	}

	response := CheckUpdateResponse{
		Status:      "update_available",
		Version:     firmware.Version,
		DownloadURL: "/api/fota/download?version=" + firmware.Version,
		FileSize:    firmware.FileSize,
		Checksum:    firmware.Checksum,
		Description: firmware.Description,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Update available for device %s: %s -> %s", deviceID, currentVersion, firmware.Version)
}

func (h *Handler) DownloadBin(w http.ResponseWriter, r *http.Request) {
	version := r.URL.Query().Get("version")
	deviceID := r.URL.Query().Get("device_id")

	log.Printf("FOTA download: device=%s, version=%s", deviceID, version)

	// Получаем информацию о прошивке из БД
	query := `SELECT blob_name, blob_url, file_size, checksum FROM firmwares WHERE version = $1 AND is_active = TRUE`

	var blobName, blobURL string
	var fileSize int64
	var checksum string
	err := h.db.QueryRow(query, version).Scan(&blobName, &blobURL, &fileSize, &checksum)

	if err == sql.ErrNoRows {
		log.Printf("Firmware version %s not found in database", version)
		http.Error(w, "Firmware not found", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Printf("Failed to query firmware: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Скачиваем из Azure Blob Storage
	ctx := context.Background()
	blobClient := h.blobClient.ServiceClient().NewContainerClient(h.containerName).NewBlobClient(blobName)

	// Скачиваем blob
	downloadResp, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		log.Printf("Failed to download blob: %v", err)
		http.Error(w, "Failed to download firmware", http.StatusInternalServerError)
		return
	}
	defer downloadResp.Body.Close()

	// Устанавливаем заголовки
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=firmware_"+version+".bin")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
	w.Header().Set("X-Firmware-Version", version)
	w.Header().Set("X-Firmware-Checksum", checksum)

	log.Printf("Serving firmware %s (%d bytes) to device %s", version, fileSize, deviceID)

	// Отдаём файл
	_, err = io.Copy(w, downloadResp.Body)
	if err != nil {
		log.Printf("Failed to send firmware: %v", err)
		return
	}

	log.Printf("Firmware %s successfully downloaded by device %s", version, deviceID)
}

func (h *Handler) UploadBin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart form (максимум 100MB)
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Получаем параметры
	version := r.FormValue("version")
	description := r.FormValue("description")

	if version == "" {
		http.Error(w, "Version is required", http.StatusBadRequest)
		return
	}

	// Получаем файл
	file, header, err := r.FormFile("firmware")
	if err != nil {
		log.Printf("Failed to get firmware file: %v", err)
		http.Error(w, "Firmware file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Uploading firmware: version=%s, filename=%s, size=%d", version, header.Filename, header.Size)

	// Читаем файл в память и вычисляем MD5
	hash := md5.New()
	fileContent, err := io.ReadAll(io.TeeReader(file, hash))
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fileSize := int64(len(fileContent))
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	log.Printf("File read: size=%d, checksum=%s", fileSize, checksum)

	// Загружаем в Azure Blob Storage
	ctx := context.Background()
	blobName := fmt.Sprintf("firmware_v%s.bin", version)
	blobClient := h.blobClient.ServiceClient().NewContainerClient(h.containerName).NewBlockBlobClient(blobName)

	contentType := "application/octet-stream"
	_, err = blobClient.UploadBuffer(ctx, fileContent, &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
	})
	if err != nil {
		log.Printf("Failed to upload to Azure Blob Storage: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Генерируем SAS URL для скачивания (действителен 10 лет)
	sasURL, err := h.generateSASURL(blobName, 10*365*24*time.Hour)
	if err != nil {
		log.Printf("Failed to generate SAS URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Blob uploaded successfully: %s", blobName)

	// Сохраняем в БД
	query := `
		INSERT INTO firmwares (version, blob_name, blob_url, description, file_size, checksum, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE)
		ON CONFLICT (version) DO UPDATE 
		SET blob_name = EXCLUDED.blob_name,
			blob_url = EXCLUDED.blob_url,
			description = EXCLUDED.description,
			file_size = EXCLUDED.file_size,
			checksum = EXCLUDED.checksum,
			is_active = EXCLUDED.is_active,
			created_at = NOW()
	`

	_, err = h.db.Exec(query, version, blobName, sasURL, description, fileSize, checksum)
	if err != nil {
		log.Printf("Failed to save firmware info to database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Firmware %s uploaded successfully to Azure Blob Storage", version)

	// Возвращаем информацию о загруженной прошивке
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"version":     version,
		"file_size":   fileSize,
		"checksum":    checksum,
		"description": description,
		"blob_name":   blobName,
		"blob_url":    sasURL,
	})
}

func (h *Handler) generateSASURL(blobName string, duration time.Duration) (string, error) {
	// Получаем blob client
	containerClient := h.blobClient.ServiceClient().NewContainerClient(h.containerName)
	blobClient := containerClient.NewBlobClient(blobName)

	// Создаем SAS токен
	startTime := time.Now().Add(-15 * time.Minute)
	expiryTime := time.Now().Add(duration)

	sasQueryParams, err := blobClient.GetSASURL(
		sas.BlobPermissions{Read: true},
		expiryTime,
		&blob.GetSASURLOptions{
			StartTime: &startTime,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate SAS URL: %w", err)
	}

	return sasQueryParams, nil
}
