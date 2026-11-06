package models

import (
	"database/sql"
	"errors"
)

// CreateGroup inserts a group and adds the creator as a member.
func CreateGroup(db *sql.DB, id, creatorID, title, description, memberRowID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO groups (id, creator_id, title, description) VALUES (?, ?, ?, ?)`,
		id, creatorID, title, description); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO group_members (id, group_id, user_id, status) VALUES (?, ?, ?, 'member')`,
		memberRowID, id, creatorID); err != nil {
		return err
	}
	return tx.Commit()
}

// GetGroup returns a group with member count and the viewer's membership status.
func GetGroup(db *sql.DB, id, viewerID string) (*Group, error) {
	var g Group
	var membership sql.NullString
	err := db.QueryRow(`SELECT g.id, g.creator_id, g.title, g.description, g.created_at,
		(SELECT COUNT(*) FROM group_members m WHERE m.group_id = g.id AND m.status = 'member'),
		(SELECT status FROM group_members m WHERE m.group_id = g.id AND m.user_id = ?)
		FROM groups g WHERE g.id = ?`, viewerID, id).
		Scan(&g.ID, &g.CreatorID, &g.Title, &g.Description, &g.CreatedAt, &g.MemberCount, &membership)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if membership.Valid {
		g.Membership = membership.String
	}
	return &g, nil
}

// ListGroups returns all groups with the viewer's membership status.
func ListGroups(db *sql.DB, viewerID string) ([]*Group, error) {
	rows, err := db.Query(`SELECT g.id, g.creator_id, g.title, g.description, g.created_at,
		(SELECT COUNT(*) FROM group_members m WHERE m.group_id = g.id AND m.status = 'member'),
		(SELECT status FROM group_members m WHERE m.group_id = g.id AND m.user_id = ?)
		FROM groups g ORDER BY g.created_at DESC`, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := []*Group{}
	for rows.Next() {
		var g Group
		var membership sql.NullString
		if err := rows.Scan(&g.ID, &g.CreatorID, &g.Title, &g.Description, &g.CreatedAt,
			&g.MemberCount, &membership); err != nil {
			return nil, err
		}
		if membership.Valid {
			g.Membership = membership.String
		}
		groups = append(groups, &g)
	}
	return groups, rows.Err()
}

// IsGroupMember reports whether the user is an accepted member of the group.
func IsGroupMember(db *sql.DB, groupID, userID string) (bool, error) {
	var ok bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'member')`, groupID, userID).Scan(&ok)
	return ok, err
}

// GetMembership returns the membership status row for a user in a group.
func GetMembership(db *sql.DB, groupID, userID string) (string, error) {
	var status string
	err := db.QueryRow(`SELECT status FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return status, err
}

// UpsertMember inserts or updates a membership row with the given status.
func UpsertMember(db *sql.DB, id, groupID, userID, status string, invitedBy *string) error {
	_, err := db.Exec(`INSERT INTO group_members (id, group_id, user_id, status, invited_by)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(group_id, user_id) DO UPDATE SET status = excluded.status, invited_by = excluded.invited_by`,
		id, groupID, userID, status, invitedBy)
	return err
}

// SetMemberStatus updates the status of a membership row.
func SetMemberStatus(db *sql.DB, groupID, userID, status string) error {
	_, err := db.Exec(`UPDATE group_members SET status = ? WHERE group_id = ? AND user_id = ?`,
		status, groupID, userID)
	return err
}

// RemoveMember deletes a membership row (decline invite/request).
func RemoveMember(db *sql.DB, groupID, userID string) error {
	_, err := db.Exec(`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`, groupID, userID)
	return err
}

// ListGroupMembers returns the accepted members of a group.
func ListGroupMembers(db *sql.DB, groupID string) ([]*User, error) {
	return queryUsers(db, `SELECT `+prefixed("u")+` FROM users u
		JOIN group_members m ON m.user_id = u.id
		WHERE m.group_id = ? AND m.status = 'member' ORDER BY u.first_name`, groupID)
}

// ListGroupRequesters returns users with a pending request to join the group.
func ListGroupRequesters(db *sql.DB, groupID string) ([]*User, error) {
	return queryUsers(db, `SELECT `+prefixed("u")+` FROM users u
		JOIN group_members m ON m.user_id = u.id
		WHERE m.group_id = ? AND m.status = 'requested' ORDER BY u.first_name`, groupID)
}

// ListGroupMemberIDs returns the ids of accepted members (for notifications/broadcast).
func ListGroupMemberIDs(db *sql.DB, groupID string) ([]string, error) {
	rows, err := db.Query(`SELECT user_id FROM group_members WHERE group_id = ? AND status = 'member'`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
