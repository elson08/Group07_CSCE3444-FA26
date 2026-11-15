package internal

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database"
)

func validateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return errors.New("username must be 3-32 letters, digits, hyphens, or underscores")
	}
	return nil
}

func validateEmail(email string) error {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 254 {
		return errors.New("a valid email address is required")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return errors.New("password must be 8-72 bytes")
	}
	return nil
}

func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if err := errors.Join(validateUsername(input.Username), validateEmail(input.Email)); err != nil {
		writeError(w, http.StatusBadRequest, strings.ReplaceAll(err.Error(), "\n", "; "))
		return
	}

	updated, err := s.database.UpdateUser(r.Context(), user.ID, input.Username, input.Email)
	if errors.Is(err, database.ErrUserExists) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update account")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)) != nil {
		writeError(w, http.StatusForbidden, "current password is incorrect")
		return
	}

	if err := validatePassword(input.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err == nil {
		err = s.database.UpdatePassword(r.Context(), user.ID, string(hash))
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not change password")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteMe(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}

	var input struct {
		Password string `json:"password"`
	}
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		writeError(w, http.StatusForbidden, "password is incorrect")
		return
	}

	if err := s.database.DeleteUser(r.Context(), user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete account")
		return
	}

	setSessionCookie(w, r, "", time.Time{})
	w.WriteHeader(http.StatusNoContent)
}
