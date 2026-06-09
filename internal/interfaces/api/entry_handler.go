package api

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/application/entry"
)

func (s *Server) handleListEntries(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, err := s.entryService.List(r.Context(), userID, entry.EntryFilterRequest{
		Query:    r.URL.Query().Get("q"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"entries": resp.Entries,
		"total":   resp.Total,
		"page":    resp.Page,
	})
}

func (s *Server) handleGetEntry(w http.ResponseWriter, r *http.Request) {
	entryID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	resp, err := s.entryService.Get(r.Context(), entryID)
	if err != nil {
		writeError(w, httpStatusCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListPrompts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"prompts": []map[string]interface{}{
			{"name": "tweet-summarize", "description": "Summarize an entry as a single tweet"},
			{"name": "tweet-thread", "description": "Convert an entry into a multi-tweet thread"},
			{"name": "blog-summarize", "description": "Summarize an entry as a short blog post"},
		},
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := userUUIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID.String(),
	})
}
