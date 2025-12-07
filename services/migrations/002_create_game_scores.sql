-- Active: 1765105621778@@127.0.0.1@5432@crisp_db
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
CREATE INDEX idx_game_scores_device_id ON game_scores(device_id);
CREATE INDEX idx_game_scores_game_code ON game_scores(game_code);
CREATE INDEX idx_game_scores_created_at ON game_scores(created_at DESC);
CREATE INDEX idx_game_scores_score ON game_scores(score DESC);

-- Create composite index for leaderboard queries
CREATE INDEX idx_game_scores_game_score ON game_scores(game_code, score DESC, created_at DESC);
