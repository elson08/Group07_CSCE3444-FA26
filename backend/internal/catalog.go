package internal

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database/model"
)

// listTitles searches the public title catalog.
func (s *Server) listTitles(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	kind := r.URL.Query().Get("kind")

	if kind != "" && kind != model.KindMovie && kind != model.KindShow {
		writeError(w, http.StatusBadRequest, "kind must be movie or show")
		return
	}

	limit, offset := 20, 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 50 {
			writeError(w, http.StatusBadRequest, "limit must be 1-50")
			return
		}
		limit = value
	}

	if raw := r.URL.Query().Get("offset"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			writeError(w, http.StatusBadRequest, "offset must be nonnegative")
			return
		}
		offset = value
	}

	titles, err := s.database.ListTitles(r.Context(), query, kind, limit, offset)
	if err == nil && s.fillDetails(r.Context(), titles) {
		titles, err = s.database.ListTitles(r.Context(), query, kind, limit, offset)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list titles")
		return
	}

	writeJSON(w, http.StatusOK, titles)
}

// getTitle returns a title with its details.
func (s *Server) getTitle(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	title, err := s.database.TitleByID(r.Context(), id)
	if err == nil && s.fillDetails(r.Context(), []model.Title{title}) {
		title, err = s.database.TitleByID(r.Context(), id)
	}

	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "title not found")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load title")
		return
	}

	writeJSON(w, http.StatusOK, title)
}

// listServices returns the public service catalog.
func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	services, err := s.database.Services(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list services")
		return
	}

	writeJSON(w, http.StatusOK, services)
}

// listSavedTitles returns the signed-in user’s saved titles.
func (s *Server) listSavedTitles(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	titles, err := s.database.SavedTitles(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list saved titles")
		return
	}

	writeJSON(w, http.StatusOK, titles)
}

func (s *Server) listRecommendations(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	const limit = 8
	titles, err := s.database.RecommendedTitles(r.Context(), user.ID, limit)
	if err == nil && s.fillDetails(r.Context(), titles) {
		titles, err = s.database.RecommendedTitles(r.Context(), user.ID, limit)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load recommendations")
		return
	}

	writeJSON(w, http.StatusOK, titles)
}

// saveTitle adds a title to the signed-in user’s list.
func (s *Server) saveTitle(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	id, ok := pathID(w, r)
	if !ok {
		return
	}

	exists, err := s.database.TitleExists(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load title")
		return
	}

	if !exists {
		writeError(w, http.StatusNotFound, "title not found")
		return
	}

	if err := s.database.SaveTitle(r.Context(), user.ID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save title")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// removeTitle removes a title from the signed-in user’s list.
func (s *Server) removeTitle(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	id, ok := pathID(w, r)
	if !ok {
		return
	}

	if err := s.database.RemoveTitle(r.Context(), user.ID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not remove title")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// listSavedServices returns the signed-in user’s services.
func (s *Server) listSavedServices(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	services, err := s.database.Services(r.Context(), &user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list saved services")
		return
	}

	writeJSON(w, http.StatusOK, services)
}

// saveService adds a service to the signed-in user’s list.
func (s *Server) saveService(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	id, ok := pathID(w, r)
	if !ok {
		return
	}

	exists, err := s.database.ServiceExists(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load service")
		return
	}

	if !exists {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}

	if err := s.database.SaveService(r.Context(), user.ID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save service")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// removeService removes a service from the signed-in user’s list.
func (s *Server) removeService(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	id, ok := pathID(w, r)
	if !ok {
		return
	}

	if err := s.database.RemoveService(r.Context(), user.ID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not remove service")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
