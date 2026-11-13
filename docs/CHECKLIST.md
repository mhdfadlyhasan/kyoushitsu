# Build Checklist — v1

Tracks the build order in [DESIGN.md](DESIGN.md). The design doc is the source of
truth; this file is for tracking and verification only. Tick boxes as work lands.

Roles: `super_admin`, `admin`, `teacher`, `student`.
Config: `ADDR`, `DB_PATH`, `JWT_SECRET`, `SUPERADMIN_EMAIL`, `SUPERADMIN_PASSWORD`.

## 1. Auth + roles + super-admin seed

**Goal:** users log in, receive a JWT, and every endpoint is role-gated; exactly
one super admin is seeded on first boot.
**Depends:** none.

Tasks:
- [x] Add `SUPERADMIN_EMAIL` / `SUPERADMIN_PASSWORD` to config.
- [x] `users` table + migration with role check (`super_admin|admin|teacher|student`).
- [x] bcrypt hash/verify helper.
- [x] JWT issue/verify + auth middleware.
- [ ] Role guard middleware usable on routes.
- [ ] `POST /api/auth/login`, `GET /api/me`.
- [ ] Seed super admin on boot only if no super admin exists.
- [ ] Frontend login screen, token persistence, role-aware shell.

Acceptance:
- [ ] Valid credentials → `{token, user}`; invalid → 401.
- [ ] `/me` without/invalid token → 401; with token → identity + role.
- [ ] No public registration endpoint exists.
- [ ] Restarting does not duplicate the super admin.
- [ ] Protected routes reject missing role.

Verify: `cd backend && go build ./... && go vet ./...`; `cd frontend && npm run build && npm run lint`; manual login + `/me` + restart.

## 2. Admin user management

**Goal:** admins create and manage teachers/students; only the super admin
creates admins.
**Depends:** 1.

Tasks:
- [ ] Store: list, create, get, update, reset password, delete users.
- [ ] Handlers for `/api/admin/users` (GET, POST, PATCH, DELETE) + password reset.
- [ ] Enforce role rules: admin → teacher/student only; super admin → admins too.
- [ ] Reject deleting or demoting the super admin.
- [ ] Reject duplicate email; always hash passwords.
- [ ] Frontend Users page: list, create, edit name/role, reset password, delete.
- [ ] Hide Users nav from teacher/student.

Acceptance:
- [ ] Admin can CRUD teacher/student accounts.
- [ ] Admin cannot create/promote admins or modify the super admin → 403.
- [ ] Super admin can create admin accounts and any other user.
- [ ] Super admin cannot be deleted.
- [ ] Duplicate email → error (4xx).
- [ ] Teacher/student hitting `/api/admin/*` → 403.

Verify: role-gate requests as each role; build/vet/lint.

## 3. Class + join code

**Goal:** teachers create classes with a join code; students join with it.
**Depends:** 1.

Tasks:
- [ ] `classes` and `class_members` tables + migration.
- [ ] Store methods for create/list/get/join/members.
- [ ] Unique join code generation with retry.
- [ ] Handlers: `POST/GET /api/classes`, `POST /api/classes/join`, `GET /api/classes/:id`.
- [ ] Frontend teacher dashboard: create/list classes.
- [ ] Frontend student dashboard: join by code, open classes.
- [ ] Class detail shows join code and members (teacher only for members).

Acceptance:
- [ ] Teacher creates a class and gets a unique join code.
- [ ] Student joins by code and the class appears in their list.
- [ ] Invalid code → 4xx.
- [ ] Student cannot create a class; teacher/student lists are role-scoped.
- [ ] Re-joining does not create duplicate membership.
- [ ] Non-member cannot read class details or members.

Verify: create/join with two accounts; build/vet/lint.

## 4. Session lifecycle

**Goal:** teacher creates a scheduled session, starts it manually, and ends it.
**Depends:** 3.

Tasks:
- [ ] `sessions` table + migration (`status`, `created_at`, `started_at`, `ended_at`).
- [ ] Handlers: create (`scheduled`), `POST /start` (`scheduled` → `live`),
      `POST /end` (`live` → `ended`), `GET /api/sessions/:id`.
- [ ] Verify class ownership on every transition.
- [ ] Frontend class session controls and status display.

Acceptance:
- [ ] Create → `scheduled`; start → `live`; end → `ended`.
- [ ] Invalid transitions rejected (e.g. start an ended session).
- [ ] Non-owner teacher → 403; student cannot start/end.
- [ ] `GET` returns status and timestamps.
- [ ] Students only see sessions for classes they joined.

Verify: transition tests by role; build/vet/lint.

## 5. Teacher board + text notes

**Goal:** teacher presents locally and sends text notes that persist and push live.
**Depends:** 4.

Tasks:
- [ ] Session-scoped WebSocket hub with token auth and role registration.
- [ ] `notes` table + migration (`body` text).
- [ ] `POST /api/sessions/:id/notes {body}` stores + broadcasts; `GET` lists.
- [ ] Teacher board (pen, eraser, clear) — local only, never sent or stored.
- [ ] Teacher note composer in the session screen.
- [ ] Student session screen: fetch note history + subscribe to live notes.

Acceptance:
- [ ] Note is persisted; a late joiner receives history via `GET`.
- [ ] Live students receive the note immediately over WebSocket.
- [ ] Teacher board content is never transmitted or stored.
- [ ] Empty note rejected.
- [ ] Only the session's teacher can send notes.
- [ ] Non-member cannot read notes.

Verify: two browsers (teacher + student); reconnect and refetch; build/vet/lint.

## 6. Student board + ephemeral submission

**Goal:** student submits a board snapshot; teacher sees it live and nothing is stored.
**Depends:** 5.

Tasks:
- [ ] `POST /api/sessions/:id/submissions {image}` validates PNG data URL and
      relays to the teacher's session sockets; no database write.
- [ ] WebSocket `submission` message to the teacher.
- [ ] Student board + Submit action.
- [ ] Teacher live submissions feed (ephemeral).

Acceptance:
- [ ] Student submit → teacher receives it over WebSocket; no DB row.
- [ ] Non-member student → 403; teacher submit → 403.
- [ ] Invalid/oversized image rejected.
- [ ] Teacher offline at submit → submission is lost (documented best-effort).

Verify: submit from phone + PC; confirm no table/rows; build/vet/lint.

## 7. Polish

**Goal:** usable on PC and phone with clear states.
**Depends:** 1–6.

Tasks:
- [ ] Notes history and empty states.
- [ ] Error handling and loading/disabled states.
- [ ] Mobile/responsive pass (touch drawing, submit).
- [ ] README: env vars, run instructions, super-admin seeding.

Acceptance:
- [ ] Full teacher/student flow works on desktop and phone.
- [ ] No console errors; lint/build pass.
- [ ] All previous acceptance criteria still pass.

## Cross-cutting

- [ ] Tests cover role gates and session transitions.
- [ ] CI runs backend build/vet/test and frontend build/lint.
