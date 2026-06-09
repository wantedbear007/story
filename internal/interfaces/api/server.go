package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/application/content"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/infrastructure/auth"
)

type Server struct {
	host         string
	port         int
	tweetService *content.Service
	entryService *entry.Service
	jwtService   *auth.JWTTokenService
	httpServer   *http.Server
}

func NewServer(
	host string,
	port int,
	tweetService *content.Service,
	entryService *entry.Service,
	jwtService *auth.JWTTokenService,
) *Server {
	return &Server{
		host:         host,
		port:         port,
		tweetService: tweetService,
		entryService: entryService,
		jwtService:   jwtService,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/tweets", s.authMiddleware(s.handleListTweets))
	mux.HandleFunc("GET /api/tweets/{id}", s.authMiddleware(s.handleGetTweet))
	mux.HandleFunc("POST /api/tweets/generate", s.authMiddleware(s.handleGenerateTweet))
	mux.HandleFunc("POST /api/tweets/{id}/regenerate", s.authMiddleware(s.handleRegenerateTweet))
	mux.HandleFunc("PUT /api/tweets/{id}", s.authMiddleware(s.handleUpdateTweet))
	mux.HandleFunc("POST /api/tweets/{id}/approve", s.authMiddleware(s.handleApproveTweet))
	mux.HandleFunc("POST /api/tweets/{id}/review", s.authMiddleware(s.handleReviewTweet))
	mux.HandleFunc("POST /api/tweets/{id}/reject", s.authMiddleware(s.handleRejectTweet))
	mux.HandleFunc("POST /api/tweets/{id}/schedule", s.authMiddleware(s.handleScheduleTweet))
	mux.HandleFunc("POST /api/tweets/{id}/archive", s.authMiddleware(s.handleArchiveTweet))
	mux.HandleFunc("GET /api/tweets/{id}/audits", s.authMiddleware(s.handleGetAudits))

	mux.HandleFunc("GET /api/entries", s.authMiddleware(s.handleListEntries))
	mux.HandleFunc("GET /api/entries/{id}", s.authMiddleware(s.handleGetEntry))

	mux.HandleFunc("GET /api/prompts", s.authMiddleware(s.handleListPrompts))

	mux.HandleFunc("GET /api/me", s.authMiddleware(s.handleMe))

	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/", fs)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: corsMiddleware(mux),
	}

	addr := s.httpServer.Addr
	fmt.Printf("Story dashboard: http://%s\n", addr)

	go openBrowser(fmt.Sprintf("http://%s", addr))

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(shutdownCtx)
	}()

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

type ctxKey string

const (
	ctxUserID  ctxKey = "user_id"
	ctxSession ctxKey = "session_id"
)

func userUUIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ctxUserID)
	if v == nil {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func sessionUUIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ctxSession)
	if v == nil {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func (s *Server) authMiddleware(next func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractBearerToken(r)
		if tokenStr == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization token")
			return
		}

		userID, sessionID, err := s.jwtService.ValidateAccessToken(tokenStr)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, userID)
		ctx = context.WithValue(ctx, ctxSession, sessionID)
		next(w, r.WithContext(ctx))
	}
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	if err != nil {
		fmt.Printf("Could not open browser: %v\n", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
