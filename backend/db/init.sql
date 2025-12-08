-- db/init.sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    nickname VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    total_games INT NOT NULL DEFAULT 0,
    wins INT NOT NULL DEFAULT 0,
    losses INT NOT NULL DEFAULT 0,
    draws INT NOT NULL DEFAULT 0,
    elo_rating INT NOT NULL DEFAULT 1000,
    coins INT NOT NULL DEFAULT 0,
    active_skin VARCHAR(50) NOT NULL DEFAULT 'default'
    );

CREATE TABLE IF NOT EXISTS inventory (
    user_id INT NOT NULL,
    item_id VARCHAR(50) NOT NULL,
    purchased_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, item_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);