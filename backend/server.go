// Command server is the entry point of the social network backend. It wires
// together the three major parts of the project: the Server (this HTTP server),
// the App (handlers + middleware + websocket hub) and the Database (sqlite).
package main

import (
	"log"
	"net/http"
	"os"

	"socialnetwork/pkg/db/sqlite"
	"socialnetwork/pkg/handlers"
	"socialnetwork/pkg/middleware"
	"socialnetwork/pkg/ws"
)

func main() {
	cfg := loadConfig()

	// --- Database: open connection and apply migrations ---
	db, err := sqlite.Open(cfg.dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := sqlite.RunMigrations(db, cfg.migrationsPath); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	// --- App: websocket hub + HTTP handlers ---
	hub := ws.NewHub(db)
	go hub.Run()

	h := handlers.New(db, hub, cfg.uploadDir)
	mux := http.NewServeMux()
	registerRoutes(mux, h, cfg.uploadDir)

	// --- Server: middleware chain + listen ---
	// Only the first-party frontend may make credentialed cross-origin calls.
	handler := middleware.Recover(middleware.Logger(
		middleware.CORS(mux, cfg.frontendURL, "http://localhost:3000")))

	log.Printf("backend listening on :%s (frontend: %s)", cfg.port, cfg.frontendURL)
	if err := http.ListenAndServe(":"+cfg.port, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// registerRoutes maps every HTTP route. Route patterns use Go 1.22 method-aware
// ServeMux matching (e.g. "POST /api/posts").
func registerRoutes(mux *http.ServeMux, h *handlers.Handler, uploadDir string) {
	// Health check.
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth.
	mux.HandleFunc("POST /api/register", h.Register)
	mux.HandleFunc("POST /api/login", h.Login)
	mux.HandleFunc("POST /api/logout", h.Logout)
	mux.HandleFunc("GET /api/me", h.Me)

	// Users / profile.
	mux.HandleFunc("GET /api/users", h.ListUsers)
	mux.HandleFunc("GET /api/users/{id}", h.GetProfile)
	mux.HandleFunc("PUT /api/profile/privacy", h.SetPrivacy)

	// Followers.
	mux.HandleFunc("POST /api/users/{id}/follow", h.Follow)
	mux.HandleFunc("POST /api/users/{id}/unfollow", h.Unfollow)
	mux.HandleFunc("GET /api/follow-requests", h.ListFollowRequests)
	mux.HandleFunc("POST /api/follow-requests/{id}/respond", h.RespondFollowRequest)

	// Posts & comments.
	mux.HandleFunc("POST /api/posts", h.CreatePost)
	mux.HandleFunc("GET /api/posts", h.GetFeed)
	mux.HandleFunc("GET /api/posts/{id}", h.GetPost)
	mux.HandleFunc("POST /api/posts/{id}/comments", h.CreateComment)
	mux.HandleFunc("GET /api/posts/{id}/comments", h.ListComments)

	// Groups.
	mux.HandleFunc("POST /api/groups", h.CreateGroup)
	mux.HandleFunc("GET /api/groups", h.ListGroups)
	mux.HandleFunc("GET /api/groups/{id}", h.GetGroup)
	mux.HandleFunc("POST /api/groups/{id}/invite", h.InviteToGroup)
	mux.HandleFunc("POST /api/groups/{id}/request", h.RequestJoinGroup)
	mux.HandleFunc("POST /api/groups/{id}/leave", h.LeaveGroup)
	mux.HandleFunc("POST /api/groups/{id}/invitation/respond", h.RespondInvitation)
	mux.HandleFunc("POST /api/groups/{id}/requests/{userId}/respond", h.RespondJoinRequest)
	mux.HandleFunc("GET /api/groups/{id}/messages", h.GetGroupMessages)

	// Events.
	mux.HandleFunc("POST /api/groups/{id}/events", h.CreateEvent)
	mux.HandleFunc("POST /api/events/{id}/respond", h.RespondEvent)

	// Chat.
	mux.HandleFunc("GET /api/chat/partners", h.ChatPartners)
	mux.HandleFunc("GET /api/messages/{id}", h.GetPrivateMessages)
	mux.HandleFunc("GET /ws", h.ServeWS)

	// Notifications.
	mux.HandleFunc("GET /api/notifications", h.ListNotifications)
	mux.HandleFunc("POST /api/notifications/read-all", h.MarkAllNotificationsRead)
	mux.HandleFunc("POST /api/notifications/{id}/read", h.MarkNotificationRead)

	// Uploaded images.
	fileServer := http.FileServer(http.Dir(uploadDir))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", fileServer))
}

type config struct {
	port           string
	dbPath         string
	uploadDir      string
	migrationsPath string
	frontendURL    string
}

func loadConfig() config {
	return config{
		port:           env("PORT", "8080"),
		dbPath:         env("DB_PATH", "data/social.db"),
		uploadDir:      env("UPLOAD_DIR", "uploads"),
		migrationsPath: env("MIGRATIONS_PATH", "pkg/db/migrations/sqlite"),
		frontendURL:    env("FRONTEND_URL", "http://localhost:3000"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
