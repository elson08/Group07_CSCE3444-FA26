package internal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/config"
	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database"
)

// Server owns the HTTP router, listener, and database dependency.
type Server struct {
	database   *database.Database
	mux        *http.ServeMux
	httpServer *http.Server
}

// NewServer builds the API server from its database and configuration.
func NewServer(db *database.Database, cfg config.Config) *Server {
	server := &Server{database: db, mux: http.NewServeMux()}
	server.httpServer = &http.Server{
		Addr:              cfg.ListenAddress(),
		Handler:           server,
		ReadHeaderTimeout: 5 * time.Second,
	}

	server.registerRoutes()
	return server
}

// ServeHTTP checks the request origin and dispatches it to a route.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !validOrigin(r) {
		writeError(w, http.StatusForbidden, "cross-origin request denied")
		return
	}

	s.mux.ServeHTTP(w, r)
}

// Listen serves requests until an interrupt or termination signal arrives.
func (s *Server) Listen() error {
	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	finished := make(chan error, 1)

	log.Printf("API listening on http://%s", s.httpServer.Addr)
	go func() {
		finished <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-finished:
		return listenerError(err)
	case <-signalContext.Done():
		stop()
		return s.shutdown(finished)
	}
}

// shutdown waits for active requests and the listener to finish.
func (s *Server) shutdown(finished <-chan error) error {
	log.Print("shutting down API")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return errors.Join(s.Close(ctx), listenerError(<-finished))
}

// listenerError treats a normal server close as success.
func listenerError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

// Close gracefully stops the HTTP server within the given context.
func (s *Server) Close(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.httpServer.Close()
		return err
	}

	return nil
}

// health reports whether the database is available.
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.database.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// registerRoutes connects API handlers and serves the frontend.
func (s *Server) registerRoutes() {
	s.handle("GET /api/health", s.health)

	s.handle("POST /api/users", s.register)
	s.handle("POST /api/sessions", s.login)
	s.handle("DELETE /api/sessions/current", s.logout)
	s.handle("GET /api/me", s.me)

	s.handle("GET /api/titles", s.listTitles)
	s.handle("GET /api/titles/{id}", s.getTitle)
	s.handle("GET /api/services", s.listServices)

	s.handle("GET /api/me/titles", s.listSavedTitles)
	s.handle("PUT /api/me/titles/{id}", s.saveTitle)
	s.handle("DELETE /api/me/titles/{id}", s.removeTitle)

	s.handle("GET /api/me/services", s.listSavedServices)
	s.handle("PUT /api/me/services/{id}", s.saveService)
	s.handle("DELETE /api/me/services/{id}", s.removeService)

	s.handle("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "endpoint not found")
	})
	s.mux.Handle("/", http.FileServer(http.Dir("../frontend")))
}

// handle registers a route on the server router.
func (s *Server) handle(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

// readJSON decodes one JSON object with a size limit.
func readJSON(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("request must contain one JSON object")
	}
	return nil
}

// writeJSON sends a JSON response.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// pathID parses the positive ID in a route.
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}

	return id, true
}

// validOrigin checks the origin of requests that can change data.
func validOrigin(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	parsed, err := url.Parse(origin)
	return err == nil && strings.EqualFold(parsed.Host, r.Host)
}
