CREATE TABLE IF NOT EXISTS event_responses (
    id        TEXT PRIMARY KEY,
    event_id  TEXT NOT NULL,
    user_id   TEXT NOT NULL,
    response  TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (event_id, user_id),
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
