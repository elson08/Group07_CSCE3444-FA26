package database

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

const testRatings = `tconst	averageRating	numVotes
tt0000001	7.1	5000
tt0000002	6.0	10
tt0000003	8.0	2000
tt0000004	7.5	3000
tt0000005	6.2	4000
tt0000006	6.8	1500
tt0000007	5.0	1000
`

const testBasics = `tconst	titleType	primaryTitle	originalTitle	isAdult	startYear	endYear	runtimeMinutes	genres
tt0000001	movie	Popular Movie	Popular Movie	0	2020	\N	120	Action,Drama
tt0000002	movie	Obscure Movie	Obscure Movie	0	2021	\N	90	Drama
tt0000003	movie	Adult Movie	Adult Movie	1	2019	\N	80	Drama
tt0000004	tvSeries	Popular Show	Popular Show	0	2015	2020	45	Comedy
tt0000005	tvEpisode	An Episode	An Episode	0	2016	\N	30	Comedy
tt0000006	movie	Unknown Year	Unknown Year	0	\N	\N	\N	\N
tt0000007	tvMiniSeries	Short Series	Short Series	0	2022	2022	50	Drama,History
broken row
`

func newTestDatabase(t *testing.T) *Database {
	t.Helper()

	db, err := New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	return db
}

func gzipped(t *testing.T, text string) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write([]byte(text)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	return buffer.Bytes()
}

func fakeIMDb(t *testing.T, ratings, basics string) (string, *atomic.Int32) {
	t.Helper()

	files := map[string][]byte{
		"/title.ratings.tsv.gz": gzipped(t, ratings),
		"/title.basics.tsv.gz":  gzipped(t, basics),
	}

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		data, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}

		w.Write(data)
	}))
	t.Cleanup(server.Close)

	return server.URL + "/", &requests
}

func TestSeedCatalogImportsOnlyPopularMoviesAndShows(t *testing.T) {
	db := newTestDatabase(t)
	baseURL, _ := fakeIMDb(t, testRatings, testBasics)

	count, err := db.seedCatalog(context.Background(), baseURL)
	if err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatalf("imported %d titles, want 4", count)
	}

	rows, err := db.conn.Query(`SELECT provider_id, kind FROM titles ORDER BY provider_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var id, kind string
		if err := rows.Scan(&id, &kind); err != nil {
			t.Fatal(err)
		}
		got = append(got, id+":"+kind)
	}

	want := "tt0000001:movie,tt0000004:show,tt0000006:movie,tt0000007:show"
	if strings.Join(got, ",") != want {
		t.Fatalf("imported %v, want %s", got, want)
	}
}

func TestSeedCatalogStoresGenresAndYears(t *testing.T) {
	db := newTestDatabase(t)
	baseURL, _ := fakeIMDb(t, testRatings, testBasics)

	if _, err := db.seedCatalog(context.Background(), baseURL); err != nil {
		t.Fatal(err)
	}

	var genres string
	err := db.conn.QueryRow(`SELECT group_concat(g.name, ',') FROM (
			SELECT g.name FROM genres g
			JOIN title_genres tg ON tg.genre_id = g.id
			JOIN titles t ON t.id = tg.title_id
			WHERE t.provider_id = 'tt0000001'
			ORDER BY g.name
		) g`).Scan(&genres)
	if err != nil {
		t.Fatal(err)
	}
	if genres != "Action,Drama" {
		t.Fatalf("genres for tt0000001 are %q, want Action,Drama", genres)
	}

	var year sql.NullInt64
	var genreCount int
	err = db.conn.QueryRow(`SELECT t.release_year, (SELECT COUNT(*) FROM title_genres WHERE title_id = t.id)
		FROM titles t WHERE t.provider_id = 'tt0000006'`).Scan(&year, &genreCount)
	if err != nil {
		t.Fatal(err)
	}
	if year.Valid || genreCount != 0 {
		t.Fatalf("title with unknown year and genres stored year %v and %d genres", year, genreCount)
	}

	var distinctGenres int
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM genres`).Scan(&distinctGenres); err != nil {
		t.Fatal(err)
	}
	if distinctGenres != 4 {
		t.Fatalf("stored %d genres, want 4 (Action, Comedy, Drama, History)", distinctGenres)
	}
}

func TestSeedCatalogSkipsWhenCatalogExists(t *testing.T) {
	db := newTestDatabase(t)
	baseURL, requests := fakeIMDb(t, testRatings, testBasics)

	if _, err := db.seedCatalog(context.Background(), baseURL); err != nil {
		t.Fatal(err)
	}
	before := requests.Load()

	count, err := db.seedCatalog(context.Background(), baseURL)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 || requests.Load() != before {
		t.Fatalf("second import added %d titles and made %d downloads, want none", count, requests.Load()-before)
	}
}

func TestSeedCatalogFailsCleanlyWithoutMatches(t *testing.T) {
	db := newTestDatabase(t)
	baseURL, _ := fakeIMDb(t, "tconst\taverageRating\tnumVotes\n", testBasics)

	if _, err := db.seedCatalog(context.Background(), baseURL); err == nil {
		t.Fatal("import with no matching titles succeeded")
	}

	var titles int
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM titles`).Scan(&titles); err != nil {
		t.Fatal(err)
	}
	if titles != 0 {
		t.Fatalf("failed import left %d titles behind", titles)
	}
}

func TestSeedCatalogReportsDownloadErrors(t *testing.T) {
	db := newTestDatabase(t)

	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	if _, err := db.seedCatalog(context.Background(), server.URL+"/"); err == nil {
		t.Fatal("import succeeded although the download failed")
	}
}

func TestCreateUserRejectsDuplicates(t *testing.T) {
	db := newTestDatabase(t)
	ctx := context.Background()

	if _, err := db.CreateUser(ctx, "moviefan", "fan@example.com", "hash"); err != nil {
		t.Fatal(err)
	}

	if _, err := db.CreateUser(ctx, "MOVIEFAN", "other@example.com", "hash"); !errors.Is(err, ErrUserExists) {
		t.Fatalf("duplicate username error is %v, want ErrUserExists", err)
	}

	if _, err := db.CreateUser(ctx, "another", "fan@example.com", "hash"); !errors.Is(err, ErrUserExists) {
		t.Fatalf("duplicate email error is %v, want ErrUserExists", err)
	}
}

func TestSessionsExpireAndCanBeDeleted(t *testing.T) {
	db := newTestDatabase(t)
	ctx := context.Background()

	user, err := db.CreateUser(ctx, "moviefan", "fan@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}

	if err := db.CreateSession(ctx, "active", user.ID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSession(ctx, "expired", user.ID, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	found, err := db.UserBySession(ctx, "active")
	if err != nil || found.ID != user.ID {
		t.Fatalf("active session returned %+v, %v", found, err)
	}

	if _, err := db.UserBySession(ctx, "expired"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expired session error is %v, want sql.ErrNoRows", err)
	}

	if err := db.DeleteSession(ctx, "active"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.UserBySession(ctx, "active"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted session error is %v, want sql.ErrNoRows", err)
	}
}
