# Study Companion

A study companion for teachers and students.

See **[docs/DESIGN.md](docs/DESIGN.md)** for the design doc (source of truth).

## Stack

- **Backend:** Go (`chi`, `gorilla/websocket`, `golang-jwt`, `pgx`, PostgreSQL)
- **Frontend:** React + Vite + TypeScript (plain `<canvas>` for drawing)

## Layout

```
backend/        Go API + WebSocket server
  cmd/server/   entrypoint
  internal/     config, store, handler, auth, ws
frontend/       React + Vite + TS
  src/          app code (pages/, components/, lib/)
docs/           DESIGN.md and other docs
```

## Prerequisites

- Go 1.26+
- Node 20+ and npm
- PostgreSQL 16+ (or Docker for the bundled `docker-compose.yml`)

## Run the backend

Start PostgreSQL (creates the `study_companion` database):

```sh
docker compose up -d
```

Then run the API:

```sh
cd backend
go run ./cmd/server
# listens on :8080, connects to DATABASE_URL
```

Create the super admin (once, before first login):

```sh
cd backend
SUPERADMIN_EMAIL=owner@example.com SUPERADMIN_PASSWORD=change-me \
  go run ./cmd/server seed-superadmin
```

The command is idempotent: it does nothing if a super admin already exists.
Normal server boot never writes users.

Config via env vars (defaults apply):

| Var | Default | Purpose |
|---|---|---|
| `ADDR` | `:8080` | HTTP listen address |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/study_companion?sslmode=disable` | PostgreSQL connection string |
| `JWT_SECRET` | `dev-secret-change-me` | JWT signing secret (change in prod) |
| `SUPERADMIN_EMAIL` | — | Super admin email used by `seed-superadmin` |
| `SUPERADMIN_PASSWORD` | — | Super admin password (hashed at seed time) |

Health check: `curl http://localhost:8080/healthz` → `ok`

## Run the frontend

```sh
cd frontend
npm install
npm run dev
# serves on http://localhost:5173
```

The Vite dev server proxies `/api` (including WebSockets) to `http://localhost:8080`.

## Scripts

```sh
# backend
cd backend && go build ./... && go vet ./...

# backend tests (database-backed tests are skipped without TEST_DATABASE_URL)
cd backend && TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/study_companion?sslmode=disable' go test ./...

# frontend
cd frontend && npm run build && npm run lint
```
