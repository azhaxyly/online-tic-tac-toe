-- users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    nickname VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    total_games INT NOT NULL DEFAULT 0,
    wins INT NOT NULL DEFAULT 0,
    losses INT NOT NULL DEFAULT 0,
    draws INT NOT NULL DEFAULT 0
);

-- games
CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    player_x_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    player_o_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    started_at TIMESTAMP DEFAULT NOW(),
    ended_at TIMESTAMP,
    result VARCHAR(10) NOT NULL CHECK (result IN ('X', 'O', 'draw', 'forfeit')),
    moves JSONB
);
