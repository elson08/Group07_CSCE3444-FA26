package database

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database/model"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

//go:embed schema.sql
var schema string

// ErrUserExists means a username or email is already registered.
var ErrUserExists = errors.New("username or email already in use")

// Database owns the SQLite connection.
type Database struct {
	conn *sql.DB
}

type rowScanner interface{ Scan(...any) error }

// New opens SQLite and initializes its schema.
func New(path string) (*Database, error) {
	if path == "" {
		path = "data/local.sqlite"
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(1)
	for _, statement := range []string{enableForeignKeysQuery, schema} {
		if _, err := conn.Exec(statement); err != nil {
			conn.Close()
			return nil, err
		}
	}

	return &Database{conn: conn}, nil
}

// Ping checks that the database is reachable.
func (d *Database) Ping(ctx context.Context) error {
	return d.conn.PingContext(ctx)
}

// Close closes the database connection.
func (d *Database) Close() error {
	return d.conn.Close()
}

// CreateUser stores a new account and returns it.
func (d *Database) CreateUser(ctx context.Context, username, email, hash string) (model.User, error) {
	result, err := d.conn.ExecContext(ctx, createUserQuery, username, email, hash)
	if err != nil {
		var sqliteError *sqlite.Error
		if errors.As(err, &sqliteError) && sqliteError.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return model.User{}, ErrUserExists
		}

		return model.User{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.User{}, err
	}

	return d.UserByID(ctx, id)
}

// UserByID finds an account by its ID.
func (d *Database) UserByID(ctx context.Context, id int64) (model.User, error) {
	return scanUser(d.conn.QueryRowContext(ctx, userByIDQuery, id))
}

// UserByIdentity finds an account by username or email.
func (d *Database) UserByIdentity(ctx context.Context, identity string) (model.User, error) {
	return scanUser(d.conn.QueryRowContext(ctx, userByIdentityQuery, identity, identity))
}

// CreateSession stores a hashed login token.
func (d *Database) CreateSession(ctx context.Context, hash string, userID int64, expires time.Time) error {
	_, err := d.conn.ExecContext(ctx, createSessionQuery, hash, userID, expires.Unix())
	return err
}

// UserBySession finds the user for an unexpired session.
func (d *Database) UserBySession(ctx context.Context, hash string) (model.User, error) {
	return scanUser(d.conn.QueryRowContext(ctx, userBySessionQuery, hash, time.Now().Unix()))
}

// DeleteSession removes a login session.
func (d *Database) DeleteSession(ctx context.Context, hash string) error {
	_, err := d.conn.ExecContext(ctx, deleteSessionQuery, hash)
	return err
}

// ListTitles searches the catalog with pagination.
func (d *Database) ListTitles(ctx context.Context, query, kind string, limit, offset int) ([]model.Title, error) {
	return d.queryTitles(ctx, listTitlesQuery, query, query, kind, kind, limit, offset)
}

// TitleByID loads a title with its genres and services.
func (d *Database) TitleByID(ctx context.Context, id int64) (model.Title, error) {
	title, err := scanTitle(d.conn.QueryRowContext(ctx, titleByIDQuery, id))
	if err != nil {
		return title, err
	}

	title.Genres, err = d.titleGenres(ctx, id)
	if err != nil {
		return title, err
	}

	title.Services, err = d.queryServices(ctx, servicesByTitleQuery, id)
	return title, err
}

func (d *Database) SaveTitleDetails(ctx context.Context, titleID int64, details model.TitleDetails) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, saveTitleDetailsQuery, details.Synopsis, details.PosterURL, details.TrailerURL, titleID); err != nil {
		return err
	}

	for _, name := range details.Services {
		var serviceID int64
		if err := tx.QueryRowContext(ctx, upsertServiceQuery, name).Scan(&serviceID); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, linkTitleServiceQuery, titleID, serviceID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// TitleExists reports whether a title ID exists.
func (d *Database) TitleExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := d.conn.QueryRowContext(ctx, titleExistsQuery, id).Scan(&exists)
	return exists, err
}

// SavedTitles lists the titles saved by a user.
func (d *Database) SavedTitles(ctx context.Context, userID int64) ([]model.Title, error) {
	return d.queryTitles(ctx, savedTitlesQuery, userID)
}

func (d *Database) RecommendedTitles(ctx context.Context, userID int64, limit int) ([]model.Title, error) {
	return d.queryTitles(ctx, recommendedTitlesQuery, userID, userID, userID, limit)
}

// SaveTitle adds a title to a user’s saved list.
func (d *Database) SaveTitle(ctx context.Context, userID, titleID int64) error {
	_, err := d.conn.ExecContext(ctx, saveTitleQuery, userID, titleID)
	return err
}

// RemoveTitle removes a title from a user’s saved list.
func (d *Database) RemoveTitle(ctx context.Context, userID, titleID int64) error {
	_, err := d.conn.ExecContext(ctx, removeTitleQuery, userID, titleID)
	return err
}

// Services lists all services or those saved by a user.
func (d *Database) Services(ctx context.Context, userID *int64) ([]model.Service, error) {
	if userID != nil {
		return d.queryServices(ctx, userServicesQuery, *userID)
	}

	return d.queryServices(ctx, listServicesQuery)
}

// ServiceExists reports whether a service ID exists.
func (d *Database) ServiceExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := d.conn.QueryRowContext(ctx, serviceExistsQuery, id).Scan(&exists)
	return exists, err
}

// SaveService adds a service to a user’s list.
func (d *Database) SaveService(ctx context.Context, userID, serviceID int64) error {
	_, err := d.conn.ExecContext(ctx, saveServiceQuery, userID, serviceID)
	return err
}

// RemoveService removes a service from a user’s list.
func (d *Database) RemoveService(ctx context.Context, userID, serviceID int64) error {
	_, err := d.conn.ExecContext(ctx, removeServiceQuery, userID, serviceID)
	return err
}

// scanUser reads a user from a database row.
func scanUser(row rowScanner) (model.User, error) {
	var user model.User
	var created string

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &created)
	if err != nil {
		return user, err
	}

	user.CreatedAt, err = time.Parse("2006-01-02 15:04:05", created)
	return user, err
}

// scanTitle reads a title from a database row.
func scanTitle(row rowScanner) (model.Title, error) {
	var title model.Title
	err := row.Scan(
		&title.ID, &title.ProviderID, &title.Kind, &title.Name,
		&title.ReleaseYear, &title.Synopsis, &title.PosterURL, &title.TrailerURL,
	)

	return title, err
}

// queryTitles runs a query and reads its title rows.
func (d *Database) queryTitles(ctx context.Context, query string, args ...any) ([]model.Title, error) {
	rows, err := d.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	titles := []model.Title{}

	for rows.Next() {
		title, err := scanTitle(rows)
		if err != nil {
			return nil, err
		}

		titles = append(titles, title)
	}

	return titles, rows.Err()
}

// titleGenres loads the genres assigned to a title.
func (d *Database) titleGenres(ctx context.Context, titleID int64) ([]model.Genre, error) {
	rows, err := d.conn.QueryContext(ctx, genresByTitleQuery, titleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []model.Genre
	for rows.Next() {
		var genre model.Genre
		if err := rows.Scan(&genre.ID, &genre.Name); err != nil {
			return nil, err
		}

		genres = append(genres, genre)
	}

	return genres, rows.Err()
}

// queryServices runs a query and reads its service rows.
func (d *Database) queryServices(ctx context.Context, query string, args ...any) ([]model.Service, error) {
	rows, err := d.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []model.Service
	for rows.Next() {
		var service model.Service
		if err := rows.Scan(&service.ID, &service.Name); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, rows.Err()
}
