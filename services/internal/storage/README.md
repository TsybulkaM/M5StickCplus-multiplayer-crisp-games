# Storage Implementation

## Overview

The FOTA (Firmware Over-The-Air) system uses a unified storage interface powered by [Go Cloud Development Kit (gocloud.dev)](https://gocloud.dev/howto/blob/), which provides seamless switching between Azure Blob Storage (production) and local filesystem (development).

## Architecture

```
┌─────────────────┐
│  FOTA Handler   │
└────────┬────────┘
         │
         v
┌─────────────────┐
│ Storage Interface│
└────────┬────────┘
         │
    ┌────┴─────┐
    v          v
┌────────┐  ┌─────────┐
│ Azure  │  │  Local  │
│ Blob   │  │  File   │
└────────┘  └─────────┘
```

## Storage Interface

Located in `internal/storage/storage.go`:

```go
type Storage interface {
    Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error)
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    GetURL(ctx context.Context, key string) (string, error)
    List(ctx context.Context, prefix string) ([]string, error)
    Close() error
}
```

## Implementations

### 1. Azure Blob Storage (Production)

**When**: Automatically used when `AZURE_STORAGE_ACCOUNT` and `AZURE_STORAGE_KEY` are set.

**Environment Variables**:
```bash
STORAGE_TYPE=azure                                    # Optional, auto-detected
AZURE_STORAGE_ACCOUNT=your-storage-account           # Required
AZURE_STORAGE_KEY=your-storage-key                   # Required
AZURE_BLOB_CONTAINER=firmware                        # Optional, defaults to "firmware"
```

**Terraform Configuration**:
- Storage account: `terraform/storage.tf`
- Container: `firmware` (auto-created)
- Outputs: `storage_account_name`, `storage_account_key`, `firmware_container_name`

### 2. Local Filesystem (Development)

**When**: Used when Azure credentials are not present, or explicitly set with `STORAGE_TYPE=local`.

**Environment Variables**:
```bash
STORAGE_TYPE=local                                    # Optional
LOCAL_STORAGE_PATH=./firmware_storage                # Optional, default
BASE_URL=http://localhost:8081/api/fota/download     # Optional, for URL generation
```

**Docker Compose**: Volume mounted at `./services/firmware_storage:/root/firmware_storage`

## Configuration

### Docker Compose (Local Development)

```yaml
engine:
  environment:
    - STORAGE_TYPE=local
    - LOCAL_STORAGE_PATH=./firmware_storage
    - BASE_URL=http://localhost:8081/api/fota/download
  volumes:
    - ./services/firmware_storage:/root/firmware_storage
```

### Kubernetes (Production)

```yaml
env:
  - name: AZURE_STORAGE_ACCOUNT
    valueFrom:
      secretKeyRef:
        name: app-secrets
        key: AZURE_STORAGE_ACCOUNT
  - name: AZURE_STORAGE_KEY
    valueFrom:
      secretKeyRef:
        name: app-secrets
        key: AZURE_STORAGE_KEY
  - name: AZURE_BLOB_CONTAINER
    value: "firmware"
  - name: STORAGE_TYPE
    value: "azure"
```

## Usage Example

### Upload Firmware

```bash
curl -X POST http://localhost:8081/api/fota/upload \
  -F "version=1.0.0" \
  -F "description=Initial release" \
  -F "firmware=@/home/mikita/Programming/M5StickCplus-multiplayer-crisp-games/esp-crisp/.pio/build/m5stick-c-plus/firmware.bin"
```

Response:
```json
{
  "status": "success",
  "version": "1.0.0",
  "file_size": 1048576,
  "checksum": "abc123...",
  "description": "Initial release",
  "blob_name": "firmware_v1.0.0.bin",
  "blob_url": "https://storage.blob.core.windows.net/firmware/firmware_v1.0.0.bin"
}
```

### Check for Updates

```bash
curl "http://localhost:8081/api/fota/check?device_id=ESP32-001&current_version=0.9.0"
```

Response:
```json
{
  "status": "update_available",
  "version": "1.0.0",
  "download_url": "/api/fota/download?version=1.0.0",
  "file_size": 1048576,
  "checksum": "abc123...",
  "description": "Initial release"
}
```

### Download Firmware

```bash
curl "http://localhost:8081/api/fota/download?version=1.0.0&device_id=ESP32-001" -o firmware.bin
```

## Migration Guide

### From Old Implementation

**Before**:
```go
fotaHandler, err := fota.NewHandler(db, storageAccount, storageKey, containerName)
```

**After**:
```go
stor, err := storage.NewStorage(context.Background())
if err != nil {
    log.Fatal(err)
}
defer stor.Close()

fotaHandler, err := fota.NewHandler(db, stor)
```

### Benefits

1. **Environment-aware**: Automatically selects the right storage based on configuration
2. **Portable**: Same code works locally and in production
3. **Testable**: Easy to mock for unit tests
4. **Cloud-native**: Uses Go Cloud CDK, can add GCS/S3 support easily
5. **Clean**: Separates storage logic from business logic

## Testing

### Local Testing
```bash
cd services
export STORAGE_TYPE=local
export LOCAL_STORAGE_PATH=./firmware_storage
go run ./cmd/engine
```

### Azure Testing
```bash
cd services
export STORAGE_TYPE=azure
export AZURE_STORAGE_ACCOUNT=your-account
export AZURE_STORAGE_KEY=your-key
export AZURE_BLOB_CONTAINER=firmware
go run ./cmd/engine
```

## Troubleshooting

### Issue: "Failed to open bucket"
- **Local**: Check `LOCAL_STORAGE_PATH` exists and is writable
- **Azure**: Verify `AZURE_STORAGE_ACCOUNT` and `AZURE_STORAGE_KEY` are correct

### Issue: "Failed to download firmware"
- Check blob name exists in database
- For local: Verify file exists in `LOCAL_STORAGE_PATH`
- For Azure: Check container permissions and blob exists

### Issue: "Unknown storage type"
- Set `STORAGE_TYPE` to either `azure` or `local`
- Or provide appropriate credentials for auto-detection

## Future Enhancements

- [ ] Add Google Cloud Storage support
- [ ] Add AWS S3 support
- [ ] Implement caching layer
- [ ] Add multipart upload for large files
- [ ] Generate signed URLs for direct client downloads
