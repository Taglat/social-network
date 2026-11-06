package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"socialnetwork/pkg/models"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Same backend serves trusted first-party clients (web app + desktop app).
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client is a single websocket connection belonging to a user.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID string
	send   chan []byte
}

// incoming is the shape of messages sent by clients.
type incoming struct {
	Type    string `json:"type"`
	To      string `json:"to"`      // recipient user id (private message)
	GroupID string `json:"groupId"` // group id (group message)
	Content string `json:"content"`
}

// HandleWS upgrades an HTTP request to a websocket and registers the client.
// The caller must have already authenticated the user.
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	c := &Client{hub: h, conn: conn, userID: userID, send: make(chan []byte, 64)}
	h.register <- c
	go c.writePump()
	go c.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(8192)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var msg incoming
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "private_message":
			c.handlePrivate(msg)
		case "group_message":
			c.handleGroup(msg)
		}
	}
}

func (c *Client) handlePrivate(msg incoming) {
	if msg.To == "" || msg.Content == "" {
		return
	}
	connected, err := models.AreConnected(c.hub.db, c.userID, msg.To)
	if err != nil || !connected {
		c.sendError("You can only message users you follow or who follow you")
		return
	}
	id := uuid.NewString()
	if err := models.CreatePrivateMessage(c.hub.db, id, c.userID, msg.To, msg.Content); err != nil {
		c.sendError("Failed to send message")
		return
	}
	full, err := models.ListPrivateMessages(c.hub.db, c.userID, msg.To)
	if err != nil || len(full) == 0 {
		return
	}
	out := envelope("message", "message", full[len(full)-1])
	c.hub.SendToUser(msg.To, out)
	c.hub.SendToUser(c.userID, out) // echo so all sender's devices stay in sync
}

func (c *Client) handleGroup(msg incoming) {
	if msg.GroupID == "" || msg.Content == "" {
		return
	}
	member, err := models.IsGroupMember(c.hub.db, msg.GroupID, c.userID)
	if err != nil || !member {
		c.sendError("You are not a member of this group")
		return
	}
	id := uuid.NewString()
	if err := models.CreateGroupMessage(c.hub.db, id, c.userID, msg.GroupID, msg.Content); err != nil {
		c.sendError("Failed to send message")
		return
	}
	all, err := models.ListGroupMessages(c.hub.db, msg.GroupID)
	if err != nil || len(all) == 0 {
		return
	}
	out := envelope("message", "message", all[len(all)-1])
	memberIDs, _ := models.ListGroupMemberIDs(c.hub.db, msg.GroupID)
	for _, uid := range memberIDs {
		c.hub.SendToUser(uid, out)
	}
}

func (c *Client) sendError(message string) {
	b, _ := json.Marshal(map[string]any{"type": "error", "error": message})
	select {
	case c.send <- b:
	default:
	}
}

// envelope wraps a payload value under {"type": t, key: v}.
func envelope(t, key string, v any) []byte {
	b, _ := json.Marshal(map[string]any{"type": t, key: v})
	return b
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
