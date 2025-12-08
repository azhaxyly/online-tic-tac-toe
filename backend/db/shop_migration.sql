-- Add coins and active_skin to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS coins INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS active_skin VARCHAR(50) NOT NULL DEFAULT 'default';

-- Create inventory table
CREATE TABLE IF NOT EXISTS inventory (
    user_id INT NOT NULL,
    item_id VARCHAR(50) NOT NULL,
    purchased_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, item_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
