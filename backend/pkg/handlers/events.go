package handlers

import (
	"net/http"
	"strings"
	"time"

	"socialnetwork/pkg/models"
	"socialnetwork/pkg/utils"
)

// CreateEvent creates an event in a group and notifies all members.
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	groupID := r.PathValue("id")
	member, _ := models.IsGroupMember(h.DB, groupID, user.ID)
	if !member {
		utils.Error(w, http.StatusForbidden, "only members can create events")
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		EventTime   string `json:"eventTime"` // RFC3339 / datetime-local
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	body.Description = strings.TrimSpace(body.Description)
	if body.Title == "" || body.Description == "" || body.EventTime == "" {
		utils.Error(w, http.StatusBadRequest, "title, description and date/time are required")
		return
	}

	eventTime, err := parseTime(body.EventTime)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid event time")
		return
	}

	eventID := newID()
	if err := models.CreateEvent(h.DB, eventID, groupID, user.ID, body.Title, body.Description, eventTime); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	group, _ := models.GetGroup(h.DB, groupID, user.ID)
	memberIDs, _ := models.ListGroupMemberIDs(h.DB, groupID)
	for _, uid := range memberIDs {
		if uid == user.ID {
			continue
		}
		h.notify(uid, "event_created", strptr(user.ID), strptr(eventID),
			"New event \""+body.Title+"\" in "+group.Title)
	}

	event, _ := models.GetEvent(h.DB, eventID, user.ID)
	utils.JSON(w, http.StatusCreated, event)
}

// RespondEvent records the user's going/not_going choice for an event.
func (h *Handler) RespondEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	eventID := r.PathValue("id")
	event, err := models.GetEvent(h.DB, eventID, user.ID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "event not found")
		return
	}
	member, _ := models.IsGroupMember(h.DB, event.GroupID, user.ID)
	if !member {
		utils.Error(w, http.StatusForbidden, "only members can respond")
		return
	}

	var body struct {
		Response string `json:"response"` // "going" | "not_going"
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Response != "going" && body.Response != "not_going" {
		utils.Error(w, http.StatusBadRequest, "response must be going or not_going")
		return
	}
	if err := models.RespondToEvent(h.DB, newID(), eventID, user.ID, body.Response); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to respond")
		return
	}
	updated, _ := models.GetEvent(h.DB, eventID, user.ID)
	utils.JSON(w, http.StatusOK, updated)
}

// parseTime accepts both RFC3339 and the HTML datetime-local format.
func parseTime(s string) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02T15:04:05"}
	var lastErr error
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, lastErr
}
