// Package models contains the data structures and database access functions
// (repositories) for every entity in the social network. It is the lowest
// layer of the App and depends only on database/sql and the standard library.
package models

import "time"

// User mirrors a row of the users table. PasswordHash is never serialized.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	DateOfBirth  string    `json:"dateOfBirth"`
	Avatar       *string   `json:"avatar"`
	Nickname     *string   `json:"nickname"`
	AboutMe      *string   `json:"aboutMe"`
	IsPublic     bool      `json:"isPublic"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Follower represents a follow relationship with its status (pending/accepted).
type Follower struct {
	ID          string    `json:"id"`
	FollowerID  string    `json:"followerId"`
	FollowingID string    `json:"followingId"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Post mirrors a row of the posts table, enriched with author info for display.
type Post struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	GroupID      *string   `json:"groupId"`
	Content      string    `json:"content"`
	Image        *string   `json:"image"`
	Privacy      string    `json:"privacy"`
	CreatedAt    time.Time `json:"createdAt"`
	AuthorName   string    `json:"authorName"`
	AuthorAvatar *string   `json:"authorAvatar"`
	CommentCount int       `json:"commentCount"`
}

// Comment mirrors a row of the comments table with author info.
type Comment struct {
	ID           string    `json:"id"`
	PostID       string    `json:"postId"`
	UserID       string    `json:"userId"`
	Content      string    `json:"content"`
	Image        *string   `json:"image"`
	CreatedAt    time.Time `json:"createdAt"`
	AuthorName   string    `json:"authorName"`
	AuthorAvatar *string   `json:"authorAvatar"`
}

// Group mirrors a row of the groups table.
type Group struct {
	ID          string    `json:"id"`
	CreatorID   string    `json:"creatorId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	MemberCount int       `json:"memberCount"`
	Membership  string    `json:"membership"` // "", "member", "invited", "requested"
}

// Event mirrors a row of the events table, with vote tallies for display.
type Event struct {
	ID           string    `json:"id"`
	GroupID      string    `json:"groupId"`
	CreatorID    string    `json:"creatorId"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	EventTime    time.Time `json:"eventTime"`
	CreatedAt    time.Time `json:"createdAt"`
	GoingCount   int       `json:"goingCount"`
	NotGoing     int       `json:"notGoingCount"`
	MyResponse   *string   `json:"myResponse"`
}

// Message mirrors a row of the messages table (private or group).
type Message struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"senderId"`
	ReceiverID *string   `json:"receiverId"`
	GroupID    *string   `json:"groupId"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
	SenderName string    `json:"senderName"`
}

// Notification mirrors a row of the notifications table. Actionable is not
// stored: it is computed when listing (or pushing) a notification and tells
// the frontend whether the accept/decline actions would still succeed.
type Notification struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Type       string    `json:"type"`
	ActorID    *string   `json:"actorId"`
	EntityID   *string   `json:"entityId"`
	Message    string    `json:"message"`
	IsRead     bool      `json:"isRead"`
	CreatedAt  time.Time `json:"createdAt"`
	Actionable bool      `json:"actionable"`
}
