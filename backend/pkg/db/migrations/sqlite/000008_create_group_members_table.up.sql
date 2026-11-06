CREATE TABLE IF NOT EXISTS group_members (
    id          TEXT PRIMARY KEY,
    group_id    TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'requested',
    invited_by  TEXT,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_group_members_group ON group_members(group_id);
CREATE INDEX IF NOT EXISTS idx_group_members_user ON group_members(user_id);
