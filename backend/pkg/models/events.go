package models

import (
	"database/sql"
	"errors"
	"time"
)

// CreateEvent inserts a group event.
func CreateEvent(db *sql.DB, id, groupID, creatorID, title, description string, eventTime time.Time) error {
	_, err := db.Exec(`INSERT INTO events (id, group_id, creator_id, title, description, event_time)
		VALUES (?, ?, ?, ?, ?, ?)`, id, groupID, creatorID, title, description, eventTime)
	return err
}

// RespondToEvent records or updates a user's going/not_going choice.
func RespondToEvent(db *sql.DB, id, eventID, userID, response string) error {
	_, err := db.Exec(`INSERT INTO event_responses (id, event_id, user_id, response)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(event_id, user_id) DO UPDATE SET response = excluded.response`,
		id, eventID, userID, response)
	return err
}

func scanEvent(s interface{ Scan(dest ...any) error }) (*Event, error) {
	var e Event
	var mine sql.NullString
	if err := s.Scan(&e.ID, &e.GroupID, &e.CreatorID, &e.Title, &e.Description, &e.EventTime,
		&e.CreatedAt, &e.GoingCount, &e.NotGoing, &mine); err != nil {
		return nil, err
	}
	if mine.Valid {
		e.MyResponse = &mine.String
	}
	return &e, nil
}

const eventSelect = `SELECT e.id, e.group_id, e.creator_id, e.title, e.description, e.event_time,
	e.created_at,
	(SELECT COUNT(*) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'going'),
	(SELECT COUNT(*) FROM event_responses r WHERE r.event_id = e.id AND r.response = 'not_going'),
	(SELECT response FROM event_responses r WHERE r.event_id = e.id AND r.user_id = ?)
	FROM events e`

// GetEvent returns a single event with vote tallies for the viewer.
func GetEvent(db *sql.DB, id, viewerID string) (*Event, error) {
	row := db.QueryRow(eventSelect+` WHERE e.id = ?`, viewerID, id)
	e, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return e, err
}

// ListGroupEvents returns events of a group with tallies for the viewer.
func ListGroupEvents(db *sql.DB, groupID, viewerID string) ([]*Event, error) {
	rows, err := db.Query(eventSelect+` WHERE e.group_id = ? ORDER BY e.event_time ASC`, viewerID, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []*Event{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
