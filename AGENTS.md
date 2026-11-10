# AGENTS.md

Instructions for agents working in this repository.

## Design is the source of truth

Before changing any code, read **[docs/DESIGN.md](docs/DESIGN.md)**.

- Implement only what the design doc specifies. Do not invent features, endpoints,
  tables, or screens that are not described there.
- If a request conflicts with the design, stop and surface the conflict; update
  `docs/DESIGN.md` first (or ask), then change code.
- Keep code, API shapes, data model, and terminology aligned with the doc
  (Class, Session, Note, Submission, Board). Use the same names in code.
- Respect the "Out of scope (v1)" section. Do not add live ink streaming, vector
  notes, chat, grades, etc.
- Build in the order given in "Build order" unless told otherwise.

## Project layout

```
backend/        Go API + WebSocket server
  cmd/server/   entrypoint
  internal/     config, store, handler, auth, ws
frontend/       React + Vite + TypeScript
  src/          app code (App.tsx, components/, assets/)
docs/           DESIGN.md (source of truth)
```

Stack: Go (`chi`, `gorilla/websocket`, `golang-jwt`, `modernc.org/sqlite`,
`bcrypt`) and React 19 + Vite + TS. Board drawing uses plain `<canvas>` +
Pointer Events, with `perfect-freehand` for ink.

## Commands

```sh
# backend
cd backend && go build ./... && go vet ./...
cd backend && go run ./cmd/server

# frontend
cd frontend && npm install
cd frontend && npm run build && npm run lint
cd frontend && npm run dev
```

Run the relevant build/lint/typecheck before finishing. Backend config comes from
`ADDR`, `DB_PATH`, `JWT_SECRET` (see README).

## Conventions

- No comments unless they explain non-obvious intent (existing code uses short
  package/doc comments).
- Backend: `internal/<feature>` packages mounted from `internal/handler/router.go`.
  Keep handlers thin; SQL lives in `internal/store`.
- Frontend: function components, TypeScript, co-located CSS. Prefer existing
  patterns in `src/components/HandwritingBoard.tsx`.
- Gate every endpoint by auth + role (teacher vs student) per the design.

## Checklist before finishing

1. Change matches `docs/DESIGN.md` (update the doc first if it does not).
2. Build/vet/lint pass.
3. No new scope from the "Out of scope" list.
