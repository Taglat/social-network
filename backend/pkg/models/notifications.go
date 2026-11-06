package models

import "database/sql"

// CreateNotification stores a notification for a user and returns it (so the
// caller can also push it live over the websocket).
func CreateNotification(db *sql.DB, id, userID, ntype string, actorID, entityID *string, message string) (*Notification, error) {
	_, err := db.Exec(`INSERT INTO notifications (id, user_id, type, actor_id, entity_id, message)
		VALUES (?, ?, ?, ?, ?, ?)`, id, userID, ntype, actorID, entityID, message)
	if err != nil {
		return nil, err
	}
	return GetNotification(db, id)
}

// GetNotification fetches a single notification by id.
func GetNotification(db *sql.DB, id string) (*Notification, error) {
	var n Notification
	err := db.QueryRow(`SELECT id, user_id, type, actor_id, entity_id, message, is_read, created_at
		FROM notifications WHERE id = ?`, id).
		Scan(&n.ID, &n.UserID, &n.Type, &n.ActorID, &n.EntityID, &n.Message, &n.IsRead, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// ListNotifications returns a user's notifications, newest first.
func ListNotifications(db *sql.DB, userID string) ([]*Notification, error) {
	rows, err := db.Query(`SELECT id, user_id, type, actor_id, entity_id, message, is_read, created_at
		FROM notifications WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notes := []*Notification{}
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.ActorID, &n.EntityID,
			&n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, &n)
	}
	return notes, rows.Err()
}

// MarkNotificationRead flags a single notification as read for its owner.
func MarkNotificationRead(db *sql.DB, id, userID string) error {
	_, err := db.Exec(`UPDATE notifications SET is_read = 1 WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

// MarkAllNotificationsRead flags all of a user's notifications as read.
func MarkAllNotificationsRead(db *sql.DB, userID string) error {
	_, err := db.Exec(`UPDATE notifications SET is_read = 1 WHERE user_id = ?`, userID)
	return err
}
