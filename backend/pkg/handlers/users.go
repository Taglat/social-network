package handlers

import (
	"net/http"

	"socialnetwork/pkg/models"
	"socialnetwork/pkg/utils"
)

// profileResponse is the payload returned for a profile view.
type profileResponse struct {
	User        *models.User   `json:"user"`
	Posts       []*models.Post `json:"posts"`
	Followers   []*models.User `json:"followers"`
	Following   []*models.User `json:"following"`
	IsSelf      bool           `json:"isSelf"`
	IsFollowing bool           `json:"isFollowing"`     // viewer -> profile (accepted)
	RequestSent bool           `json:"requestSent"`     // viewer -> profile (pending)
	CanViewFull bool           `json:"canViewFull"`     // privacy allows full view
	Online      bool           `json:"online"`
}

// GetProfile returns a user's profile, respecting public/private visibility.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	viewer, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	target, err := models.GetUserByID(h.DB, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "user not found")
		return
	}

	isSelf := viewer.ID == target.ID
	isFollowing, _ := models.IsAcceptedFollower(h.DB, viewer.ID, target.ID)
	requestSent := false
	if f, err := models.GetFollow(h.DB, viewer.ID, target.ID); err == nil && f.Status == "pending" {
		requestSent = true
	}

	// Full view = own profile, public profile, or accepted follower.
	canViewFull := isSelf || target.IsPublic || isFollowing

	resp := profileResponse{
		User:        target,
		IsSelf:      isSelf,
		IsFollowing: isFollowing,
		RequestSent: requestSent,
		CanViewFull: canViewFull,
		Online:      h.Hub.IsOnline(target.ID),
		Posts:       []*models.Post{},
		Followers:   []*models.User{},
		Following:   []*models.User{},
	}

	// Followers/following and posts are always derived but only sent on full view.
	if canViewFull {
		if posts, err := models.ListUserPosts(h.DB, target.ID, viewer.ID); err == nil {
			resp.Posts = posts
		}
		if followers, err := models.ListFollowers(h.DB, target.ID); err == nil {
			resp.Followers = followers
		}
		if following, err := models.ListFollowing(h.DB, target.ID); err == nil {
			resp.Following = following
		}
	}
	utils.JSON(w, http.StatusOK, resp)
}

// ListUsers returns all other users (used by the chat user picker).
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	viewer, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	users, err := models.ListUsers(h.DB, viewer.ID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	// Annotate online status for convenience.
	type userStatus struct {
		*models.User
		Online bool `json:"online"`
	}
	out := make([]userStatus, 0, len(users))
	for _, u := range users {
		out = append(out, userStatus{User: u, Online: h.Hub.IsOnline(u.ID)})
	}
	utils.JSON(w, http.StatusOK, out)
}

// SetPrivacy toggles the authenticated user's profile between public/private.
func (h *Handler) SetPrivacy(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	var body struct {
		IsPublic bool `json:"isPublic"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := models.SetPrivacy(h.DB, user.ID, body.IsPublic); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to update privacy")
		return
	}
	user.IsPublic = body.IsPublic
	utils.JSON(w, http.StatusOK, user)
}
