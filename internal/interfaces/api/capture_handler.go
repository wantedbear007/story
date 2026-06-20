package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/application/raw_entry"
	"github.com/anomalyco/story/internal/domain"
	"github.com/anomalyco/story/web"
)

type CaptureServer struct {
	host        string
	port        int
	rawEntrySvc *raw_entry.Service
	handler     http.Handler
	httpServer  *http.Server
}

func NewCaptureServer(host string, port int, rawEntrySvc *raw_entry.Service) *CaptureServer {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/capture", handleCapture(rawEntrySvc))
	mux.HandleFunc("GET /api/capture/whoami", handleWhoami)

	sub, err := fs.Sub(web.Assets, ".")
	if err == nil {
		mux.Handle("GET /{path...}", http.FileServer(http.FS(sub)))
	}

	return &CaptureServer{
		host:        host,
		port:        port,
		rawEntrySvc: rawEntrySvc,
		handler:     mux,
	}
}

func (s *CaptureServer) Start(ctx context.Context) error {
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: s.handler,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(shutdownCtx)
	}()

	fmt.Printf("Capture server: http://%s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *CaptureServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

type captureRequest struct {
	Content string `json:"content"`
	UserID  string `json:"user_id"`
}

func handleCapture(rawEntrySvc *raw_entry.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req captureRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Content == "" {
			writeError(w, http.StatusBadRequest, "content is required")
			return
		}
		if req.UserID == "" {
			writeError(w, http.StatusBadRequest, "user_id is required")
			return
		}
		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user_id")
			return
		}

		resp, err := rawEntrySvc.Create(r.Context(), userID, raw_entry.CreateRawEntryRequest{
			Content: req.Content,
			Source:  domain.RawEntrySourceNotificationCapture,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, resp)
	}
}

type sessionJSON struct {
	UserID string `json:"user_id"`
}

func handleWhoami(w http.ResponseWriter, r *http.Request) {
	home, err := os.UserHomeDir()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"user_id": ""})
		return
	}
	data, err := os.ReadFile(filepath.Join(home, ".story", "session.json"))
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"user_id": ""})
		return
	}
	var s sessionJSON
	if err := json.Unmarshal(data, &s); err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"user_id": ""})
		return
	}
	if s.UserID == "" {
		writeJSON(w, http.StatusOK, map[string]string{"user_id": ""})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"user_id": s.UserID})
}
