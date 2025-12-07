-- Create devices table
CREATE TABLE IF NOT EXISTS devices (
    id VARCHAR(50) PRIMARY KEY,
    user_id INTEGER,
    firmware_ver VARCHAR(20),
    last_seen TIMESTAMP,
    is_banned BOOLEAN DEFAULT FALSE
);

-- Create indexes
CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_last_seen ON devices(last_seen DESC);
