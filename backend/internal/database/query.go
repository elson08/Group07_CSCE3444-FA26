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
		t.release_year, t.synopsis, t.poster_url, t.trailer_url`

	listTitlesQuery = `SELECT ` + titleColumns + `
		FROM titles t
		WHERE (? = '' OR t.name LIKE '%' || ? || '%')
			AND (? = '' OR t.kind = ?)
		ORDER BY t.name
		LIMIT ? OFFSET ?`

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
		ORDER BY t.name`

	saveTitleQuery = `INSERT OR IGNORE INTO user_titles (user_id, title_id)
		VALUES (?, ?)`

	removeTitleQuery = `DELETE FROM user_titles
		WHERE user_id = ? AND title_id = ?`
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
