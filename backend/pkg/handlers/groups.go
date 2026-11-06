package handlers

import (
	"net/http"
	"strings"

	"socialnetwork/pkg/models"
	"socialnetwork/pkg/utils"
)

// CreateGroup creates a group with the current user as creator/first member.
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	body.Description = strings.TrimSpace(body.Description)
	if body.Title == "" || body.Description == "" {
		utils.Error(w, http.StatusBadRequest, "title and description are required")
		return
	}
	groupID := newID()
	if err := models.CreateGroup(h.DB, groupID, user.ID, body.Title, body.Description, newID()); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to create group")
		return
	}
	group, _ := models.GetGroup(h.DB, groupID, user.ID)
	utils.JSON(w, http.StatusCreated, group)
}

// ListGroups returns all groups for browsing.
func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	groups, err := models.ListGroups(h.DB, user.ID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	utils.JSON(w, http.StatusOK, groups)
}

// groupDetail is the payload for a single group page.
type groupDetail struct {
	Group           *models.Group   `json:"group"`
	Members         []*models.User  `json:"members"`
	Posts           []*models.Post  `json:"posts"`
	Events          []*models.Event `json:"events"`
	PendingRequests []*models.User  `json:"pendingRequests"`
	NonMembers      []*models.User  `json:"nonMembers"`
}

// GetGroup returns a group with its members, posts and events. Posts/events are
// only included for members.
func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	groupID := r.PathValue("id")
	group, err := models.GetGroup(h.DB, groupID, user.ID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "group not found")
		return
	}

	detail := groupDetail{
		Group:           group,
		Members:         []*models.User{},
		Posts:           []*models.Post{},
		Events:          []*models.Event{},
		PendingRequests: []*models.User{},
		NonMembers:      []*models.User{},
	}
	if members, err := models.ListGroupMembers(h.DB, groupID); err == nil {
		detail.Members = members
	}

	if group.Membership == "member" {
		if posts, err := models.ListGroupPosts(h.DB, groupID); err == nil {
			detail.Posts = posts
		}
		if events, err := models.ListGroupEvents(h.DB, groupID, user.ID); err == nil {
			detail.Events = events
		}
		// Users not yet in the group, so members can invite them. Users with a
		// pending join request are excluded too: inviting them would overwrite
		// the request the creator still has to accept or decline.
		if all, err := models.ListUsers(h.DB, user.ID); err == nil {
			for _, u := range all {
				if status, err := models.GetMembership(h.DB, groupID, u.ID); err != nil ||
					(status != "member" && status != "invited" && status != "requested") {
					detail.NonMembers = append(detail.NonMembers, u)
				}
			}
		}
	}
	// The creator can see pending join requests.
	if group.CreatorID == user.ID {
		if reqs, err := models.ListGroupRequesters(h.DB, groupID); err == nil {
			detail.PendingRequests = reqs
		}
	}
	utils.JSON(w, http.StatusOK, detail)
}

// InviteToGroup lets a member invite another user into the group.
func (h *Handler) InviteToGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	groupID := r.PathValue("id")
	var body struct {
		UserID string `json:"userId"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	member, _ := models.IsGroupMember(h.DB, groupID, user.ID)
	if !member {
		utils.Error(w, http.StatusForbidden, "only members can invite")
		return
	}
	if existing, err := models.GetMembership(h.DB, groupID, body.UserID); err == nil {
		switch existing {
		case "member", "invited":
			utils.Error(w, http.StatusConflict, "user is already a member or invited")
			return
		case "requested":
			// Do not overwrite a pending join request with an invitation; the
			// creator still has to accept or decline the request.
			utils.Error(w, http.StatusConflict, "user already requested to join; the group creator can accept the request")
			return
		}
	}
	if err := models.UpsertMember(h.DB, newID(), groupID, body.UserID, "invited", strptr(user.ID)); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to invite")
		return
	}
	group, _ := models.GetGroup(h.DB, groupID, user.ID)
	h.notify(body.UserID, "group_invite", strptr(user.ID), strptr(groupID),
		user.FirstName+" "+user.LastName+" invited you to join "+group.Title)
	utils.JSON(w, http.StatusOK, map[string]string{"status": "invited"})
}

// RequestJoinGroup lets a non-member ask to join; the creator must approve.
func (h *Handler) RequestJoinGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	groupID := r.PathValue("id")
	group, err := models.GetGroup(h.DB, groupID, user.ID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "group not found")
		return
	}
	switch group.Membership {
	case "member":
		utils.Error(w, http.StatusConflict, "you are already a member")
		return
	case "invited":
		// Keep the invitation instead of overwriting it with a join request;
		// the user should accept or decline the invitation.
		utils.Error(w, http.StatusConflict, "you already have an invitation to this group; accept or decline it")
		return
	case "requested":
		// Idempotent: the request already exists, do not re-notify the creator.
		utils.JSON(w, http.StatusOK, map[string]string{"status": "requested"})
		return
	}
	if err := models.UpsertMember(h.DB, newID(), groupID, user.ID, "requested", nil); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to request")
		return
	}
	h.notify(group.CreatorID, "group_join_request", strptr(user.ID), strptr(groupID),
		user.FirstName+" "+user.LastName+" requested to join "+group.Title)
	utils.JSON(w, http.StatusOK, map[string]string{"status": "requested"})
}

// LeaveGroup lets a member leave a group. The creator cannot leave their own
// group: they are the only one who can manage join requests.
func (h *Handler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	groupID := r.PathValue("id")
	group, err := models.GetGroup(h.DB, groupID, user.ID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "group not found")
		return
	}
	if group.CreatorID == user.ID {
		utils.Error(w, http.StatusBadRequest, "the group creator cannot leave their own group")
		return
	}
	if group.Membership != "member" {
		utils.Error(w, http.StatusBadRequest, "you are not a member of this group")
		return
	}
	if err := models.RemoveMember(h.DB, groupID, user.ID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to leave group")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "left"})
}

// RespondInvitation lets an invited user accept or decline their invitation.
func (h *Handler) RespondInvitation(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	groupID := r.PathValue("id")
	var body struct {
		Action string `json:"action"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	status, err := models.GetMembership(h.DB, groupID, user.ID)
	if err != nil || status != "invited" {
		utils.Error(w, http.StatusNotFound, "no pending invitation")
		return
	}
	if body.Action == "accept" {
		_ = models.SetMemberStatus(h.DB, groupID, user.ID, "member")
	} else {
		_ = models.RemoveMember(h.DB, groupID, user.ID)
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": body.Action})
}

// RespondJoinRequest lets the group creator accept or decline a join request.
func (h *Handler) RespondJoinRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	groupID := r.PathValue("id")
	requesterID := r.PathValue("userId")
	group, err := models.GetGroup(h.DB, groupID, user.ID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "group not found")
		return
	}
	if group.CreatorID != user.ID {
		utils.Error(w, http.StatusForbidden, "only the creator can respond to join requests")
		return
	}
	var body struct {
		Action string `json:"action"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	status, err := models.GetMembership(h.DB, groupID, requesterID)
	if err != nil || status != "requested" {
		utils.Error(w, http.StatusNotFound, "no pending request from this user")
		return
	}
	if body.Action == "accept" {
		_ = models.SetMemberStatus(h.DB, groupID, requesterID, "member")
		h.notify(requesterID, "group_request_accepted", strptr(user.ID), strptr(groupID),
			"Your request to join "+group.Title+" was accepted")
	} else {
		_ = models.RemoveMember(h.DB, groupID, requesterID)
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": body.Action})
}
