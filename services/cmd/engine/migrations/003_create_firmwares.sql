-- +goose Up
-- Create firmware table
CREATE TABLE IF NOT EXISTS firmwares (
    id SERIAL PRIMARY KEY,
    version VARCHAR(20) NOT NULL UNIQUE,
    blob_name VARCHAR(255) NOT NULL,
    blob_url TEXT NOT NULL,
    description TEXT,
    file_size BIGINT NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

-- Create index for active firmware queries
CREATE INDEX IF NOT EXISTS idx_firmwares_is_active ON firmwares(is_active, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS firmwares CASCADE;
