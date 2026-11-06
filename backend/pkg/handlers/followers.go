package handlers

import (
	"net/http"

	"socialnetwork/pkg/models"
	"socialnetwork/pkg/utils"
)

// Follow sends a follow request (private target) or follows immediately
// (public target).
func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
	viewer, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	targetID := r.PathValue("id")
	if targetID == viewer.ID {
		utils.Error(w, http.StatusBadRequest, "you cannot follow yourself")
		return
	}
	target, err := models.GetUserByID(h.DB, targetID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "user not found")
		return
	}

	// If a relationship already exists (accepted follow or pending request),
	// keep it as-is: this must not downgrade an accepted follow back to pending
	// (e.g. after the target switched to private) or re-send notifications.
	if existing, err := models.GetFollow(h.DB, viewer.ID, targetID); err == nil {
		utils.JSON(w, http.StatusOK, map[string]string{"status": existing.Status})
		return
	}

	status := "pending"
	if target.IsPublic {
		status = "accepted"
	}
	if err := models.UpsertFollow(h.DB, newID(), viewer.ID, targetID, status); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to follow")
		return
	}

	if status == "pending" {
		h.notify(targetID, "follow_request", strptr(viewer.ID), strptr(viewer.ID),
			viewer.FirstName+" "+viewer.LastName+" wants to follow you")
	} else {
		h.notify(targetID, "new_follower", strptr(viewer.ID), strptr(viewer.ID),
			viewer.FirstName+" "+viewer.LastName+" started following you")
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": status})
}

// Unfollow removes an existing follow relationship.
func (h *Handler) Unfollow(w http.ResponseWriter, r *http.Request) {
	viewer, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	targetID := r.PathValue("id")
	if err := models.Unfollow(h.DB, viewer.ID, targetID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to unfollow")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "unfollowed"})
}

// RespondFollowRequest accepts or declines a pending follow request, where the
// authenticated user is the one being followed.
func (h *Handler) RespondFollowRequest(w http.ResponseWriter, r *http.Request) {
	viewer, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	requesterID := r.PathValue("id")
	var body struct {
		Action string `json:"action"` // "accept" | "decline"
	}
	if !decodeJSON(w, r, &body) {
		return
	}

	f, err := models.GetFollow(h.DB, requesterID, viewer.ID)
	if err != nil || f.Status != "pending" {
		utils.Error(w, http.StatusNotFound, "no pending request from this user")
		return
	}

	switch body.Action {
	case "accept":
		if err := models.SetFollowStatus(h.DB, requesterID, viewer.ID, "accepted"); err != nil {
			utils.Error(w, http.StatusInternalServerError, "failed to accept")
			return
		}
		h.notify(requesterID, "follow_accepted", strptr(viewer.ID), strptr(viewer.ID),
			viewer.FirstName+" "+viewer.LastName+" accepted your follow request")
	case "decline":
		if err := models.Unfollow(h.DB, requesterID, viewer.ID); err != nil {
			utils.Error(w, http.StatusInternalServerError, "failed to decline")
			return
		}
	default:
		utils.Error(w, http.StatusBadRequest, "action must be accept or decline")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": body.Action})
}

// ListFollowRequests returns the pending follow requests for the current user.
func (h *Handler) ListFollowRequests(w http.ResponseWriter, r *http.Request) {
	viewer, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	users, err := models.ListPendingFollowRequests(h.DB, viewer.ID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	utils.JSON(w, http.StatusOK, users)
}
