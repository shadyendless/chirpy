# chirpy

`chirpy` is a small Twitter-like HTTP API written in Go. It lets you register
users, log in with JWT-based auth, post short messages ("chirps"), browse and
filter them, and delete your own. It also includes a webhook endpoint for
upgrading users to "Chirpy Red" and an admin endpoint that reports how many
times the served static app has been visited.

## Prerequisites

To run `chirpy` you'll need the following installed on your machine:

- **[Go](https://go.dev/doc/install)** (1.26+) — used to build and run the server.
- **[PostgreSQL](https://www.postgresql.org/download/)** (15+) — the database where users, chirps, and refresh tokens are stored.

Make sure Postgres is running and that you've created a database for `chirpy`
to use, for example:

```bash
createdb chirpy
```

## Configuration

`chirpy` reads its configuration from environment variables. In development it
loads them from a `.env` file in the project root (via
[godotenv](https://github.com/joho/godotenv)), so create a `.env` file like:

```bash
DB_URL="postgres://username:password@localhost:5432/chirpy?sslmode=disable"
PLATFORM=dev
SECRET="your-jwt-signing-secret"
POLKA_KEY="your-polka-webhook-api-key"
```

- `DB_URL` — the connection string for your Postgres database.
- `PLATFORM` — set to `dev` to enable the `POST /admin/reset` endpoint (which
  wipes all users). Leave it as anything else in production to keep that
  endpoint disabled.
- `SECRET` — the secret used to sign and verify JWT access tokens. Use a long,
  random value. You can generate one with `openssl rand -base64 64`.
- `POLKA_KEY` — the API key the Polka webhook must present (via
  `Authorization: ApiKey <key>`) to upgrade a user to Chirpy Red.

The `.env` file is git-ignored, so your secrets stay out of version control.

## Setting up the database

`chirpy` does **not** create its tables automatically — the schema must already
exist. Migrations are managed with
[Goose](https://github.com/pressly/goose) and live in the `sql/schema`
directory of this repository.

1. Install the Goose CLI:

   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   ```

2. From the project root, move into the schema directory:

   ```bash
   cd sql/schema
   ```

3. Run all the "up" migrations against your database, using the same connection
   string you put in `.env`:

   ```bash
   goose postgres "postgres://username:password@localhost:5432/chirpy" up
   ```

You only need to do this once (and again whenever new migrations are added). To
roll the most recent migration back, run `goose postgres "<db_url>" down`.

> **Note:** the Go type-safe database layer in `internal/database` is generated
> from the SQL in `sql/schema` and `sql/queries` with
> [sqlc](https://sqlc.dev/). You don't need sqlc to run the server — only if you
> change the queries and want to regenerate the code with `sqlc generate`.

## Running the server

From the project root, run:

```bash
go run .
```

Or build a binary and run it:

```bash
go build -o out .
./out
```

The server listens on **`http://localhost:8080`**. It also serves the static
`index.html` at `/app/` and increments a hit counter on each visit.

## API endpoints

| Method & Path | Description |
| --- | --- |
| `GET /api/healthz` | Health check; returns `OK`. |
| `POST /api/users` | Create a user. Body: `{ "email", "password" }`. |
| `PUT /api/users` | Update the logged-in user's email/password (requires bearer token). |
| `POST /api/login` | Log in; returns an access token and a refresh token. |
| `POST /api/refresh` | Exchange a refresh token for a new access token. |
| `POST /api/revoke` | Revoke a refresh token. |
| `POST /api/chirps` | Create a chirp (requires bearer token). Max 140 chars. |
| `GET /api/chirps` | List chirps. Optional `?author_id=<uuid>` and `?sort=asc\|desc`. |
| `GET /api/chirps/{chirpID}` | Get a single chirp by ID. |
| `DELETE /api/chirps/{chirpID}` | Delete your own chirp (requires bearer token). |
| `POST /api/polka/webhooks` | Upgrade a user to Chirpy Red (requires `ApiKey` auth). |
| `GET /admin/metrics` | HTML page showing the fileserver hit count. |
| `POST /admin/reset` | Delete all users (only when `PLATFORM=dev`). |
| `GET /app/` | Serve the static app. |

### Example workflow

```bash
# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"hunter2"}'

# Log in and grab the "token" from the response
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"hunter2"}'

# Post a chirp (use the access token from login)
curl -X POST http://localhost:8080/api/chirps \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"body":"Hello, chirpy!"}'

# List all chirps, newest first
curl "http://localhost:8080/api/chirps?sort=desc"
```
