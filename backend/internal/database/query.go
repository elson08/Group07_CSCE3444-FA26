package database

// Enable the relationships defined by the schema.
const enableForeignKeysQuery = "PRAGMA foreign_keys = ON"

// Account and session queries.
const (
	createUserQuery = `INSERT INTO users (username, email, password_hash)
		VALUES (?, ?, ?)`

	userByIDQuery = `SELECT id, username, email, password_hash, created_at
		FROM users
		WHERE id = ?`

	userByIdentityQuery = `SELECT id, username, email, password_hash, created_at
		FROM users
		WHERE lower(username) = lower(?) OR lower(email) = lower(?)`

	createSessionQuery = `INSERT INTO sessions (token_hash, user_id, expires_at)
		VALUES (?, ?, ?)`

	userBySessionQuery = `SELECT u.id, u.username, u.email, u.password_hash, u.created_at
		FROM users u
		JOIN sessions s ON s.user_id = u.id
		WHERE s.token_hash = ? AND s.expires_at > ?`

	deleteSessionQuery = `DELETE FROM sessions WHERE token_hash = ?`
)

// Title and saved-title queries.
const (
	titleColumns = `t.id, t.provider_id, t.kind, t.name,
		t.release_year, t.synopsis, t.poster_url, t.trailer_url,
		(SELECT ROUND(AVG(r.stars), 1) FROM ratings r WHERE r.title_id = t.id),
		(SELECT COUNT(*) FROM ratings r WHERE r.title_id = t.id)`

	titleByIDQuery = `SELECT ` + titleColumns + `
		FROM titles t
		WHERE t.id = ?`

	titleExistsQuery = `SELECT EXISTS(SELECT 1 FROM titles WHERE id = ?)`

	genresByTitleQuery = `SELECT g.id, g.name
		FROM genres g
		JOIN title_genres tg ON tg.genre_id = g.id
		WHERE tg.title_id = ?
		ORDER BY g.name`

	servicesByTitleQuery = `SELECT s.id, s.name
		FROM services s
		JOIN title_services ts ON ts.service_id = s.id
		WHERE ts.title_id = ?
		ORDER BY s.name`

	savedTitlesQuery = `SELECT ` + titleColumns + `
		FROM titles t
		JOIN user_titles ut ON ut.title_id = t.id
		WHERE ut.user_id = ?
		ORDER BY ut.rowid DESC`

	saveTitleQuery = `INSERT OR IGNORE INTO user_titles (user_id, title_id)
		VALUES (?, ?)`

	removeTitleQuery = `DELETE FROM user_titles
		WHERE user_id = ? AND title_id = ?`

	recommendedTitlesQuery = `WITH signals AS (
			SELECT title_id, 1 AS weight FROM user_titles WHERE user_id = ?
			UNION ALL
			SELECT title_id, stars - 3 FROM ratings WHERE user_id = ?
		), liked AS (
			SELECT tg.genre_id, SUM(s.weight) AS weight
			FROM signals s
			JOIN title_genres tg ON tg.title_id = s.title_id
			GROUP BY tg.genre_id
			HAVING SUM(s.weight) > 0
		), scored AS (
			SELECT tg.title_id, SUM(l.weight) AS score
			FROM title_genres tg
			JOIN liked l ON l.genre_id = tg.genre_id
			GROUP BY tg.title_id
		)
		SELECT ` + titleColumns + `
		FROM scored s
		JOIN titles t ON t.id = s.title_id
		WHERE t.id NOT IN (SELECT title_id FROM user_titles WHERE user_id = ?)
			AND t.id NOT IN (SELECT title_id FROM user_watched WHERE user_id = ?)
			AND t.id NOT IN (SELECT title_id FROM ratings WHERE user_id = ?)
		ORDER BY EXISTS(
				SELECT 1
				FROM title_services ts
				JOIN user_services us ON us.service_id = ts.service_id
				WHERE ts.title_id = t.id AND us.user_id = ?
			) DESC,
			s.score DESC, t.release_year DESC, t.name
		LIMIT ?`
)

const (
	hasTitlesQuery = `SELECT EXISTS(SELECT 1 FROM titles)`

	insertTitleQuery = `INSERT INTO titles (provider_id, kind, name, release_year)
		VALUES (?, ?, ?, ?)`

	upsertGenreQuery = `INSERT INTO genres (name)
		VALUES (?)
		ON CONFLICT(name) DO UPDATE SET name = excluded.name
		RETURNING id`

	linkTitleGenreQuery = `INSERT INTO title_genres (title_id, genre_id)
		VALUES (?, ?)`

	saveTitleDetailsQuery = `UPDATE titles
		SET synopsis = ?, poster_url = NULLIF(?, ''), trailer_url = NULLIF(?, '')
		WHERE id = ?`

	upsertServiceQuery = `INSERT INTO services (name)
		VALUES (?)
		ON CONFLICT(name) DO UPDATE SET name = excluded.name
		RETURNING id`

	linkTitleServiceQuery = `INSERT OR IGNORE INTO title_services (title_id, service_id)
		VALUES (?, ?)`
)

const (
	listGenresQuery = `SELECT id, name FROM genres ORDER BY name`

	similarTitlesQuery = `WITH source AS (
			SELECT genre_id FROM title_genres WHERE title_id = ?
		), scored AS (
			SELECT tg.title_id, COUNT(*) AS score
			FROM title_genres tg
			JOIN source s ON s.genre_id = tg.genre_id
			WHERE tg.title_id != ?
			GROUP BY tg.title_id
		)
		SELECT ` + titleColumns + `
		FROM scored sc
		JOIN titles t ON t.id = sc.title_id
		JOIN titles origin ON origin.id = ?
		WHERE t.kind = origin.kind
		ORDER BY sc.score DESC,
			ABS(COALESCE(t.release_year, 0) - COALESCE(origin.release_year, 0)),
			t.name
		LIMIT ?`

	titleStatusQuery = `SELECT
			EXISTS(SELECT 1 FROM user_titles WHERE user_id = ?1 AND title_id = ?2),
			EXISTS(SELECT 1 FROM user_watched WHERE user_id = ?1 AND title_id = ?2),
			COALESCE((SELECT stars FROM ratings WHERE user_id = ?1 AND title_id = ?2), 0),
			COALESCE((SELECT review FROM ratings WHERE user_id = ?1 AND title_id = ?2), '')`

	upsertRatingQuery = `INSERT INTO ratings (user_id, title_id, stars, review)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, title_id) DO UPDATE
		SET stars = excluded.stars, review = excluded.review, updated_at = CURRENT_TIMESTAMP`

	removeRatingQuery = `DELETE FROM ratings WHERE user_id = ? AND title_id = ?`

	userRatingsQuery = `SELECT ` + titleColumns + `, ur.stars, ur.review, ur.updated_at
		FROM ratings ur
		JOIN titles t ON t.id = ur.title_id
		WHERE ur.user_id = ?
		ORDER BY ur.updated_at DESC`

	titleReviewsQuery = `SELECT u.username, r.stars, r.review, r.updated_at
		FROM ratings r
		JOIN users u ON u.id = r.user_id
		WHERE r.title_id = ? AND r.review != ''
		ORDER BY r.updated_at DESC
		LIMIT 50`

	watchedTitlesQuery = `SELECT ` + titleColumns + `
		FROM titles t
		JOIN user_watched w ON w.title_id = t.id
		WHERE w.user_id = ?
		ORDER BY w.watched_at DESC, t.name`

	markWatchedQuery = `INSERT OR IGNORE INTO user_watched (user_id, title_id)
		VALUES (?, ?)`

	unmarkWatchedQuery = `DELETE FROM user_watched WHERE user_id = ? AND title_id = ?`

	updateUserQuery = `UPDATE users SET username = ?, email = ? WHERE id = ?`

	updatePasswordQuery = `UPDATE users SET password_hash = ? WHERE id = ?`
)

// Streaming service and saved-service queries.
const (
	listServicesQuery = `SELECT id, name FROM services ORDER BY name`

	userServicesQuery = `SELECT s.id, s.name
		FROM services s
		JOIN user_services us ON us.service_id = s.id
		WHERE us.user_id = ?
		ORDER BY s.name`

	serviceExistsQuery = `SELECT EXISTS(SELECT 1 FROM services WHERE id = ?)`

	saveServiceQuery = `INSERT OR IGNORE INTO user_services (user_id, service_id)
		VALUES (?, ?)`

	removeServiceQuery = `DELETE FROM user_services
		WHERE user_id = ? AND service_id = ?`
)
