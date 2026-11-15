package internal

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/config"
	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database"
	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database/model"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

type testTitle struct {
	id     int64
	kind   string
	name   string
	year   int
	genres []string
}

var testTitles = []testTitle{
	{1, model.KindMovie, "Action Hero", 2020, []string{"Action"}},
	{2, model.KindMovie, "Action Sequel", 2022, []string{"Action"}},
	{3, model.KindMovie, "Laugh Riot", 2021, []string{"Comedy"}},
	{4, model.KindShow, "Space Station", 2019, []string{"Action", "Sci-Fi"}},
	{5, model.KindMovie, "Quiet Drama", 2023, []string{"Drama"}},
	{6, model.KindMovie, "Explosion Comedy", 2018, []string{"Action", "Comedy"}},
}

type testApp struct {
	t      *testing.T
	server *httptest.Server
	raw    *sql.DB
	client *http.Client
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.sqlite")

	db, err := database.New(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { raw.Close() })

	seedTestCatalog(t, raw)

	server := httptest.NewServer(NewServer(db, config.DefaultConfig()))
	t.Cleanup(server.Close)

	app := &testApp{t: t, server: server, raw: raw}
	app.client = app.newClient()
	return app
}

func seedTestCatalog(t *testing.T, raw *sql.DB) {
	t.Helper()
	genreIDs := map[string]int64{}

	for _, title := range testTitles {
		_, err := raw.Exec(`INSERT INTO titles (id, provider_id, kind, name, release_year) VALUES (?, ?, ?, ?, ?)`,
			title.id, fmt.Sprintf("tt%07d", title.id), title.kind, title.name, title.year)
		if err != nil {
			t.Fatal(err)
		}

		for _, genre := range title.genres {
			id, known := genreIDs[genre]
			if !known {
				result, err := raw.Exec(`INSERT INTO genres (name) VALUES (?)`, genre)
				if err != nil {
					t.Fatal(err)
				}
				id, _ = result.LastInsertId()
				genreIDs[genre] = id
			}

			if _, err := raw.Exec(`INSERT INTO title_genres (title_id, genre_id) VALUES (?, ?)`, title.id, id); err != nil {
				t.Fatal(err)
			}
		}
	}

	for _, statement := range []string{
		`INSERT OR IGNORE INTO services (name) VALUES ('Netflix'), ('Hulu')`,
		`INSERT INTO title_services (title_id, service_id) SELECT 6, id FROM services WHERE name = 'Netflix'`,
	} {
		if _, err := raw.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
}

func (a *testApp) newClient() *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		a.t.Fatal(err)
	}

	return &http.Client{Jar: jar}
}

func (a *testApp) request(client *http.Client, method, path string, body any, headers ...string) (int, []byte) {
	a.t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			a.t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, a.server.URL+path, reader)
	if err != nil {
		a.t.Fatal(err)
	}

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		request.Header.Set(headers[i], headers[i+1])
	}

	response, err := client.Do(request)
	if err != nil {
		a.t.Fatal(err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		a.t.Fatal(err)
	}

	return response.StatusCode, data
}

func (a *testApp) do(method, path string, body any) (int, []byte) {
	a.t.Helper()
	return a.request(a.client, method, path, body)
}

func (a *testApp) expect(method, path string, body any, want int) []byte {
	a.t.Helper()

	status, data := a.do(method, path, body)
	if status != want {
		a.t.Fatalf("%s %s: status %d, want %d (body %s)", method, path, status, want, data)
	}

	return data
}

func (a *testApp) signUp(username string) {
	a.t.Helper()

	a.expect(http.MethodPost, "/api/users", map[string]string{
		"username": username,
		"email":    username + "@example.com",
		"password": "correct-horse",
	}, http.StatusCreated)

	a.expect(http.MethodPost, "/api/sessions", map[string]string{
		"identity": username,
		"password": "correct-horse",
	}, http.StatusOK)
}

func decode[T any](t *testing.T, data []byte) T {
	t.Helper()

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode %s: %v", data, err)
	}

	return value
}

func titleIDs(t *testing.T, data []byte) []int64 {
	t.Helper()

	var ids []int64
	for _, title := range decode[[]model.Title](t, data) {
		ids = append(ids, title.ID)
	}

	return ids
}

func sameIDs(got, want []int64) bool {
	if len(got) != len(want) {
		return false
	}

	seen := map[int64]int{}
	for _, id := range got {
		seen[id]++
	}
	for _, id := range want {
		seen[id]--
	}
	for _, count := range seen {
		if count != 0 {
			return false
		}
	}

	return true
}

func contains(ids []int64, id int64) bool {
	for _, value := range ids {
		if value == id {
			return true
		}
	}

	return false
}

func TestHealth(t *testing.T) {
	app := newTestApp(t)

	data := app.expect(http.MethodGet, "/api/health", nil, http.StatusOK)
	if decode[map[string]string](t, data)["status"] != "ok" {
		t.Fatalf("unexpected health response %s", data)
	}
}

func TestUnknownEndpointReturnsNotFound(t *testing.T) {
	app := newTestApp(t)
	app.expect(http.MethodGet, "/api/does-not-exist", nil, http.StatusNotFound)
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	app := newTestApp(t)

	cases := map[string]map[string]string{
		"short username":    {"username": "ab", "email": "ab@example.com", "password": "correct-horse"},
		"bad characters":    {"username": "bad name!", "email": "bad@example.com", "password": "correct-horse"},
		"invalid email":     {"username": "valid_user", "email": "not-an-email", "password": "correct-horse"},
		"short password":    {"username": "valid_user", "email": "valid@example.com", "password": "short"},
		"too long password": {"username": "valid_user", "email": "valid@example.com", "password": strings.Repeat("x", 73)},
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			app.expect(http.MethodPost, "/api/users", body, http.StatusBadRequest)
		})
	}

	app.expect(http.MethodPost, "/api/users", map[string]string{"username": "valid_user", "extra": "field"}, http.StatusBadRequest)
}

func TestRegisterRejectsDuplicateAccounts(t *testing.T) {
	app := newTestApp(t)

	app.expect(http.MethodPost, "/api/users", map[string]string{
		"username": "moviefan", "email": "fan@example.com", "password": "correct-horse",
	}, http.StatusCreated)

	app.expect(http.MethodPost, "/api/users", map[string]string{
		"username": "MovieFan", "email": "other@example.com", "password": "correct-horse",
	}, http.StatusConflict)

	app.expect(http.MethodPost, "/api/users", map[string]string{
		"username": "another", "email": "FAN@example.com", "password": "correct-horse",
	}, http.StatusConflict)
}

func TestRegisterNeverReturnsPassword(t *testing.T) {
	app := newTestApp(t)

	data := app.expect(http.MethodPost, "/api/users", map[string]string{
		"username": "moviefan", "email": "fan@example.com", "password": "correct-horse",
	}, http.StatusCreated)

	if bytes.Contains(data, []byte("correct-horse")) || bytes.Contains(bytes.ToLower(data), []byte("password")) {
		t.Fatalf("registration response leaks the password: %s", data)
	}

	var stored string
	if err := app.raw.QueryRow(`SELECT password_hash FROM users WHERE username = 'moviefan'`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == "correct-horse" || !strings.HasPrefix(stored, "$2") {
		t.Fatalf("password is not stored as a bcrypt hash: %q", stored)
	}
}

func TestLoginAndLogout(t *testing.T) {
	app := newTestApp(t)

	app.expect(http.MethodPost, "/api/users", map[string]string{
		"username": "moviefan", "email": "fan@example.com", "password": "correct-horse",
	}, http.StatusCreated)

	app.expect(http.MethodPost, "/api/sessions", map[string]string{"identity": "moviefan", "password": "wrong-password"}, http.StatusUnauthorized)
	app.expect(http.MethodPost, "/api/sessions", map[string]string{"identity": "nobody", "password": "correct-horse"}, http.StatusUnauthorized)
	app.expect(http.MethodPost, "/api/sessions", map[string]string{"identity": "", "password": ""}, http.StatusBadRequest)
	app.expect(http.MethodGet, "/api/me", nil, http.StatusUnauthorized)

	app.expect(http.MethodPost, "/api/sessions", map[string]string{"identity": "FAN@example.com", "password": "correct-horse"}, http.StatusOK)

	me := decode[model.User](t, app.expect(http.MethodGet, "/api/me", nil, http.StatusOK))
	if me.Username != "moviefan" || me.Email != "fan@example.com" {
		t.Fatalf("unexpected user %+v", me)
	}

	app.expect(http.MethodDelete, "/api/sessions/current", nil, http.StatusNoContent)
	app.expect(http.MethodGet, "/api/me", nil, http.StatusUnauthorized)
}

func TestSessionCookieIsHttpOnly(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	serverURL, _ := url.Parse(app.server.URL)
	cookies := app.client.Jar.Cookies(serverURL)
	if len(cookies) != 1 || cookies[0].Name != "session" {
		t.Fatalf("expected one session cookie, got %v", cookies)
	}

	response, err := http.Post(app.server.URL+"/api/sessions", "application/json",
		strings.NewReader(`{"identity":"moviefan","password":"correct-horse"}`))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	header := response.Header.Get("Set-Cookie")
	if !strings.Contains(header, "HttpOnly") || !strings.Contains(header, "SameSite=Strict") {
		t.Fatalf("session cookie is missing HttpOnly or SameSite: %s", header)
	}
}

func TestExpiredSessionIsRejected(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	_, err := app.raw.Exec(`INSERT INTO sessions (token_hash, user_id, expires_at)
		SELECT ?, id, ? FROM users WHERE username = 'moviefan'`,
		sessionHash("expired-token"), time.Now().Add(-time.Hour).Unix())
	if err != nil {
		t.Fatal(err)
	}

	client := app.newClient()
	serverURL, _ := url.Parse(app.server.URL)
	client.Jar.SetCookies(serverURL, []*http.Cookie{{Name: "session", Value: "expired-token"}})

	if status, _ := app.request(client, http.MethodGet, "/api/me", nil); status != http.StatusUnauthorized {
		t.Fatalf("expired session: status %d, want 401", status)
	}
}

func TestSignedInRoutesRequireLogin(t *testing.T) {
	app := newTestApp(t)

	routes := []struct{ method, path string }{
		{http.MethodGet, "/api/me"},
		{http.MethodGet, "/api/me/titles"},
		{http.MethodPut, "/api/me/titles/1"},
		{http.MethodDelete, "/api/me/titles/1"},
		{http.MethodGet, "/api/me/recommendations"},
		{http.MethodGet, "/api/me/services"},
		{http.MethodPut, "/api/me/services/1"},
	}

	for _, route := range routes {
		app.expect(route.method, route.path, nil, http.StatusUnauthorized)
	}
}

func TestCrossOriginWritesAreRejected(t *testing.T) {
	app := newTestApp(t)

	status, _ := app.request(app.client, http.MethodPost, "/api/users", map[string]string{
		"username": "moviefan", "email": "fan@example.com", "password": "correct-horse",
	}, "Origin", "https://evil.example")
	if status != http.StatusForbidden {
		t.Fatalf("cross-origin POST: status %d, want 403", status)
	}
}

func TestSearchTitles(t *testing.T) {
	app := newTestApp(t)

	all := titleIDs(t, app.expect(http.MethodGet, "/api/titles", nil, http.StatusOK))
	if !sameIDs(all, []int64{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("listing all titles returned %v", all)
	}

	byName := titleIDs(t, app.expect(http.MethodGet, "/api/titles?q=action", nil, http.StatusOK))
	if !sameIDs(byName, []int64{1, 2}) {
		t.Fatalf("searching for \"action\" returned %v", byName)
	}

	shows := titleIDs(t, app.expect(http.MethodGet, "/api/titles?kind=show", nil, http.StatusOK))
	if !sameIDs(shows, []int64{4}) {
		t.Fatalf("filtering by show returned %v", shows)
	}

	none := titleIDs(t, app.expect(http.MethodGet, "/api/titles?q=zzzz", nil, http.StatusOK))
	if len(none) != 0 {
		t.Fatalf("search with no matches returned %v", none)
	}

	page := titleIDs(t, app.expect(http.MethodGet, "/api/titles?limit=2", nil, http.StatusOK))
	next := titleIDs(t, app.expect(http.MethodGet, "/api/titles?limit=2&offset=2", nil, http.StatusOK))
	if len(page) != 2 || len(next) != 2 || contains(page, next[0]) || contains(page, next[1]) {
		t.Fatalf("pagination returned overlapping pages %v and %v", page, next)
	}
}

func TestSearchTitlesRejectsBadParameters(t *testing.T) {
	app := newTestApp(t)

	for _, query := range []string{"kind=podcast", "limit=0", "limit=51", "limit=abc", "offset=-1"} {
		app.expect(http.MethodGet, "/api/titles?"+query, nil, http.StatusBadRequest)
	}
}

func TestGetTitle(t *testing.T) {
	app := newTestApp(t)

	title := decode[model.Title](t, app.expect(http.MethodGet, "/api/titles/4", nil, http.StatusOK))
	if title.Name != "Space Station" || title.Kind != model.KindShow {
		t.Fatalf("unexpected title %+v", title)
	}

	var genres []string
	for _, genre := range title.Genres {
		genres = append(genres, genre.Name)
	}
	if strings.Join(genres, ",") != "Action,Sci-Fi" {
		t.Fatalf("title genres are %v, want Action and Sci-Fi", genres)
	}

	withService := decode[model.Title](t, app.expect(http.MethodGet, "/api/titles/6", nil, http.StatusOK))
	if len(withService.Services) != 1 || withService.Services[0].Name != "Netflix" {
		t.Fatalf("title services are %+v, want Netflix", withService.Services)
	}

	app.expect(http.MethodGet, "/api/titles/999", nil, http.StatusNotFound)
	app.expect(http.MethodGet, "/api/titles/abc", nil, http.StatusBadRequest)
	app.expect(http.MethodGet, "/api/titles/-1", nil, http.StatusBadRequest)
}

func TestSaveAndRemoveTitles(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	app.expect(http.MethodPut, "/api/me/titles/1", nil, http.StatusNoContent)
	app.expect(http.MethodPut, "/api/me/titles/1", nil, http.StatusNoContent)
	app.expect(http.MethodPut, "/api/me/titles/3", nil, http.StatusNoContent)
	app.expect(http.MethodPut, "/api/me/titles/999", nil, http.StatusNotFound)

	saved := titleIDs(t, app.expect(http.MethodGet, "/api/me/titles", nil, http.StatusOK))
	if !sameIDs(saved, []int64{1, 3}) {
		t.Fatalf("saved titles are %v, want 1 and 3", saved)
	}

	app.expect(http.MethodDelete, "/api/me/titles/1", nil, http.StatusNoContent)

	saved = titleIDs(t, app.expect(http.MethodGet, "/api/me/titles", nil, http.StatusOK))
	if !sameIDs(saved, []int64{3}) {
		t.Fatalf("after removing, saved titles are %v, want 3", saved)
	}
}

func TestSavedTitlesArePerUser(t *testing.T) {
	app := newTestApp(t)
	app.signUp("first")
	app.expect(http.MethodPut, "/api/me/titles/1", nil, http.StatusNoContent)

	app.client = app.newClient()
	app.signUp("second")

	saved := titleIDs(t, app.expect(http.MethodGet, "/api/me/titles", nil, http.StatusOK))
	if len(saved) != 0 {
		t.Fatalf("second user sees another user's saved titles: %v", saved)
	}
}

func TestSaveAndRemoveServices(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	services := decode[[]model.Service](t, app.expect(http.MethodGet, "/api/services", nil, http.StatusOK))
	if len(services) == 0 {
		t.Fatal("service catalog is empty")
	}

	path := "/api/me/services/" + itoa(services[0].ID)
	app.expect(http.MethodPut, path, nil, http.StatusNoContent)
	app.expect(http.MethodPut, "/api/me/services/999", nil, http.StatusNotFound)

	mine := decode[[]model.Service](t, app.expect(http.MethodGet, "/api/me/services", nil, http.StatusOK))
	if len(mine) != 1 || mine[0].ID != services[0].ID {
		t.Fatalf("saved services are %+v, want %s", mine, services[0].Name)
	}

	app.expect(http.MethodDelete, path, nil, http.StatusNoContent)

	mine = decode[[]model.Service](t, app.expect(http.MethodGet, "/api/me/services", nil, http.StatusOK))
	if len(mine) != 0 {
		t.Fatalf("after removing, saved services are %+v", mine)
	}
}

func TestRecommendationsFollowSavedGenres(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	empty := titleIDs(t, app.expect(http.MethodGet, "/api/me/recommendations", nil, http.StatusOK))
	if len(empty) != 0 {
		t.Fatalf("user with nothing saved got recommendations %v", empty)
	}

	app.expect(http.MethodPut, "/api/me/titles/1", nil, http.StatusNoContent)

	recommended := titleIDs(t, app.expect(http.MethodGet, "/api/me/recommendations", nil, http.StatusOK))
	if !sameIDs(recommended, []int64{2, 4, 6}) {
		t.Fatalf("after saving an action movie, recommendations are %v, want the other action titles 2, 4 and 6", recommended)
	}
	if contains(recommended, 1) {
		t.Fatal("recommendations include a title the user already saved")
	}
}

func TestRecommendationsRankStrongestGenreMatchFirst(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	app.expect(http.MethodPut, "/api/me/titles/1", nil, http.StatusNoContent)
	app.expect(http.MethodPut, "/api/me/titles/3", nil, http.StatusNoContent)

	recommended := titleIDs(t, app.expect(http.MethodGet, "/api/me/recommendations", nil, http.StatusOK))
	if len(recommended) == 0 || recommended[0] != 6 {
		t.Fatalf("recommendations are %v, want 6 (action and comedy) first", recommended)
	}
}

func TestRecommendationsPreferUsersServices(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")
	app.expect(http.MethodPut, "/api/me/titles/1", nil, http.StatusNoContent)

	before := titleIDs(t, app.expect(http.MethodGet, "/api/me/recommendations", nil, http.StatusOK))
	if len(before) == 0 || before[0] == 6 {
		t.Fatalf("without services, recommendations are %v; expected newer titles before 6", before)
	}

	var netflix int64
	if err := app.raw.QueryRow(`SELECT id FROM services WHERE name = 'Netflix'`).Scan(&netflix); err != nil {
		t.Fatal(err)
	}
	app.expect(http.MethodPut, "/api/me/services/"+itoa(netflix), nil, http.StatusNoContent)

	after := titleIDs(t, app.expect(http.MethodGet, "/api/me/recommendations", nil, http.StatusOK))
	if len(after) == 0 || after[0] != 6 {
		t.Fatalf("with Netflix saved, recommendations are %v, want 6 (on Netflix) first", after)
	}
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
