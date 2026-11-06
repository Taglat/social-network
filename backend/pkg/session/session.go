// Package session implements cookie-based session authentication on top of the
// sessions table. A session id (UUID) is stored in an HttpOnly cookie and
// mapped to a user server-side.
package session

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"socialnetwork/pkg/models"
)

// CookieName is the name of the session cookie.
const CookieName = "session"

// Duration is how long a session stays valid.
const Duration = 7 * 24 * time.Hour

// ErrNoSession indicates the request carried no valid session.
var ErrNoSession = errors.New("no session")

// Create starts a new session for the user and writes the session cookie.
func Create(w http.ResponseWriter, db *sql.DB, userID string) error {
	id := uuid.NewString()
	expires := time.Now().Add(Duration)
	if err := models.CreateSession(db, id, userID, expires); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    id,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(Duration.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// Destroy ends the current session and clears the cookie.
func Destroy(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if c, err := r.Cookie(CookieName); err == nil {
		_ = models.DeleteSession(db, c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// CurrentUser returns the authenticated user for the request, or ErrNoSession.
func CurrentUser(r *http.Request, db *sql.DB) (*models.User, error) {
	c, err := r.Cookie(CookieName)
	if err != nil || c.Value == "" {
		return nil, ErrNoSession
	}
	u, err := models.GetSessionUser(db, c.Value)
	if errors.Is(err, models.ErrNotFound) {
		return nil, ErrNoSession
	}
	return u, err
}
