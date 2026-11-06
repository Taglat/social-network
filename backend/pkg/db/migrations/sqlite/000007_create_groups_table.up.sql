CREATE TABLE IF NOT EXISTS groups (
    id           TEXT PRIMARY KEY,
    creator_id   TEXT NOT NULL,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);
