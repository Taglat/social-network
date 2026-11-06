// Package handlers contains the HTTP handlers that implement the core logic of
// the social network (the "App"). They read/write through the models layer and
// push realtime events through the websocket hub.
package handlers

import (
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"socialnetwork/pkg/models"
	"socialnetwork/pkg/session"
	"socialnetwork/pkg/utils"
	"socialnetwork/pkg/ws"
)

// newID generates a new UUID string used as a primary key.
func newID() string { return uuid.NewString() }

// Handler bundles the dependencies shared by every HTTP handler.
type Handler struct {
	DB        *sql.DB
	Hub       *ws.Hub
	UploadDir string
}

// New builds a Handler.
func New(db *sql.DB, hub *ws.Hub, uploadDir string) *Handler {
	return &Handler{DB: db, Hub: hub, UploadDir: uploadDir}
}

// currentUser returns the authenticated user or writes a 401 and returns false.
func (h *Handler) currentUser(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	u, err := session.CurrentUser(r, h.DB)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "not authenticated")
		return nil, false
	}
	return u, true
}

// notify creates a notification and pushes it live to the recipient.
func (h *Handler) notify(userID, ntype string, actorID, entityID *string, message string) {
	n, err := models.CreateNotification(h.DB, newID(), userID, ntype, actorID, entityID, message)
	if err != nil {
		return
	}
	// A freshly created request-type notification is by definition actionable.
	switch ntype {
	case "follow_request", "group_invite", "group_join_request":
		n.Actionable = true
	}
	h.Hub.SendJSON(userID, map[string]any{"type": "notification", "notification": n})
}

// strptr returns a pointer to s (helper for optional notification fields).
func strptr(s string) *string { return &s }
