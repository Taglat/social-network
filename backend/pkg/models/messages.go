package models

import "database/sql"

// CreatePrivateMessage stores a direct message between two users.
func CreatePrivateMessage(db *sql.DB, id, senderID, receiverID, content string) error {
	_, err := db.Exec(`INSERT INTO messages (id, sender_id, receiver_id, content) VALUES (?, ?, ?, ?)`,
		id, senderID, receiverID, content)
	return err
}

// CreateGroupMessage stores a message in a group chat room.
func CreateGroupMessage(db *sql.DB, id, senderID, groupID, content string) error {
	_, err := db.Exec(`INSERT INTO messages (id, sender_id, group_id, content) VALUES (?, ?, ?, ?)`,
		id, senderID, groupID, content)
	return err
}

func scanMessages(rows *sql.Rows) ([]*Message, error) {
	defer rows.Close()
	msgs := []*Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.GroupID, &m.Content,
			&m.CreatedAt, &m.SenderName); err != nil {
			return nil, err
		}
		msgs = append(msgs, &m)
	}
	return msgs, rows.Err()
}

const messageSelect = `SELECT m.id, m.sender_id, m.receiver_id, m.group_id, m.content,
	m.created_at, (u.first_name || ' ' || u.last_name) AS sender_name
	FROM messages m JOIN users u ON u.id = m.sender_id`

// ListPrivateMessages returns the conversation between two users, oldest first.
func ListPrivateMessages(db *sql.DB, a, b string) ([]*Message, error) {
	rows, err := db.Query(messageSelect+`
		WHERE m.group_id IS NULL AND (
			(m.sender_id = ? AND m.receiver_id = ?) OR
			(m.sender_id = ? AND m.receiver_id = ?))
		ORDER BY m.created_at ASC`, a, b, b, a)
	if err != nil {
		return nil, err
	}
	return scanMessages(rows)
}

// ListGroupMessages returns the messages of a group chat, oldest first.
func ListGroupMessages(db *sql.DB, groupID string) ([]*Message, error) {
	rows, err := db.Query(messageSelect+` WHERE m.group_id = ? ORDER BY m.created_at ASC`, groupID)
	if err != nil {
		return nil, err
	}
	return scanMessages(rows)
}
