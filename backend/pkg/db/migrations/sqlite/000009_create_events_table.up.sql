CREATE TABLE IF NOT EXISTS events (
    id           TEXT PRIMARY KEY,
    group_id     TEXT NOT NULL,
    creator_id   TEXT NOT NULL,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL,
    event_time   DATETIME NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_events_group ON events(group_id);
