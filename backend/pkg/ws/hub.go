// Package ws implements the realtime layer: a hub that tracks connected users,
// relays private and group chat messages, pushes notifications, and broadcasts
// online/offline presence to a user's followers.
package ws

import (
	"database/sql"
	"encoding/json"
	"sync"

	"socialnetwork/pkg/models"
)

// Hub maintains the set of active clients and routes messages between them.
type Hub struct {
	db         *sql.DB
	mu         sync.RWMutex
	clients    map[string]map[*Client]bool // userID -> set of connections
	register   chan *Client
	unregister chan *Client
}

// NewHub creates a hub bound to the given database.
func NewHub(db *sql.DB) *Hub {
	return &Hub{
		db:         db,
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run processes register/unregister events. Call it in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			first := len(h.clients[c.userID]) == 0
			if h.clients[c.userID] == nil {
				h.clients[c.userID] = make(map[*Client]bool)
			}
			h.clients[c.userID][c] = true
			h.mu.Unlock()

			c.send <- h.onlineUsersPayload()
			if first {
				h.broadcastPresence(c.userID, true)
			}
		case c := <-h.unregister:
			h.mu.Lock()
			last := false
			if conns, ok := h.clients[c.userID]; ok {
				if _, ok := conns[c]; ok {
					delete(conns, c)
					close(c.send)
				}
				if len(conns) == 0 {
					delete(h.clients, c.userID)
					last = true
				}
			}
			h.mu.Unlock()
			if last {
				h.broadcastPresence(c.userID, false)
			}
		}
	}
}

// IsOnline reports whether a user has at least one active connection.
func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID]) > 0
}

// onlineUserIDs returns the ids of all currently connected users.
func (h *Hub) onlineUserIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]string, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}

// SendToUser delivers a raw JSON payload to every connection of a user.
// The sends happen while holding the read lock: send channels are only closed
// while the write lock is held (in Run), so a send can never race the close
// and panic. Sends are non-blocking, so holding the lock here is cheap.
func (h *Hub) SendToUser(userID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[userID] {
		select {
		case c.send <- payload:
		default:
		}
	}
}

// SendJSON marshals v and sends it to a user.
func (h *Hub) SendJSON(userID string, v any) {
	if b, err := json.Marshal(v); err == nil {
		h.SendToUser(userID, b)
	}
}

func (h *Hub) onlineUsersPayload() []byte {
	b, _ := json.Marshal(map[string]any{
		"type":    "online_users",
		"userIds": h.onlineUserIDs(),
	})
	return b
}

// broadcastPresence informs a user's followers and followed users that they
// went online/offline so chat lists update in realtime.
func (h *Hub) broadcastPresence(userID string, online bool) {
	payload, _ := json.Marshal(map[string]any{
		"type":   "presence",
		"userId": userID,
		"online": online,
	})

	recipients := map[string]bool{}
	if followers, err := models.ListFollowers(h.db, userID); err == nil {
		for _, u := range followers {
			recipients[u.ID] = true
		}
	}
	if following, err := models.ListFollowing(h.db, userID); err == nil {
		for _, u := range following {
			recipients[u.ID] = true
		}
	}
	for id := range recipients {
		h.SendToUser(id, payload)
	}
}
