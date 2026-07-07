# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go run main.go          # Run the server (reads .env, requires JWT_SECRET)
go build -o gopost-api   # Build binary
go fmt ./...             # Format
go vet ./...             # Lint
```

No test suite exists yet in this repo.

### Database setup
MySQL 8+ required. Schema lives at `database/schema.sql` (drops and recreates `users`/`posts` tables):
```bash
mysql -u root -p gopost < database/schema.sql
```

Required `.env` (not committed): `PORT` (default `:5050`), `JWT_SECRET` (required, no default — `config.LoadConfig` calls `log.Fatal` if missing), `DATABASE_URL` (default `root:password@tcp(localhost:3306)/gopost`).

## Architecture

Layered: **Handler → Service → Repository → Database (MySQL via `database/sql`)**. Each layer only talks to the one directly below it; handlers never touch `repositories` or `database/sql` directly, and services never touch `net/http`.

- `models/` — plain structs (`User`, `Post`, `SignUpInput`, `LoginInput`) shared across all layers, including JSON tags.
- `repositories/` — raw SQL via prepared statements (`db.QueryRowContext`/`ExecContext`), one struct per table (`UserRepository`, `PostRepository`), constructed with `New*Repository(db *sql.DB)`.
- `services/` — business logic and validation (email/password format, bcrypt hashing/comparison, JWT creation), constructed with `New*Service(repo)`. Returns domain errors as plain `fmt.Errorf` strings, which handlers surface to the client — so service error messages are user-facing, keep them in Spanish and non-revealing (e.g. login always returns "credenciales incorrectas", never distinguishing missing email vs wrong password).
- `handlers/` — HTTP-facing: decode request, call service, map result/error to a `server.Context` JSON response. `handlers/response.go` defines `AppError`/`RespondError` for consistent `{error, message, code}` bodies.
- `server/` — a small custom HTTP framework (no external router library), built on `net/http.ServeMux` and Go 1.22+ method+path patterns (`"GET /posts/{id}"`):
  - `server/server.go`: `App` wraps `*http.ServeMux`, `RunServer` boots `http.Server` and prints a banner.
  - `server/router.go`: `App.Get/Post/Put/Delete(path, handler)` register routes; every handler receives a fresh `*server.Context` wrapping `http.ResponseWriter`/`*http.Request`.
  - `server/context.go`: `Context` has `JSON`, `BindJSON`, `Send`, `Status` helpers, plus `SetUserID`/`GetUserID` for passing the authenticated user id from middleware to handlers. Path params are read directly via `c.Request.PathValue("id")` (stdlib, not a custom mux feature).
- `middleware/auth.go` — `AuthMiddleware(next server.HandleFunc) server.HandleFunc` decorator: parses `Authorization: Bearer <token>`, validates the JWT against `config.AppConfig.JWTSecret`, extracts `user_id` claim, calls `c.SetUserID`, then invokes `next`. Applied per-route in `main.go`, not globally.
- `config/config.go` — loads `.env` via `godotenv`, exposes a package-level `config.AppConfig` singleton (read by `middleware/auth.go` and `services/user_service.go` for the JWT secret).
- `main.go` — composition root: wires `database.Connect` → repositories → services → handlers → routes, in that order. New endpoints are added here by registering `app.Get/Post/Put/Delete` with `middleware.AuthMiddleware(...)` wrapping any handler that requires auth.

### Adding a new resource
Follow the existing Post/User pattern: add a model in `models/`, a repository in `repositories/` (SQL + errors wrapped with `fmt.Errorf("...: %w", err)`), a service in `services/` (validation + business rules), a handler in `handlers/` (bind → call service → `c.JSON`/`RespondError`), then wire it in `main.go`.

### Auth flow
Passwords hashed with bcrypt (`bcrypt.DefaultCost`). JWT signed HS256, `user_id` + `exp` (72h) claims, secret from `config.AppConfig.JWTSecret`. Ownership checks (e.g. only a post's author can update/delete it) are enforced in the service layer by comparing `c.GetUserID()` (passed in from the handler) against the resource's `user_id`.
