package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/application/content"
	"github.com/anomalyco/story/internal/domain"
)

func (s *Server) handleListTweets(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entryIDStr := r.URL.Query().Get("entry_id")
	var parsedEntryID *uuid.UUID
	if entryIDStr != "" {
		id, err := uuid.Parse(entryIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid entry_id")
			return
		}
		parsedEntryID = &id
	}

	statusStr := r.URL.Query().Get("status")
	var parsedStatus *domain.TweetStatus
	if statusStr != "" {
		s := domain.TweetStatus(statusStr)
		parsedStatus = &s
	}

	limit := parseIntParam(r.URL.Query().Get("limit"), 20)
	offset := parseIntParam(r.URL.Query().Get("offset"), 0)

	resp, err := s.tweetService.List(r.Context(), content.ListRequest{
		UserID:  userID,
		EntryID: parsedEntryID,
		Status:  parsedStatus,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tweets": resp.Tweets,
		"total":  resp.Total,
	})
}

func (s *Server) handleGetTweet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tweetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tweet id")
		return
	}

	resp, err := s.tweetService.Get(r.Context(), userID, tweetID)
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGenerateTweet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		EntryID     string  `json:"entry_id"`
		PromptName  string  `json:"prompt_name"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"max_tokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	entryID, err := uuid.Parse(req.EntryID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entry_id")
		return
	}

	resp, err := s.tweetService.Generate(r.Context(), userID, content.GenerateRequest{
		EntryID:     entryID,
		PromptName:  req.PromptName,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	})
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleRegenerateTweet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tweetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tweet id")
		return
	}

	var req struct {
		PromptName  string  `json:"prompt_name"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"max_tokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	resp, err := s.tweetService.Regenerate(r.Context(), userID, content.RegenerateRequest{
		TweetID:     tweetID,
		PromptName:  req.PromptName,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	})
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleUpdateTweet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tweetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tweet id")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	resp, err := s.tweetService.UpdateContent(r.Context(), userID, tweetID, req.Content)
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleApproveTweet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tweetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tweet id")
		return
	}

	resp, err := s.tweetService.Approve(r.Context(), userID, content.ApproveRequest{TweetID: tweetID})
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleReviewTweet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tweetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tweet id")
		return
	}

	resp, err := s.tweetService.Review(r.Context(), userID, tweetID)
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRejectTweet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tweetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tweet id")
		return
	}

	resp, err := s.tweetService.Reject(r.Context(), userID, tweetID)
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleScheduleTweet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tweetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tweet id")
		return
	}

	var req struct {
		ScheduledAt string `json:"scheduled_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	scheduledAt, err := timeParse(req.ScheduledAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid scheduled_at format, use ISO 8601")
		return
	}

	resp, err := s.tweetService.Schedule(r.Context(), userID, content.ScheduleRequest{
		TweetID:     tweetID,
		ScheduledAt: scheduledAt,
	})
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleArchiveTweet(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tweetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tweet id")
		return
	}

	resp, err := s.tweetService.Archive(r.Context(), userID, tweetID)
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetAudits(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tweetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tweet id")
		return
	}

	audits, err := s.tweetService.GetAudits(r.Context(), userID, tweetID)
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"audits": audits,
	})
}

func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return defaultVal
	}
	return n
}

func timeParse(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse %q as datetime", s)
}

func httpStatusCode(err error) int {
	if err == domain.ErrNotFound {
		return http.StatusNotFound
	}
	if err == domain.ErrForbidden || err == domain.ErrUnauthorized {
		return http.StatusForbidden
	}
	if err == domain.ErrInvalidInput || err == domain.ErrValidationFailed {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
