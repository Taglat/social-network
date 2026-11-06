package models

import (
	"database/sql"
	"errors"
)

// UpsertFollow creates or updates a follow row with the given status and
// returns whether a brand-new relationship row was inserted.
func UpsertFollow(db *sql.DB, id, followerID, followingID, status string) error {
	_, err := db.Exec(`INSERT INTO followers (id, follower_id, following_id, status)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(follower_id, following_id) DO UPDATE SET status = excluded.status`,
		id, followerID, followingID, status)
	return err
}

// GetFollow returns the follow relationship between two users, or ErrNotFound.
func GetFollow(db *sql.DB, followerID, followingID string) (*Follower, error) {
	var f Follower
	err := db.QueryRow(`SELECT id, follower_id, following_id, status, created_at
		FROM followers WHERE follower_id = ? AND following_id = ?`, followerID, followingID).
		Scan(&f.ID, &f.FollowerID, &f.FollowingID, &f.Status, &f.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// SetFollowStatus updates the status of a follow request (accept/decline).
func SetFollowStatus(db *sql.DB, followerID, followingID, status string) error {
	_, err := db.Exec(`UPDATE followers SET status = ? WHERE follower_id = ? AND following_id = ?`,
		status, followerID, followingID)
	return err
}

// Unfollow removes a follow relationship.
func Unfollow(db *sql.DB, followerID, followingID string) error {
	_, err := db.Exec(`DELETE FROM followers WHERE follower_id = ? AND following_id = ?`,
		followerID, followingID)
	return err
}

// IsAcceptedFollower reports whether followerID follows followingID (accepted).
func IsAcceptedFollower(db *sql.DB, followerID, followingID string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM followers
		WHERE follower_id = ? AND following_id = ? AND status = 'accepted')`,
		followerID, followingID).Scan(&exists)
	return exists, err
}

// AreConnected reports whether at least one of the two users follows the other
// (the requirement for being allowed to chat privately).
func AreConnected(db *sql.DB, a, b string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM followers
		WHERE status = 'accepted' AND (
			(follower_id = ? AND following_id = ?) OR
			(follower_id = ? AND following_id = ?)))`,
		a, b, b, a).Scan(&exists)
	return exists, err
}

// ListFollowers returns the users who follow the given user (accepted).
func ListFollowers(db *sql.DB, userID string) ([]*User, error) {
	return queryUsers(db, `SELECT `+prefixed("u")+` FROM users u
		JOIN followers f ON f.follower_id = u.id
		WHERE f.following_id = ? AND f.status = 'accepted' ORDER BY u.first_name`, userID)
}

// ListFollowing returns the users the given user follows (accepted).
func ListFollowing(db *sql.DB, userID string) ([]*User, error) {
	return queryUsers(db, `SELECT `+prefixed("u")+` FROM users u
		JOIN followers f ON f.following_id = u.id
		WHERE f.follower_id = ? AND f.status = 'accepted' ORDER BY u.first_name`, userID)
}

// ListPendingFollowRequests returns users who have a pending request to follow
// the given user.
func ListPendingFollowRequests(db *sql.DB, userID string) ([]*User, error) {
	return queryUsers(db, `SELECT `+prefixed("u")+` FROM users u
		JOIN followers f ON f.follower_id = u.id
		WHERE f.following_id = ? AND f.status = 'pending' ORDER BY u.first_name`, userID)
}

// ListChatPartners returns users the given user may privately chat with: those
// they follow or who follow them (at least one accepted relationship).
func ListChatPartners(db *sql.DB, userID string) ([]*User, error) {
	return queryUsers(db, `SELECT DISTINCT `+prefixed("u")+` FROM users u
		JOIN followers f ON (
			(f.follower_id = ? AND f.following_id = u.id) OR
			(f.following_id = ? AND f.follower_id = u.id))
		WHERE f.status = 'accepted' AND u.id != ? ORDER BY u.first_name`, userID, userID, userID)
}

// prefixed returns the user column list prefixed with a table alias.
func prefixed(alias string) string {
	return alias + ".id, " + alias + ".email, " + alias + ".password_hash, " +
		alias + ".first_name, " + alias + ".last_name, " + alias + ".date_of_birth, " +
		alias + ".avatar_path, " + alias + ".nickname, " + alias + ".about_me, " +
		alias + ".is_public, " + alias + ".created_at"
}

func queryUsers(db *sql.DB, query string, args ...any) ([]*User, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []*User{} // non-nil so it serializes as [] not null
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
