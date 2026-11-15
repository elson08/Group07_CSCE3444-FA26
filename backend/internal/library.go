package internal

import (
	"net/http"
	"strings"
	"unicode/utf8"
)

const maxReviewLength = 2000

func (s *Server) requireTitle(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, ok := pathID(w, r)
	if !ok {
		return 0, false
	}

	exists, err := s.database.TitleExists(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load title")
		return 0, false
	}

	if !exists {
		writeError(w, http.StatusNotFound, "title not found")
		return 0, false
	}

	return id, true
}

func (s *Server) listGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := s.database.Genres(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list genres")
		return
	}

	writeJSON(w, http.StatusOK, genres)
}

func (s *Server) listSimilar(w http.ResponseWriter, r *http.Request) {
	id, ok := s.requireTitle(w, r)
	if !ok {
		return
	}

	const limit = 12
	titles, err := s.database.SimilarTitles(r.Context(), id, limit)
	if err == nil && s.fillDetails(r.Context(), titles) {
		titles, err = s.database.SimilarTitles(r.Context(), id, limit)
	}

	if err == nil {
		err = s.database.AttachServices(r.Context(), titles)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load similar titles")
		return
	}

	writeJSON(w, http.StatusOK, titles)
}

func (s *Server) listReviews(w http.ResponseWriter, r *http.Request) {
	id, ok := s.requireTitle(w, r)
	if !ok {
		return
	}

	reviews, err := s.database.TitleReviews(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load reviews")
		return
	}

	writeJSON(w, http.StatusOK, reviews)
}

func (s *Server) titleStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	id, ok := s.requireTitle(w, r)
	if !ok {
		return
	}

	status, err := s.database.TitleStatus(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load title status")
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func (s *Server) listRatings(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	ratings, err := s.database.UserRatings(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list ratings")
		return
	}

	writeJSON(w, http.StatusOK, ratings)
}

func (s *Server) saveRating(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	id, ok := s.requireTitle(w, r)
	if !ok {
		return
	}

	var input struct {
		Stars  int    `json:"stars"`
		Review string `json:"review"`
	}
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	input.Review = strings.TrimSpace(input.Review)
	if input.Stars < 1 || input.Stars > 5 {
		writeError(w, http.StatusBadRequest, "stars must be 1-5")
		return
	}

	if utf8.RuneCountInString(input.Review) > maxReviewLength {
		writeError(w, http.StatusBadRequest, "review must be 2000 characters or fewer")
		return
	}

	if err := s.database.SaveRating(r.Context(), user.ID, id, input.Stars, input.Review); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save rating")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) removeRating(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	id, ok := pathID(w, r)
	if !ok {
		return
	}

	if err := s.database.RemoveRating(r.Context(), user.ID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not remove rating")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listWatched(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	titles, err := s.database.WatchedTitles(r.Context(), user.ID)
	if err == nil {
		err = s.database.AttachServices(r.Context(), titles)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list watched titles")
		return
	}

	writeJSON(w, http.StatusOK, titles)
}

func (s *Server) markWatched(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	id, ok := s.requireTitle(w, r)
	if !ok {
		return
	}

	if err := s.database.MarkWatched(r.Context(), user.ID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not mark title as watched")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) unmarkWatched(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	id, ok := pathID(w, r)
	if !ok {
		return
	}

	if err := s.database.UnmarkWatched(r.Context(), user.ID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not update watched titles")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
