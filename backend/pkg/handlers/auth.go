package handlers

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"socialnetwork/pkg/models"
	"socialnetwork/pkg/session"
	"socialnetwork/pkg/utils"
)

// Register creates a new user account. It accepts multipart/form-data so an
// optional avatar image can be uploaded alongside the text fields.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// Fall back to plain form parsing when no file part is present.
		if err := r.ParseForm(); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid form data")
			return
		}
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")
	firstName := strings.TrimSpace(r.FormValue("firstName"))
	lastName := strings.TrimSpace(r.FormValue("lastName"))
	dob := strings.TrimSpace(r.FormValue("dateOfBirth"))

	if email == "" || password == "" || firstName == "" || lastName == "" || dob == "" {
		utils.Error(w, http.StatusBadRequest, "email, password, first name, last name and date of birth are required")
		return
	}
	if len(password) < 6 {
		utils.Error(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	exists, err := models.EmailExists(h.DB, email)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "database error")
		return
	}
	if exists {
		utils.Error(w, http.StatusConflict, "an account with this email already exists")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := &models.User{
		ID:           newID(),
		Email:        email,
		PasswordHash: string(hash),
		FirstName:    firstName,
		LastName:     lastName,
		DateOfBirth:  dob,
		Nickname:     optional(r.FormValue("nickname")),
		AboutMe:      optional(r.FormValue("aboutMe")),
		IsPublic:     r.FormValue("isPublic") == "true",
	}

	// Optional avatar upload.
	if file, header, ferr := utils.FormFileOptional(r, "avatar"); ferr == nil && file != nil {
		defer file.Close()
		path, serr := utils.SaveImage(file, header, h.UploadDir)
		if serr != nil {
			utils.Error(w, http.StatusBadRequest, serr.Error())
			return
		}
		user.Avatar = &path
	}

	if err := models.CreateUser(h.DB, user); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	if err := session.Create(w, h.DB, user.ID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to start session")
		return
	}
	utils.JSON(w, http.StatusCreated, user)
}

// Login authenticates a user with email + password and starts a session.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	email := strings.TrimSpace(strings.ToLower(body.Email))

	user, err := models.GetUserByEmail(h.DB, email)
	if err != nil {
		// Same response for unknown email and wrong password (no user enumeration).
		utils.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)) != nil {
		utils.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := session.Create(w, h.DB, user.ID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to start session")
		return
	}
	utils.JSON(w, http.StatusOK, user)
}

// Logout ends the current session.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session.Destroy(w, r, h.DB)
	utils.JSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// Me returns the currently authenticated user.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}
	utils.JSON(w, http.StatusOK, user)
}

// optional returns a pointer to a trimmed string, or nil when empty.
func optional(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
