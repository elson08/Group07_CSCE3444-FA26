package internal

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database"
	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database/model"
)

// listTitles searches the public title catalog.
func (s *Server) listTitles(w http.ResponseWriter, r *http.Request) {
	filter, err := s.titleFilter(r)
	if errors.Is(err, errSignInRequired) {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	titles, total, err := s.database.BrowseTitles(r.Context(), filter)
	if err == nil && s.fillDetails(r.Context(), titles) {
		titles, total, err = s.database.BrowseTitles(r.Context(), filter)
	}

	if err == nil {
		err = s.database.AttachServices(r.Context(), titles)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list titles")
		return
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	writeJSON(w, http.StatusOK, titles)
}

var errSignInRequired = errors.New("sign in to filter by your services")

func (s *Server) titleFilter(r *http.Request) (model.TitleFilter, error) {
	values := r.URL.Query()
	filter := model.TitleFilter{
		Query: strings.TrimSpace(values.Get("q")),
		Kind:  values.Get("kind"),
		Sort:  values.Get("sort"),
		Limit: 20,
	}

	if filter.Kind != "" && filter.Kind != model.KindMovie && filter.Kind != model.KindShow {
		return filter, errors.New("kind must be movie or show")
	}

	if filter.Sort != "" && !database.ValidSort(filter.Sort) {
		return filter, errors.New("sort must be newest, oldest, top, or az")
	}

	for _, genre := range values["genre"] {
		if genre = strings.TrimSpace(genre); genre != "" {
			filter.Genres = append(filter.Genres, genre)
		}
	}

	for _, raw := range values["service"] {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			return filter, errors.New("service must be a service id")
		}
		filter.ServiceIDs = append(filter.ServiceIDs, id)
	}

	if len(filter.Genres) > 30 || len(filter.ServiceIDs) > 30 {
		return filter, errors.New("too many filters")
	}

	if values.Get("mine") == "1" {
		user, err := s.currentUser(r)
		if err != nil {
			return filter, errSignInRequired
		}
		filter.ServicesOfUser = user.ID
	}

	numbers := []struct {
		name     string
		min, max int
		target   *int
	}{
		{"limit", 1, 50, &filter.Limit},
		{"offset", 0, 100000, &filter.Offset},
		{"year_from", 1800, 2100, &filter.YearFrom},
		{"year_to", 1800, 2100, &filter.YearTo},
	}

	for _, number := range numbers {
		raw := values.Get(number.name)
		if raw == "" {
			continue
		}

		value, err := strconv.Atoi(raw)
		if err != nil || value < number.min || value > number.max {
			return filter, fmt.Errorf("%s must be %d-%d", number.name, number.min, number.max)
		}
		*number.target = value
	}

	if raw := values.Get("min_rating"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil || value < 0 || value > 5 {
			return filter, errors.New("min_rating must be 0-5")
		}
		filter.MinRating = value
	}

	return filter, nil
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
	if err == nil {
		err = s.database.AttachServices(r.Context(), titles)
	}

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

	if err == nil {
		err = s.database.AttachServices(r.Context(), titles)
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
