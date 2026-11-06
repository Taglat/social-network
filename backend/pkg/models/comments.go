package models

import "database/sql"

// CreateComment inserts a comment on a post. image may be nil.
func CreateComment(db *sql.DB, id, postID, userID, content string, image *string) error {
	_, err := db.Exec(`INSERT INTO comments (id, post_id, user_id, content, image_path)
		VALUES (?, ?, ?, ?, ?)`, id, postID, userID, content, image)
	return err
}

// ListComments returns the comments on a post, oldest first, with author info.
func ListComments(db *sql.DB, postID string) ([]*Comment, error) {
	rows, err := db.Query(`SELECT c.id, c.post_id, c.user_id, c.content, c.image_path,
		c.created_at, (u.first_name || ' ' || u.last_name) AS author_name, u.avatar_path
		FROM comments c JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ? ORDER BY c.created_at ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	comments := []*Comment{}
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.Image,
			&c.CreatedAt, &c.AuthorName, &c.AuthorAvatar); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}
	return comments, rows.Err()
}
