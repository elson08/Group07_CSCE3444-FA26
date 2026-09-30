package internal

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
	
	"golang.org/x/crypto/bcrypt"
	
	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database"
	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database/model"
)

const sessionDuration = 7 * 24 * time.Hour

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,32}$`)

type registrationRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// register creates an account after validating its details.
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var input registrationRequest
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if err := input.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create account")
		return
	}
	
	user, err := s.database.CreateUser(r.Context(), input.Username, input.Email, string(hash))
	if err != nil {
		if errors.Is(err, database.ErrUserExists) {
			writeError(w, http.StatusConflict, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "could not create account")
		}
		return
	}
	
	writeJSON(w, http.StatusCreated, user)
}

// login verifies credentials and creates a session cookie.
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}
	if err := readJSON(w, r, &input); err != nil || input.Identity == "" || input.Password == "" {
		writeError(w, http.StatusBadRequest, "identity and password are required")
		return
	}
	
	user, err := s.database.UserByIdentity(r.Context(), strings.TrimSpace(input.Identity))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "database unavailable")
		return
	}
	
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	
	token := rand.Text()
	expires := time.Now().Add(sessionDuration)
	if err := s.database.CreateSession(r.Context(), sessionHash(token), user.ID, expires); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	
	setSessionCookie(w, r, token, expires)
	writeJSON(w, http.StatusOK, user)
}

// logout deletes the current session and clears its cookie.
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		if err := s.database.DeleteSession(r.Context(), sessionHash(cookie.Value)); err != nil {
			writeError(w, http.StatusInternalServerError, "could not end session")
			return
		}
	}
	
	setSessionCookie(w, r, "", time.Time{})
	w.WriteHeader(http.StatusNoContent)
}

// me returns the signed-in account.
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	
	writeJSON(w, http.StatusOK, user)
}

// currentUser finds the account for the session cookie.
func (s *Server) currentUser(r *http.Request) (model.User, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return model.User{}, err
	}
	
	return s.database.UserBySession(r.Context(), sessionHash(cookie.Value))
}

// requireUser returns the signed-in user or writes an error.
func (s *Server) requireUser(w http.ResponseWriter, r *http.Request) (model.User, bool) {
	user, err := s.currentUser(r)
	if err == nil {
		return user, true
	}
	
	if errors.Is(err, http.ErrNoCookie) || errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusUnauthorized, "authentication required")
	} else {
		writeError(w, http.StatusInternalServerError, "database unavailable")
	}
	
	return model.User{}, false
}

// sessionHash hashes a token before database storage.
func sessionHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// setSessionCookie writes or clears the session cookie.
func setSessionCookie(w http.ResponseWriter, r *http.Request, token string, expires time.Time) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	}
	if token == "" {
		cookie.MaxAge = -1
	}
	
	http.SetCookie(w, cookie)
}

// validate checks the account fields before registration.
func (input registrationRequest) validate() error {
	if !usernamePattern.MatchString(input.Username) {
		return errors.New("username must be 3-32 letters, digits, hyphens, or underscores")
	}
	
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email || len(input.Email) > 254 {
		return errors.New("a valid email address is required")
	}
	
	if len(input.Password) < 8 || len(input.Password) > 72 {
		return errors.New("password must be 8-72 bytes")
	}
	
	return nil
}
