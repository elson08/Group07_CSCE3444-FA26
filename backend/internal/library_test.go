package internal

import (
	"net/http"
	"strings"
	"testing"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database/model"
)

func (a *testApp) netflixID() int64 {
	a.t.Helper()

	var id int64
	if err := a.raw.QueryRow(`SELECT id FROM services WHERE name = 'Netflix'`).Scan(&id); err != nil {
		a.t.Fatal(err)
	}

	return id
}

func (a *testApp) browse(query string) []int64 {
	a.t.Helper()
	return titleIDs(a.t, a.expect(http.MethodGet, "/api/titles?"+query, nil, http.StatusOK))
}

func (a *testApp) rate(titleID int64, stars int, review string) {
	a.t.Helper()
	a.expect(http.MethodPut, "/api/me/ratings/"+itoa(titleID), map[string]any{"stars": stars, "review": review}, http.StatusNoContent)
}

func (a *testApp) title(id int64) model.Title {
	a.t.Helper()
	return decode[model.Title](a.t, a.expect(http.MethodGet, "/api/titles/"+itoa(id), nil, http.StatusOK))
}

func (a *testApp) status(id int64) model.TitleStatus {
	a.t.Helper()
	return decode[model.TitleStatus](a.t, a.expect(http.MethodGet, "/api/me/titles/"+itoa(id), nil, http.StatusOK))
}

func sameOrder(got, want []int64) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

func TestBrowseFiltersByGenre(t *testing.T) {
	app := newTestApp(t)

	if got := app.browse("genre=Comedy"); !sameIDs(got, []int64{3, 6}) {
		t.Fatalf("Comedy titles are %v, want 3 and 6", got)
	}

	if got := app.browse("genre=Comedy&genre=Drama"); !sameIDs(got, []int64{3, 5, 6}) {
		t.Fatalf("Comedy or Drama titles are %v, want 3, 5 and 6", got)
	}

	if got := app.browse("genre=Action&kind=movie"); !sameIDs(got, []int64{1, 2, 6}) {
		t.Fatalf("Action movies are %v, want 1, 2 and 6", got)
	}
}

func TestBrowseFiltersByServiceAndYear(t *testing.T) {
	app := newTestApp(t)

	if got := app.browse("service=" + itoa(app.netflixID())); !sameIDs(got, []int64{6}) {
		t.Fatalf("Netflix titles are %v, want 6", got)
	}

	if got := app.browse("year_from=2021"); !sameIDs(got, []int64{2, 3, 5}) {
		t.Fatalf("titles from 2021 on are %v, want 2, 3 and 5", got)
	}

	if got := app.browse("year_to=2019"); !sameIDs(got, []int64{4, 6}) {
		t.Fatalf("titles up to 2019 are %v, want 4 and 6", got)
	}

	if got := app.browse("year_from=2019&year_to=2020"); !sameIDs(got, []int64{1, 4}) {
		t.Fatalf("titles from 2019 to 2020 are %v, want 1 and 4", got)
	}
}

func TestBrowseOnlyMyServices(t *testing.T) {
	app := newTestApp(t)
	app.expect(http.MethodGet, "/api/titles?mine=1", nil, http.StatusUnauthorized)

	app.signUp("moviefan")
	if got := app.browse("mine=1"); len(got) != 0 {
		t.Fatalf("user without services got %v", got)
	}

	app.expect(http.MethodPut, "/api/me/services/"+itoa(app.netflixID()), nil, http.StatusNoContent)
	if got := app.browse("mine=1"); !sameIDs(got, []int64{6}) {
		t.Fatalf("titles on my services are %v, want 6", got)
	}
}

func TestBrowseSortOrders(t *testing.T) {
	app := newTestApp(t)

	cases := map[string][]int64{
		"":            {5, 2, 3, 1, 4, 6},
		"sort=newest": {5, 2, 3, 1, 4, 6},
		"sort=oldest": {6, 4, 1, 3, 2, 5},
		"sort=az":     {1, 2, 6, 3, 5, 4},
	}

	for query, want := range cases {
		if got := app.browse(query); !sameOrder(got, want) {
			t.Fatalf("%q returned %v, want %v", query, got, want)
		}
	}
}

func TestBrowseReportsTotalCount(t *testing.T) {
	app := newTestApp(t)

	response, err := app.client.Get(app.server.URL + "/api/titles?limit=2&genre=Action")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if total := response.Header.Get("X-Total-Count"); total != "4" {
		t.Fatalf("X-Total-Count is %q, want 4", total)
	}
}

func TestBrowseRejectsBadFilters(t *testing.T) {
	app := newTestApp(t)

	queries := []string{
		"sort=random",
		"year_from=1700",
		"year_to=3000",
		"min_rating=6",
		"min_rating=abc",
		"service=abc",
		"service=0",
		"genre=" + strings.Repeat("x&genre=", 30) + "x",
	}

	for _, query := range queries {
		app.expect(http.MethodGet, "/api/titles?"+query, nil, http.StatusBadRequest)
	}
}

func TestGenresAndSimilarTitles(t *testing.T) {
	app := newTestApp(t)

	var names []string
	for _, genre := range decode[[]model.Genre](t, app.expect(http.MethodGet, "/api/genres", nil, http.StatusOK)) {
		names = append(names, genre.Name)
	}
	if strings.Join(names, ",") != "Action,Comedy,Drama,Sci-Fi" {
		t.Fatalf("genres are %v", names)
	}

	similar := titleIDs(t, app.expect(http.MethodGet, "/api/titles/1/similar", nil, http.StatusOK))
	if !sameIDs(similar, []int64{2, 6}) {
		t.Fatalf("titles similar to an action movie are %v, want the other action movies 2 and 6", similar)
	}

	app.expect(http.MethodGet, "/api/titles/999/similar", nil, http.StatusNotFound)
}

func TestRatingsUpdateCommunityScores(t *testing.T) {
	app := newTestApp(t)

	app.signUp("first")
	app.rate(1, 5, "Great action")

	app.client = app.newClient()
	app.signUp("second")
	app.rate(1, 3, "")

	title := app.title(1)
	if title.RatingCount != 2 || title.RatingAverage == nil || *title.RatingAverage != 4 {
		t.Fatalf("title 1 has %d ratings averaging %v, want 2 averaging 4", title.RatingCount, title.RatingAverage)
	}

	app.rate(1, 4, "")
	title = app.title(1)
	if title.RatingCount != 2 || *title.RatingAverage != 4.5 {
		t.Fatalf("after changing a rating, title 1 has %d ratings averaging %v, want 2 averaging 4.5", title.RatingCount, *title.RatingAverage)
	}

	if got := app.browse("min_rating=4.5"); !sameIDs(got, []int64{1}) {
		t.Fatalf("titles rated 4.5 or higher are %v, want 1", got)
	}
	if got := app.browse("min_rating=4.6"); len(got) != 0 {
		t.Fatalf("titles rated 4.6 or higher are %v, want none", got)
	}
	if got := app.browse("sort=top"); len(got) == 0 || got[0] != 1 {
		t.Fatalf("top rated order is %v, want 1 first", got)
	}

	app.expect(http.MethodDelete, "/api/me/ratings/1", nil, http.StatusNoContent)
	title = app.title(1)
	if title.RatingCount != 1 || *title.RatingAverage != 5 {
		t.Fatalf("after removing a rating, title 1 has %d ratings averaging %v, want 1 averaging 5", title.RatingCount, *title.RatingAverage)
	}
}

func TestRatingValidation(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	for _, stars := range []int{0, 6, -1} {
		app.expect(http.MethodPut, "/api/me/ratings/1", map[string]any{"stars": stars}, http.StatusBadRequest)
	}

	app.expect(http.MethodPut, "/api/me/ratings/1", map[string]any{"stars": 4, "review": strings.Repeat("a", 2001)}, http.StatusBadRequest)
	app.expect(http.MethodPut, "/api/me/ratings/1", map[string]any{"stars": 4, "review": strings.Repeat("é", 2000)}, http.StatusNoContent)
	app.expect(http.MethodPut, "/api/me/ratings/999", map[string]any{"stars": 4}, http.StatusNotFound)
}

func TestRatingsAndReviewsAreListed(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	app.rate(1, 5, "  Loved it  ")
	app.rate(3, 2, "")

	ratings := decode[[]model.Rating](t, app.expect(http.MethodGet, "/api/me/ratings", nil, http.StatusOK))
	if len(ratings) != 2 {
		t.Fatalf("user has %d ratings, want 2", len(ratings))
	}

	reviews := decode[[]model.Review](t, app.expect(http.MethodGet, "/api/titles/1/reviews", nil, http.StatusOK))
	if len(reviews) != 1 || reviews[0].Username != "moviefan" || reviews[0].Stars != 5 || reviews[0].Review != "Loved it" {
		t.Fatalf("reviews for title 1 are %+v", reviews)
	}

	if reviews := decode[[]model.Review](t, app.expect(http.MethodGet, "/api/titles/3/reviews", nil, http.StatusOK)); len(reviews) != 0 {
		t.Fatalf("a rating without text shows up as a review: %+v", reviews)
	}

	app.expect(http.MethodGet, "/api/titles/999/reviews", nil, http.StatusNotFound)
}

func TestTitleStatusTracksSavedWatchedAndRating(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	if status := app.status(1); status != (model.TitleStatus{}) {
		t.Fatalf("new user's status for title 1 is %+v, want empty", status)
	}

	app.expect(http.MethodPut, "/api/me/titles/1", nil, http.StatusNoContent)
	app.rate(1, 4, "Fun")

	status := app.status(1)
	if !status.Saved || !status.Watched || status.Stars != 4 || status.Review != "Fun" {
		t.Fatalf("status after saving and rating is %+v", status)
	}

	app.expect(http.MethodGet, "/api/me/titles/999", nil, http.StatusNotFound)
}

func TestWatchedList(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	app.expect(http.MethodPut, "/api/me/watched/2", nil, http.StatusNoContent)
	app.expect(http.MethodPut, "/api/me/watched/2", nil, http.StatusNoContent)
	app.expect(http.MethodPut, "/api/me/watched/999", nil, http.StatusNotFound)

	watched := titleIDs(t, app.expect(http.MethodGet, "/api/me/watched", nil, http.StatusOK))
	if !sameIDs(watched, []int64{2}) {
		t.Fatalf("watched titles are %v, want 2", watched)
	}

	app.expect(http.MethodDelete, "/api/me/watched/2", nil, http.StatusNoContent)
	if watched := titleIDs(t, app.expect(http.MethodGet, "/api/me/watched", nil, http.StatusOK)); len(watched) != 0 {
		t.Fatalf("after unmarking, watched titles are %v", watched)
	}
}

func TestRecommendationsSkipWatchedAndRatedTitles(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	app.expect(http.MethodPut, "/api/me/titles/1", nil, http.StatusNoContent)
	app.expect(http.MethodPut, "/api/me/watched/2", nil, http.StatusNoContent)
	app.rate(4, 4, "")

	recommended := titleIDs(t, app.expect(http.MethodGet, "/api/me/recommendations", nil, http.StatusOK))
	if !sameIDs(recommended, []int64{6}) {
		t.Fatalf("recommendations are %v, want only 6 (the action title not yet saved, watched or rated)", recommended)
	}
}

func TestRecommendationsUseRatings(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	app.rate(3, 1, "")
	if got := titleIDs(t, app.expect(http.MethodGet, "/api/me/recommendations", nil, http.StatusOK)); len(got) != 0 {
		t.Fatalf("after a 1-star comedy rating, recommendations are %v, want none", got)
	}

	app.rate(1, 5, "")
	got := titleIDs(t, app.expect(http.MethodGet, "/api/me/recommendations", nil, http.StatusOK))
	if !sameIDs(got, []int64{2, 4, 6}) {
		t.Fatalf("after a 5-star action rating, recommendations are %v, want 2, 4 and 6", got)
	}
}

func TestUpdateAccount(t *testing.T) {
	app := newTestApp(t)
	app.signUp("taken")

	app.client = app.newClient()
	app.signUp("moviefan")

	app.expect(http.MethodPatch, "/api/me", map[string]string{"username": "taken", "email": "new@example.com"}, http.StatusConflict)
	app.expect(http.MethodPatch, "/api/me", map[string]string{"username": "x", "email": "new@example.com"}, http.StatusBadRequest)
	app.expect(http.MethodPatch, "/api/me", map[string]string{"username": "renamed", "email": "not-an-email"}, http.StatusBadRequest)

	updated := decode[model.User](t, app.expect(http.MethodPatch, "/api/me", map[string]string{"username": "renamed", "email": " NEW@example.com "}, http.StatusOK))
	if updated.Username != "renamed" || updated.Email != "new@example.com" {
		t.Fatalf("updated account is %+v", updated)
	}

	app.client = app.newClient()
	app.expect(http.MethodPost, "/api/sessions", map[string]string{"identity": "renamed", "password": "correct-horse"}, http.StatusOK)
	app.expect(http.MethodPost, "/api/sessions", map[string]string{"identity": "moviefan", "password": "correct-horse"}, http.StatusUnauthorized)
}

func TestChangePassword(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	app.expect(http.MethodPut, "/api/me/password", map[string]string{"current_password": "wrong-password", "new_password": "brand-new-pass"}, http.StatusForbidden)
	app.expect(http.MethodPut, "/api/me/password", map[string]string{"current_password": "correct-horse", "new_password": "short"}, http.StatusBadRequest)
	app.expect(http.MethodPut, "/api/me/password", map[string]string{"current_password": "correct-horse", "new_password": "brand-new-pass"}, http.StatusNoContent)

	app.client = app.newClient()
	app.expect(http.MethodPost, "/api/sessions", map[string]string{"identity": "moviefan", "password": "correct-horse"}, http.StatusUnauthorized)
	app.expect(http.MethodPost, "/api/sessions", map[string]string{"identity": "moviefan", "password": "brand-new-pass"}, http.StatusOK)
}

func TestDeleteAccountRemovesData(t *testing.T) {
	app := newTestApp(t)
	app.signUp("moviefan")

	app.expect(http.MethodPut, "/api/me/titles/1", nil, http.StatusNoContent)
	app.rate(1, 5, "Great")

	app.expect(http.MethodDelete, "/api/me", map[string]string{"password": "wrong-password"}, http.StatusForbidden)
	app.expect(http.MethodGet, "/api/me", nil, http.StatusOK)

	app.expect(http.MethodDelete, "/api/me", map[string]string{"password": "correct-horse"}, http.StatusNoContent)
	app.expect(http.MethodGet, "/api/me", nil, http.StatusUnauthorized)
	app.expect(http.MethodPost, "/api/sessions", map[string]string{"identity": "moviefan", "password": "correct-horse"}, http.StatusUnauthorized)

	if title := app.title(1); title.RatingCount != 0 {
		t.Fatalf("deleted user's rating still counts: %d ratings", title.RatingCount)
	}

	for _, table := range []string{"sessions", "user_titles", "ratings", "user_watched"} {
		var rows int
		if err := app.raw.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&rows); err != nil {
			t.Fatal(err)
		}
		if rows != 0 {
			t.Fatalf("%s still has %d rows after deleting the only account", table, rows)
		}
	}

	app.signUp("moviefan")
}

func TestNewSignedInRoutesRequireLogin(t *testing.T) {
	app := newTestApp(t)

	routes := []struct{ method, path string }{
		{http.MethodPatch, "/api/me"},
		{http.MethodDelete, "/api/me"},
		{http.MethodPut, "/api/me/password"},
		{http.MethodGet, "/api/me/titles/1"},
		{http.MethodGet, "/api/me/ratings"},
		{http.MethodPut, "/api/me/ratings/1"},
		{http.MethodDelete, "/api/me/ratings/1"},
		{http.MethodGet, "/api/me/watched"},
		{http.MethodPut, "/api/me/watched/1"},
		{http.MethodDelete, "/api/me/watched/1"},
	}

	for _, route := range routes {
		app.expect(route.method, route.path, nil, http.StatusUnauthorized)
	}
}
