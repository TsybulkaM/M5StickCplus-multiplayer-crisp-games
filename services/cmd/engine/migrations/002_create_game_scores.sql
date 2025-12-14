-- +goose Up
-- Create game_scores table
CREATE TABLE IF NOT EXISTS game_scores (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(50) NOT NULL,
    game_code VARCHAR(50) NOT NULL,
    score INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_device
        FOREIGN KEY(device_id) 
        REFERENCES devices(id)
        ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_game_scores_device_id ON game_scores(device_id);
CREATE INDEX IF NOT EXISTS idx_game_scores_game_code ON game_scores(game_code);
CREATE INDEX IF NOT EXISTS idx_game_scores_created_at ON game_scores(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_game_scores_score ON game_scores(score DESC);

-- Create composite index for leaderboard queries
CREATE INDEX IF NOT EXISTS idx_game_scores_game_score ON game_scores(game_code, score DESC, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS game_scores CASCADE;
