package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/application/content"
	"github.com/anomalyco/story/internal/application/entry"
	"github.com/anomalyco/story/internal/infrastructure/auth"
	"github.com/anomalyco/story/web"
)

type loginCodeEntry struct {
	token     string
	expiresAt time.Time
}

type LoginCodeStore struct {
	mu    sync.Mutex
	codes map[string]loginCodeEntry
}

func NewLoginCodeStore() *LoginCodeStore {
	return &LoginCodeStore{
		codes: make(map[string]loginCodeEntry),
	}
}

func (s *LoginCodeStore) Create(token string) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	code := string(b)

	s.mu.Lock()
	s.codes[code] = loginCodeEntry{
		token:     token,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	s.mu.Unlock()
	return code, nil
}

func (s *LoginCodeStore) Exchange(code string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.codes[code]
	if !ok {
		return "", false
	}
	delete(s.codes, code)
	if time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.token, true
}

type Server struct {
	host         string
	port         int
	tweetService *content.Service
	entryService *entry.Service
	jwtService   *auth.JWTTokenService
	loginCodes   *LoginCodeStore
	httpServer   *http.Server
	authURL      string
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
		loginCodes:   NewLoginCodeStore(),
	}
}

func (s *Server) SetPort(port int) {
	s.port = port
}

func (s *Server) CreateLoginCode(token string) (string, error) {
	return s.loginCodes.Create(token)
}

func (s *Server) ValidateToken(token string) error {
	_, _, err := s.jwtService.ValidateAccessToken(token)
	return err
}

func (s *Server) SetAuthURL(url string) {
	s.authURL = url
}

func (s *Server) handleExchangeCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing code")
		return
	}

	token, ok := s.loginCodes.Exchange(code)
	if !ok {
		writeError(w, http.StatusNotFound, "invalid or expired code")
		return
	}

	if err := s.ValidateToken(token); err != nil {
		writeError(w, http.StatusUnauthorized, "session expired, please re-authenticate")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
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

	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	})

	mux.HandleFunc("GET /api/exchange/{code}", s.handleExchangeCode)

	sub, err := fs.Sub(web.Assets, ".")
	if err != nil {
		return fmt.Errorf("web assets: %w", err)
	}
	mux.Handle("GET /{path...}", noCacheMiddleware(http.FileServer(http.FS(sub))))

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: corsMiddleware(mux),
	}

	addr := s.httpServer.Addr
	fmt.Printf("Story dashboard: http://%s\n", addr)

	browserURL := s.authURL
	if browserURL == "" {
		browserURL = fmt.Sprintf("http://%s", addr)
	}
	go openBrowser(browserURL)

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

func noCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
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
