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

Config via env vars (all optional):

| Var | Default | Purpose |
|---|---|---|
| `ADDR` | `:8080` | HTTP listen address |
| `DB_PATH` | `study-companion.db` | SQLite file path |
| `JWT_SECRET` | `dev-secret-change-me` | JWT signing secret (change in prod) |

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
