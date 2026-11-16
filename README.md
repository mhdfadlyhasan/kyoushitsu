# Study Companion

A study companion for teachers and students.

See **[docs/DESIGN.md](docs/DESIGN.md)** for the design doc (source of truth).

## Stack

- **Backend:** Go (`chi`, `gorilla/websocket`, `golang-jwt`, `modernc.org/sqlite`)
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

## Run the backend

```sh
cd backend
go run ./cmd/server
# listens on :8080, creates study-companion.db in backend/
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
| `DB_PATH` | `study-companion.db` | SQLite file path |
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

# frontend
cd frontend && npm run build && npm run lint
```
