package handlers

import (
	"net/http"
	"strings"

	"socialnetwork/pkg/models"
	"socialnetwork/pkg/utils"
)

// CreatePost creates a post (optionally in a group, optionally with an image).
// Privacy can be public, almost_private or private; for private posts the
// allowed viewer ids are taken from the repeated "allowedUsers" form field.
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid form data")
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		utils.Error(w, http.StatusBadRequest, "content is required")
		return
	}

	var groupID *string
	if gid := strings.TrimSpace(r.FormValue("groupId")); gid != "" {
		member, _ := models.IsGroupMember(h.DB, gid, user.ID)
		if !member {
			utils.Error(w, http.StatusForbidden, "you are not a member of this group")
			return
		}
		groupID = &gid
	}

	privacy := r.FormValue("privacy")
	if groupID != nil {
		privacy = "group"
	} else if privacy != "public" && privacy != "almost_private" && privacy != "private" {
		privacy = "public"
	}

	// Optional image.
	var imagePath *string
	if file, header, ferr := utils.FormFileOptional(r, "image"); ferr == nil && file != nil {
		defer file.Close()
		path, serr := utils.SaveImage(file, header, h.UploadDir)
		if serr != nil {
			utils.Error(w, http.StatusBadRequest, serr.Error())
			return
		}
		imagePath = &path
	}

	postID := newID()
	if err := models.CreatePost(h.DB, postID, user.ID, groupID, content, imagePath, privacy); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	// Allowed viewers for a private post.
	if privacy == "private" {
		for _, uid := range r.Form["allowedUsers"] {
			if uid = strings.TrimSpace(uid); uid != "" {
				_ = models.AddAllowedViewer(h.DB, postID, uid)
			}
		}
	}

	post, err := models.GetPost(h.DB, postID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to load post")
		return
	}
	utils.JSON(w, http.StatusCreated, post)
}

// GetFeed returns the home feed (non-group posts the user can see).
func (h *Handler) GetFeed(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	posts, err := models.ListFeed(h.DB, user.ID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	utils.JSON(w, http.StatusOK, posts)
}

// GetPost returns a single post if the user is allowed to see it.
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	post, err := models.GetPost(h.DB, r.PathValue("id"))
	if err != nil {
		utils.Error(w, http.StatusNotFound, "post not found")
		return
	}
	if canView, _ := models.CanViewPost(h.DB, user.ID, post); !canView {
		utils.Error(w, http.StatusForbidden, "you cannot view this post")
		return
	}
	utils.JSON(w, http.StatusOK, post)
}

// CreateComment adds a comment (optionally with an image) to a post.
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	postID := r.PathValue("id")
	post, err := models.GetPost(h.DB, postID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "post not found")
		return
	}
	if canView, _ := models.CanViewPost(h.DB, user.ID, post); !canView {
		utils.Error(w, http.StatusForbidden, "you cannot comment on this post")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid form data")
		return
	}
	content := strings.TrimSpace(r.FormValue("content"))
	var imagePath *string
	if file, header, ferr := utils.FormFileOptional(r, "image"); ferr == nil && file != nil {
		defer file.Close()
		path, serr := utils.SaveImage(file, header, h.UploadDir)
		if serr != nil {
			utils.Error(w, http.StatusBadRequest, serr.Error())
			return
		}
		imagePath = &path
	}
	if content == "" && imagePath == nil {
		utils.Error(w, http.StatusBadRequest, "comment cannot be empty")
		return
	}

	if err := models.CreateComment(h.DB, newID(), postID, user.ID, content, imagePath); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to create comment")
		return
	}
	comments, err := models.ListComments(h.DB, postID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	utils.JSON(w, http.StatusCreated, comments)
}

// ListComments returns the comments on a post (viewer must be able to see it).
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	postID := r.PathValue("id")
	post, err := models.GetPost(h.DB, postID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "post not found")
		return
	}
	if canView, _ := models.CanViewPost(h.DB, user.ID, post); !canView {
		utils.Error(w, http.StatusForbidden, "you cannot view this post")
		return
	}
	comments, err := models.ListComments(h.DB, postID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	utils.JSON(w, http.StatusOK, comments)
}
