<img width="1848" height="920" alt="Screenshot from 2026-05-24 13-16-49" src="https://github.com/user-attachments/assets/6408ca85-25ac-4483-9c2d-6cbf192111b3" />

# TaskFlow

TaskFlow is a fullstack task management application. It supports user
authentication, projects with members, tasks with priorities and deadlines, a
drag-and-drop Kanban board, a summary dashboard, and real-time notifications.

- Backend: Go (Fiber) REST API with JWT auth and PostgreSQL
- Frontend: Next.js 14 (App Router) with TypeScript, TanStack Query, and Tailwind
- Infrastructure: Docker Compose runs the database, backend, and frontend with a single command

## Table of contents

- [Quick start (Docker)](#quick-start-docker)
- [Seed data and demo credentials](#seed-data-and-demo-credentials)
- [Service URLs](#service-urls)
- [API documentation](#api-documentation)
- [Local development (without Docker)](#local-development-without-docker)
- [Environment variables](#environment-variables)
- [Project structure](#project-structure)
- [Architecture overview](#architecture-overview)
- [Design decisions](#design-decisions)
- [Security](#security)
- [Testing](#testing)

## Quick start (Docker)

Requirements: Docker and the Docker Compose plugin.

```bash
docker compose up --build
```

This starts three services and requires no manual configuration:

1. `db` — PostgreSQL 15. The backend waits for its healthcheck before starting.
2. `backend` — runs database migrations, seeds demo data on first boot, then serves the API.
3. `frontend` — the Next.js production server.

Open http://localhost:3000 and log in with one of the seeded accounts below.

To stop the stack press Ctrl+C. To remove the database volume and start from a
clean slate:

```bash
docker compose down -v
```

## Seed data and demo credentials

Demo data is seeded automatically on first boot (controlled by `APP_SEED`, which
the Compose stack sets to `true`). The seeder is idempotent: it checks for the
seed users and does nothing if they already exist.

It creates two users, two projects, and five tasks:

| Email             | Password    | Role across projects                   |
| ----------------- | ----------- | -------------------------------------- |
| alice@example.com | password123 | Owner of Project Alpha, member of Beta |
| bob@example.com   | password123 | Owner of Project Beta, member of Alpha |

## Service URLs

| Service           | URL                                      |
| ----------------- | ---------------------------------------- |
| Frontend          | http://localhost:3000                    |
| API base          | http://localhost:8080/api/v1             |
| Swagger UI        | http://localhost:8080/swagger/index.html |
| Health check      | http://localhost:8080/health             |
| PostgreSQL (host) | localhost:5433 (container port 5432)     |

## API documentation

The API is documented with OpenAPI/Swagger, generated from annotations on the
handlers. With the stack running, browse the interactive UI at
http://localhost:8080/swagger/index.html.

The raw specification is also checked into the repository:

- `backend/docs/swagger.json`
- `backend/docs/swagger.yaml`

All responses share one envelope:

```json
{ "data": <payload or null>, "message": "human readable message", "error": "machine code or null" }
```

## Local development (without Docker)

Useful when iterating on a single service. You need Go 1.25+, Node.js 20+, and a
running PostgreSQL instance.

### Backend

```bash
cd backend
cp ../.env.example .env          # then edit values to match your local Postgres
go run ./cmd
```

On startup the backend runs migrations and, because `APP_ENV=development`, seeds
the demo data. The API listens on `APP_PORT` (default 8080).

### Frontend

```bash
cd frontend
npm install
cp .env.example .env.local
npm run dev
```

The dev server runs on http://localhost:3000.

## Environment variables

Every variable is documented inline in the example files. Copy them and adjust
as needed.

- Backend: [`.env.example`](.env.example) — app, database, and JWT settings.
- Frontend: [`frontend/.env.example`](frontend/.env.example) — the API base URL.

The two JWT secrets are required: the backend refuses to start if either is
empty. Generate strong values with `openssl rand -hex 32`.

## Project structure

```
.
├── docker-compose.yml          # db + backend + frontend, one command
├── .env.example                # backend env template (documented)
├── backend/
│   ├── cmd/main.go             # entrypoint: config, migrate, seed, serve
│   ├── migrations/             # golang-migrate SQL files, run on startup
│   ├── docs/                   # generated OpenAPI/Swagger spec
│   └── internal/
│       ├── bootstrap/          # Fiber app, routes, middleware wiring
│       ├── config/             # env loading (Viper) and validation
│       ├── database/           # connection pool and migration runner
│       ├── handler/            # HTTP layer: parse, call service, respond
│       ├── service/            # business logic and input validation
│       ├── repository/         # SQL queries, the only layer touching the DB
│       ├── middleware/         # JWT auth guard
│       ├── model/              # domain types and request/response shapes
│       ├── notifier/           # in-memory SSE hub
│       ├── response/           # shared { data, message, error } envelope
│       ├── seeder/             # idempotent demo-data seeding
│       └── logger/             # structured logging setup
└── frontend/
    ├── app/                    # App Router: (auth) and (dashboard) route groups
    ├── components/             # UI, kanban, tasks, projects, forms
    ├── hooks/                  # TanStack Query hooks per domain
    ├── lib/                    # API client, auth context, lexorank, utils
    └── middleware.ts           # route protection
```

The backend follows a strict **handler → service → repository** layering.
Handlers never touch the database; repositories never contain business rules.
This keeps the business logic unit-testable in isolation (services are tested
against mock repositories) and makes the data-access boundary explicit.

The `cmd` folder is an entry point for application binary. An application can have different binaries: web, cli, etc. A subfolder inside `cmd` correspond to its binary type.

The `internal` directory is a special directory where go compiler won't expose its content to external packages: meaning our application business logic is isolated to use within this project and external project cannot import from `internal`. This is good practice to house application specific business logic.

## Architecture overview

### Backend

- **Layered design.** Each request flows handler → service → repository.
  Dependencies point inward and are wired at startup with Google Wire
  (compile-time dependency injection, no runtime reflection).
- **Persistence.** PostgreSQL accessed through `database/sql` with the `pgx`
  driver — explicit SQL, no ORM. Schema is managed by `golang-migrate`;
  migrations run automatically on startup so a fresh database is always at the
  latest version.
- **Auth.** Stateless JWT access tokens (15 minutes) paired with longer-lived
  refresh tokens (7 days) persisted in the database so they can be revoked on
  logout.
- **Ordering.** Kanban card order uses the Lexorank algorithm: each card holds a
  lexicographic rank string, so reordering only rewrites the moved card instead
  of renumbering the whole column.
- **Notifications.** Created notifications are persisted, then pushed to
  connected clients over Server-Sent Events via an in-memory hub. A background
  scheduler emits deadline reminders (3 days and 1 day before a due date).

### Frontend

- **Next.js App Router** with route groups: `(auth)` for sign-in/up and
  `(dashboard)` for the authenticated app.
- **TanStack Query** owns all server state — caching, request deduplication, and
  optimistic updates (notably for drag-and-drop, so cards move instantly and
  reconcile with the server response).
- **@dnd-kit** powers the Kanban board; **Tailwind** with a shadcn-style
  component layer handles styling.

## Design decisions

**Why Fiber (over net/http, Gin, or Echo)?**

Fiber is built on `fasthttp`, giving
low per-request overhead, and its Express-like API keeps routing, grouping, and
middleware concise. Its built-in middleware (CORS, rate limiting, request
logging) covered the cross-cutting needs here without extra dependencies.

**Why pgx with database/sql, not an ORM?**

The data model is small and the queries are well understood, so explicit SQL is clearer and faster than an ORM's
generated queries, and it avoids hidden N+1 behavior. `pgx` is the most actively
maintained, highest-performance PostgreSQL driver for Go. All queries are
parameterized, which also prevents SQL injection.

**Why golang-migrate?**

Versioned, reversible migrations make the schema
reproducible and reviewable. Running them on startup is what lets
`docker compose up` work with zero manual steps.

**Why Google Wire for DI?**

Wiring is resolved at compile time, so the dependency
graph is explicit and there is no runtime reflection or service-locator magic.
Missing dependencies are build errors, not runtime panics.

**Why Viper for configuration?**

It reads from environment variables (and an
optional `.env` for local dev) with sensible defaults, keeping all configuration
external to the binary — no secrets compiled in.

**Why JWT access + refresh tokens?**

Short-lived access tokens keep the common
path stateless and fast; persisting refresh tokens allows real logout and
revocation, which a pure stateless scheme cannot offer.

**Why Lexorank for ordering?**

Reordering a drag-and-drop list by integer
position forces rewriting many rows. Lexorank assigns a rank string between two
neighbors, so a move updates a single row and rarely needs reindexing.

**Why Server-Sent Events for notifications?**

Notifications are one-directional
(server to client). SSE runs over plain HTTP, reconnects automatically, and is
far simpler to operate than WebSockets for this use case. The persisted list
remains the source of truth; the stream just merges in live items.

**Why soft deletes?**

Projects and tasks are soft-deleted (`deleted_at`) and
surfaced in a trash view, so deletions are recoverable rather than destructive.

## Security

- Passwords are hashed with **bcrypt**; plaintext passwords are never stored.
- Secrets are supplied only through **environment variables**. The server fails
  fast at startup if a JWT secret is missing.
- API inputs are **validated** in the service layer (for example: name length,
  email format, password length, enum values); invalid input returns HTTP 400.
- **Rate limiting** is applied per IP: 100 requests/minute across the API and a
  stricter 10 requests/minute on auth endpoints to slow brute-force attempts.
- **CORS** is restricted to the configured frontend origin.
- All SQL is **parameterized**.

Note on the Compose secrets: `docker-compose.yml` ships development-only JWT
secrets so the stack runs with zero configuration. They are clearly marked as
such and can be overridden by exporting `JWT_ACCESS_SECRET` and
`JWT_REFRESH_SECRET` before `docker compose up`. Use strong, unique values in any
real deployment.

## Testing

Backend unit tests cover the service layer using mock repositories:

```bash
cd backend
go test ./...
```

An ephemeral PostgreSQL instance for integration-style tests is available via a
Compose profile (exposed on host port 5434):

```bash
docker compose --profile test up db-test
```
