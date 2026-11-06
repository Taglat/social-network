package models

import (
	"database/sql"
	"errors"
	"time"
)

// CreateSession stores a new session token for a user.
func CreateSession(db *sql.DB, id, userID string, expiresAt time.Time) error {
	_, err := db.Exec(`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		id, userID, expiresAt)
	return err
}

// GetSessionUser returns the user owning a non-expired session, or ErrNotFound.
func GetSessionUser(db *sql.DB, sessionID string) (*User, error) {
	row := db.QueryRow(`SELECT `+userColumns+` FROM users
		WHERE id = (SELECT user_id FROM sessions WHERE id = ? AND expires_at > CURRENT_TIMESTAMP)`,
		sessionID)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

// DeleteSession removes a session (logout).
func DeleteSession(db *sql.DB, sessionID string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	return err
}

// DeleteExpiredSessions purges expired sessions (housekeeping).
func DeleteExpiredSessions(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP`)
	return err
}
