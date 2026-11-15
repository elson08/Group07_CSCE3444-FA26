package database

import (
	"context"
	"errors"
	"strings"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database/model"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var browseOrders = map[string][2]string{
	"newest": {"t.release_year DESC, t.name", "p.release_year DESC, p.name"},
	"oldest": {"t.release_year, t.name", "p.release_year, p.name"},
	"az":     {"t.name, t.release_year DESC", "p.name, p.release_year DESC"},
	"top": {
		"rs.rating_avg DESC NULLS LAST, rs.rating_count DESC, t.release_year DESC, t.name",
		"p.rating_avg DESC NULLS LAST, p.rating_count DESC, p.release_year DESC, p.name",
	},
}

func ValidSort(sort string) bool {
	_, ok := browseOrders[sort]
	return ok
}

func placeholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?, ", count), ", ")
}

func (d *Database) BrowseTitles(ctx context.Context, filter model.TitleFilter) ([]model.Title, int, error) {
	var conditions []string
	var args []any

	if filter.Query != "" {
		conditions = append(conditions, "t.name LIKE '%' || ? || '%'")
		args = append(args, filter.Query)
	}

	if filter.Kind != "" {
		conditions = append(conditions, "t.kind = ?")
		args = append(args, filter.Kind)
	}

	if len(filter.Genres) > 0 {
		conditions = append(conditions, `t.id IN (
			SELECT tg.title_id FROM title_genres tg
			JOIN genres g ON g.id = tg.genre_id
			WHERE g.name IN (`+placeholders(len(filter.Genres))+`))`)
		for _, genre := range filter.Genres {
			args = append(args, genre)
		}
	}

	if len(filter.ServiceIDs) > 0 {
		conditions = append(conditions, `t.id IN (
			SELECT title_id FROM title_services
			WHERE service_id IN (`+placeholders(len(filter.ServiceIDs))+`))`)
		for _, id := range filter.ServiceIDs {
			args = append(args, id)
		}
	}

	if filter.ServicesOfUser != 0 {
		conditions = append(conditions, `t.id IN (
			SELECT ts.title_id FROM title_services ts
			JOIN user_services us ON us.service_id = ts.service_id
			WHERE us.user_id = ?)`)
		args = append(args, filter.ServicesOfUser)
	}

	if filter.YearFrom != 0 {
		conditions = append(conditions, "t.release_year >= ?")
		args = append(args, filter.YearFrom)
	}

	if filter.YearTo != 0 {
		conditions = append(conditions, "t.release_year <= ?")
		args = append(args, filter.YearTo)
	}

	if filter.MinRating > 0 {
		conditions = append(conditions, "rs.rating_avg >= ?")
		args = append(args, filter.MinRating)
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	order, ok := browseOrders[filter.Sort]
	if !ok {
		order = browseOrders["newest"]
	}

	const stats = `WITH rating_stats AS (
		SELECT title_id, AVG(stars) AS rating_avg, COUNT(*) AS rating_count
		FROM ratings
		GROUP BY title_id
	)`
	const from = " FROM titles t LEFT JOIN rating_stats rs ON rs.title_id = t.id"

	var total int
	if err := d.conn.QueryRowContext(ctx, stats+" SELECT COUNT(*)"+from+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := stats + `, page AS (
		SELECT t.id, t.name, t.release_year, rs.rating_avg, rs.rating_count` + from + where + `
		ORDER BY ` + order[0] + `
		LIMIT ? OFFSET ?
	)
	SELECT ` + titleColumns + `
	FROM page p
	JOIN titles t ON t.id = p.id
	ORDER BY ` + order[1]

	titles, err := d.queryTitles(ctx, query, append(args, filter.Limit, filter.Offset)...)
	return titles, total, err
}

func (d *Database) AttachServices(ctx context.Context, titles []model.Title) error {
	if len(titles) == 0 {
		return nil
	}

	positions := map[int64][]int{}
	args := make([]any, 0, len(titles))
	for i, title := range titles {
		if _, seen := positions[title.ID]; !seen {
			args = append(args, title.ID)
		}
		positions[title.ID] = append(positions[title.ID], i)
	}

	rows, err := d.conn.QueryContext(ctx, `SELECT ts.title_id, s.id, s.name
		FROM title_services ts
		JOIN services s ON s.id = ts.service_id
		WHERE ts.title_id IN (`+placeholders(len(args))+`)
		ORDER BY s.name`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var titleID int64
		var service model.Service
		if err := rows.Scan(&titleID, &service.ID, &service.Name); err != nil {
			return err
		}

		for _, i := range positions[titleID] {
			titles[i].Services = append(titles[i].Services, service)
		}
	}

	return rows.Err()
}

func (d *Database) Genres(ctx context.Context) ([]model.Genre, error) {
	rows, err := d.conn.QueryContext(ctx, listGenresQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	genres := []model.Genre{}
	for rows.Next() {
		var genre model.Genre
		if err := rows.Scan(&genre.ID, &genre.Name); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}

	return genres, rows.Err()
}

func (d *Database) SimilarTitles(ctx context.Context, titleID int64, limit int) ([]model.Title, error) {
	return d.queryTitles(ctx, similarTitlesQuery, titleID, titleID, titleID, limit)
}

func (d *Database) TitleStatus(ctx context.Context, userID, titleID int64) (model.TitleStatus, error) {
	var status model.TitleStatus
	err := d.conn.QueryRowContext(ctx, titleStatusQuery, userID, titleID).
		Scan(&status.Saved, &status.Watched, &status.Stars, &status.Review)
	return status, err
}

func (d *Database) SaveRating(ctx context.Context, userID, titleID int64, stars int, review string) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, upsertRatingQuery, userID, titleID, stars, review); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, markWatchedQuery, userID, titleID); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Database) RemoveRating(ctx context.Context, userID, titleID int64) error {
	_, err := d.conn.ExecContext(ctx, removeRatingQuery, userID, titleID)
	return err
}

func (d *Database) UserRatings(ctx context.Context, userID int64) ([]model.Rating, error) {
	rows, err := d.conn.QueryContext(ctx, userRatingsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ratings := []model.Rating{}
	for rows.Next() {
		var rating model.Rating
		fields := append(titleFields(&rating.Title), &rating.Stars, &rating.Review, &rating.UpdatedAt)
		if err := rows.Scan(fields...); err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}

	return ratings, rows.Err()
}

func (d *Database) TitleReviews(ctx context.Context, titleID int64) ([]model.Review, error) {
	rows, err := d.conn.QueryContext(ctx, titleReviewsQuery, titleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []model.Review{}
	for rows.Next() {
		var review model.Review
		if err := rows.Scan(&review.Username, &review.Stars, &review.Review, &review.UpdatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	return reviews, rows.Err()
}

func (d *Database) WatchedTitles(ctx context.Context, userID int64) ([]model.Title, error) {
	return d.queryTitles(ctx, watchedTitlesQuery, userID)
}

func (d *Database) MarkWatched(ctx context.Context, userID, titleID int64) error {
	_, err := d.conn.ExecContext(ctx, markWatchedQuery, userID, titleID)
	return err
}

func (d *Database) UnmarkWatched(ctx context.Context, userID, titleID int64) error {
	_, err := d.conn.ExecContext(ctx, unmarkWatchedQuery, userID, titleID)
	return err
}

func (d *Database) UpdateUser(ctx context.Context, userID int64, username, email string) (model.User, error) {
	if _, err := d.conn.ExecContext(ctx, updateUserQuery, username, email, userID); err != nil {
		var sqliteError *sqlite.Error
		if errors.As(err, &sqliteError) && sqliteError.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return model.User{}, ErrUserExists
		}
		return model.User{}, err
	}

	return d.UserByID(ctx, userID)
}

func (d *Database) UpdatePassword(ctx context.Context, userID int64, hash string) error {
	_, err := d.conn.ExecContext(ctx, updatePasswordQuery, hash, userID)
	return err
}

func (d *Database) DeleteUser(ctx context.Context, userID int64) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, table := range []string{"sessions", "user_titles", "user_services", "ratings", "user_watched"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE user_id = ?", userID); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID); err != nil {
		return err
	}

	return tx.Commit()
}
