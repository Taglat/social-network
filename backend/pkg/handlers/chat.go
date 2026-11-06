package handlers

import (
	"net/http"

	"socialnetwork/pkg/models"
	"socialnetwork/pkg/session"
	"socialnetwork/pkg/utils"
)

// ServeWS authenticates the request and upgrades it to a websocket connection.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	user, err := session.CurrentUser(r, h.DB)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	h.Hub.HandleWS(w, r, user.ID)
}

// GetPrivateMessages returns the conversation between the current user and the
// user identified in the path. Allowed only when they are connected.
func (h *Handler) GetPrivateMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	otherID := r.PathValue("id")
	connected, _ := models.AreConnected(h.DB, user.ID, otherID)
	if !connected {
		utils.Error(w, http.StatusForbidden, "you can only chat with users you follow or who follow you")
		return
	}
	msgs, err := models.ListPrivateMessages(h.DB, user.ID, otherID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	utils.JSON(w, http.StatusOK, msgs)
}

// GetGroupMessages returns the chat history of a group (members only).
func (h *Handler) GetGroupMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	groupID := r.PathValue("id")
	member, _ := models.IsGroupMember(h.DB, groupID, user.ID)
	if !member {
		utils.Error(w, http.StatusForbidden, "only members can read the group chat")
		return
	}
	msgs, err := models.ListGroupMessages(h.DB, groupID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	utils.JSON(w, http.StatusOK, msgs)
}

// ChatPartners returns users the current user may privately chat with
// (people they follow or who follow them), with online status.
func (h *Handler) ChatPartners(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	partners, err := models.ListChatPartners(h.DB, user.ID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	type partnerStatus struct {
		*models.User
		Online bool `json:"online"`
	}
	out := make([]partnerStatus, 0, len(partners))
	for _, u := range partners {
		out = append(out, partnerStatus{User: u, Online: h.Hub.IsOnline(u.ID)})
	}
	utils.JSON(w, http.StatusOK, out)
}
