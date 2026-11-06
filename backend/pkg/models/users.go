package models

import (
	"database/sql"
	"errors"
)

// ErrNotFound is returned when a lookup matches no row.
var ErrNotFound = errors.New("not found")

const userColumns = `id, email, password_hash, first_name, last_name,
	date_of_birth, avatar_path, nickname, about_me, is_public, created_at`

func scanUser(s interface {
	Scan(dest ...any) error
}) (*User, error) {
	var u User
	if err := s.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName,
		&u.DateOfBirth, &u.Avatar, &u.Nickname, &u.AboutMe, &u.IsPublic, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// CreateUser inserts a new user row.
func CreateUser(db *sql.DB, u *User) error {
	_, err := db.Exec(`INSERT INTO users (id, email, password_hash, first_name, last_name,
		date_of_birth, avatar_path, nickname, about_me, is_public)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.DateOfBirth,
		u.Avatar, u.Nickname, u.AboutMe, u.IsPublic)
	return err
}

// GetUserByEmail returns the user with the given email, or ErrNotFound.
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	row := db.QueryRow(`SELECT `+userColumns+` FROM users WHERE email = ?`, email)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

// GetUserByID returns the user with the given id, or ErrNotFound.
func GetUserByID(db *sql.DB, id string) (*User, error) {
	row := db.QueryRow(`SELECT `+userColumns+` FROM users WHERE id = ?`, id)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

// EmailExists reports whether a user with the given email already exists.
func EmailExists(db *sql.DB, email string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`, email).Scan(&exists)
	return exists, err
}

// SetPrivacy updates a user's public/private flag.
func SetPrivacy(db *sql.DB, userID string, isPublic bool) error {
	_, err := db.Exec(`UPDATE users SET is_public = ? WHERE id = ?`, isPublic, userID)
	return err
}

// ListUsers returns all users except the one given (for "start a chat" pickers).
func ListUsers(db *sql.DB, exceptID string) ([]*User, error) {
	rows, err := db.Query(`SELECT `+userColumns+` FROM users WHERE id != ? ORDER BY first_name`, exceptID)
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
