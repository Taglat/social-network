package handlers

import (
	"net/http"

	"socialnetwork/pkg/models"
	"socialnetwork/pkg/utils"
)

// ListNotifications returns the current user's notifications, newest first.
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	notes, err := models.ListNotifications(h.DB, user.ID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	h.annotateActionable(notes)
	utils.JSON(w, http.StatusOK, notes)
}

// annotateActionable marks request-type notifications whose accept/decline
// actions would still succeed. A follow request that was already accepted (or
// a group invite/request already resolved) stays in the list as history, but
// the frontend must not render its action buttons again.
func (h *Handler) annotateActionable(notes []*models.Notification) {
	for _, n := range notes {
		switch n.Type {
		case "follow_request":
			if n.ActorID != nil {
				if f, err := models.GetFollow(h.DB, *n.ActorID, n.UserID); err == nil && f.Status == "pending" {
					n.Actionable = true
				}
			}
		case "group_invite":
			if n.EntityID != nil {
				if status, err := models.GetMembership(h.DB, *n.EntityID, n.UserID); err == nil && status == "invited" {
					n.Actionable = true
				}
			}
		case "group_join_request":
			if n.EntityID != nil && n.ActorID != nil {
				if status, err := models.GetMembership(h.DB, *n.EntityID, *n.ActorID); err == nil && status == "requested" {
					n.Actionable = true
				}
			}
		}
	}
}

// MarkNotificationRead flags a single notification as read.
func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if err := models.MarkNotificationRead(h.DB, r.PathValue("id"), user.ID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "read"})
}

// MarkAllNotificationsRead flags all of the user's notifications as read.
func (h *Handler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if err := models.MarkAllNotificationsRead(h.DB, user.ID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "all read"})
}
