package models

import (
	"database/sql"
	"errors"
)

const postSelect = `SELECT p.id, p.user_id, p.group_id, p.content, p.image_path, p.privacy,
	p.created_at, (u.first_name || ' ' || u.last_name) AS author_name, u.avatar_path,
	(SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id) AS comment_count
	FROM posts p JOIN users u ON u.id = p.user_id`

func scanPost(s interface{ Scan(dest ...any) error }) (*Post, error) {
	var p Post
	if err := s.Scan(&p.ID, &p.UserID, &p.GroupID, &p.Content, &p.Image, &p.Privacy,
		&p.CreatedAt, &p.AuthorName, &p.AuthorAvatar, &p.CommentCount); err != nil {
		return nil, err
	}
	return &p, nil
}

func scanPosts(rows *sql.Rows) ([]*Post, error) {
	defer rows.Close()
	posts := []*Post{}
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

// CreatePost inserts a post. groupID and image may be nil.
func CreatePost(db *sql.DB, id, userID string, groupID *string, content string, image *string, privacy string) error {
	_, err := db.Exec(`INSERT INTO posts (id, user_id, group_id, content, image_path, privacy)
		VALUES (?, ?, ?, ?, ?, ?)`, id, userID, groupID, content, image, privacy)
	return err
}

// AddAllowedViewer grants a specific user permission to view a private post.
func AddAllowedViewer(db *sql.DB, postID, userID string) error {
	_, err := db.Exec(`INSERT OR IGNORE INTO post_allowed_users (post_id, user_id) VALUES (?, ?)`,
		postID, userID)
	return err
}

// GetPost fetches a single post (without applying privacy rules).
func GetPost(db *sql.DB, id string) (*Post, error) {
	row := db.QueryRow(postSelect+` WHERE p.id = ?`, id)
	p, err := scanPost(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

// CanViewPost reports whether viewerID is allowed to see the given post.
func CanViewPost(db *sql.DB, viewerID string, p *Post) (bool, error) {
	if p.UserID == viewerID {
		return true, nil
	}
	if p.GroupID != nil {
		return IsGroupMember(db, *p.GroupID, viewerID)
	}
	switch p.Privacy {
	case "public":
		return true, nil
	case "almost_private":
		return IsAcceptedFollower(db, viewerID, p.UserID)
	case "private":
		var ok bool
		err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM post_allowed_users WHERE post_id = ? AND user_id = ?)`,
			p.ID, viewerID).Scan(&ok)
		return ok, err
	}
	return false, nil
}

// ListFeed returns non-group posts visible to the viewer, newest first.
func ListFeed(db *sql.DB, viewerID string) ([]*Post, error) {
	rows, err := db.Query(postSelect+`
		WHERE p.group_id IS NULL AND (
			p.user_id = ?
			OR p.privacy = 'public'
			OR (p.privacy = 'almost_private' AND EXISTS(
				SELECT 1 FROM followers f WHERE f.follower_id = ? AND f.following_id = p.user_id AND f.status = 'accepted'))
			OR (p.privacy = 'private' AND EXISTS(
				SELECT 1 FROM post_allowed_users pa WHERE pa.post_id = p.id AND pa.user_id = ?))
		)
		ORDER BY p.created_at DESC`, viewerID, viewerID, viewerID)
	if err != nil {
		return nil, err
	}
	return scanPosts(rows)
}

// ListUserPosts returns posts authored by authorID that viewerID may see.
func ListUserPosts(db *sql.DB, authorID, viewerID string) ([]*Post, error) {
	rows, err := db.Query(postSelect+`
		WHERE p.group_id IS NULL AND p.user_id = ? AND (
			? = p.user_id
			OR p.privacy = 'public'
			OR (p.privacy = 'almost_private' AND EXISTS(
				SELECT 1 FROM followers f WHERE f.follower_id = ? AND f.following_id = p.user_id AND f.status = 'accepted'))
			OR (p.privacy = 'private' AND EXISTS(
				SELECT 1 FROM post_allowed_users pa WHERE pa.post_id = p.id AND pa.user_id = ?))
		)
		ORDER BY p.created_at DESC`, authorID, viewerID, viewerID, viewerID)
	if err != nil {
		return nil, err
	}
	return scanPosts(rows)
}

// ListGroupPosts returns posts belonging to a group, newest first.
func ListGroupPosts(db *sql.DB, groupID string) ([]*Post, error) {
	rows, err := db.Query(postSelect+` WHERE p.group_id = ? ORDER BY p.created_at DESC`, groupID)
	if err != nil {
		return nil, err
	}
	return scanPosts(rows)
}
