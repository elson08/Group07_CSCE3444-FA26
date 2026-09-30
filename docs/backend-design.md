# Backend API

Local base URL: `http://localhost:8080`. Requests and responses use JSON.
Login sets a session cookie. Routes marked **Yes** require that cookie.

| Endpoint | What it does | Login required |
| --- | --- | --- |
| `GET /api/health` | Checks that the backend can reach the database. | No |
| `POST /api/users` | Creates an account. Send `username`, `email`, and `password`; then log in separately. | No |
| `POST /api/sessions` | Logs in and sets the session cookie. Send `identity` (username or email) and `password`. | No |
| `DELETE /api/sessions/current` | Logs out and clears the session cookie. | No |
| `GET /api/me` | Returns the logged-in user's account information. | Yes |
| `GET /api/titles` | Lists movies and shows. Optional query parameters: `q` (title search), `kind` (`movie` or `show`), `limit` (default 20, maximum 50), and `offset` (default 0). | No |
| `GET /api/titles/{id}` | Returns one title's details, genres, and streaming services. | No |
| `GET /api/services` | Lists streaming services. | No |
| `GET /api/me/titles` | Lists the user's saved movies and shows. | Yes |
| `PUT /api/me/titles/{id}` | Saves a title to the user's list. No request body needed. | Yes |
| `DELETE /api/me/titles/{id}` | Removes a title from the user's list. | Yes |
| `GET /api/me/services` | Lists the user's selected streaming services. | Yes |
| `PUT /api/me/services/{id}` | Adds a streaming service to the user's selections. No request body needed. | Yes |
| `DELETE /api/me/services/{id}` | Removes a streaming service from the user's selections. | Yes |

`{id}` is the numeric title or service ID returned by the API.
Successful registration returns `201`; logout and saved-list changes return `204` with no response body. Other successful requests return `200`.
Errors return `{"error":"message"}`. Protected routes return `401` when the session is missing or expired.

The current frontend integration requires the frontend and API to share an origin.
