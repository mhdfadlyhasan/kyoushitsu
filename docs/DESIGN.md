# Study Companion — Design Doc (v1 draft)

> Status: DRAFT — review and edit before writing code. This document is the source
> of truth for what we build. Features are intentionally small; see "Out of scope".

## 1. Problem & audience

A lightweight study companion for a teacher and their students.

- **Super admin** — the single owner account (bootstrapped from env). Creates
  admins and, when needed, any user. Only one exists in v1.
- **Admin** — creates and manages teacher and student accounts, assigns roles,
  and resets passwords. Cannot create other admins.
- **Teacher** — needs a live digital board to present from, a way to send text
  notes to students in every session, and a way to receive students' written work.
- **Student** — needs to join classes, follow along, receive notes, and hand in
  written answers from PC or phone.

One teacher can own many classes; a student can join many classes. Each class has
short-lived sessions. Not a full LMS. Users are created by the super admin or
admins; there is no public self-registration.

## 2. Glossary

- **Class** — a durable group created by a teacher. Has a join code.
- **Session** — a time-boxed instance of a class. Created `scheduled`, started and
  ended manually by the teacher.
- **Note** — a text-only message the teacher sends to students during a session.
  Persisted per session so late joiners / history can see it.
- **Submission** — a PNG snapshot of a student's board, handed in during a
  session. Relayed live to the teacher; not persisted.
- **Board** — a drawing canvas. Teacher board is *local only*; student board is
  used to compose a submission.
- **Admin** — an account that creates and manages teacher and student accounts.
- **Super admin** — the single owner account that creates admins and all users.

## 3. User flows

### 1. Login / bootstrap
- On first boot the server seeds one super admin from `SUPERADMIN_EMAIL` and
  `SUPERADMIN_PASSWORD` if no super admin exists.
- There is no public registration. All other users are created by admins.
- User logs in with email + password, gets a JWT.
- `/me` returns identity + role. Role gates all UI and API access.

### 2. Admin manages users
- Admin opens "Users", creates a teacher or student with name, email, initial
  password, and role.
- Admin edits a user's name/role, resets their password, or deletes them.
- Super admin can do the same for any account, including creating admins. Only
  the super admin can create or remove admins.

### 3. Teacher creates a class
- Teacher clicks "Create class", enters a name.
- System generates a short join code (e.g. `XK4-9TQ`).
- Teacher sees the class with its join code and member list.

### 4. Student joins a class
- Student enters a join code.
- System adds them to `class_members`. Class appears in their list.

### 5. Teacher runs a session
- Teacher opens a class, clicks "Create session" (status `scheduled`).
- Teacher clicks "Start session" to make it `live`, so students can
  join/watch/submit. Starting is manual in v1.
- Teacher clicks "End session" to close it (status `ended`).

### 6. Teacher presents on the board (local, nothing sent)
- Teacher draws on a blank canvas. **Nothing is saved or sent** — it is a
  presentation surface shared via external screen share (Zoom/Meet).
- Tools: pen, eraser, clear. Color/size TBD.

### 7. Teacher sends a class note
- In a live session, the teacher types a note and clicks "Send note".
- System stores the text and pushes it to all students in the live session over
  WebSocket. It is persisted per session so late joiners / history can see it.

### 8. Student submits written work in board
- Student opens the "Board" screen, draws their answer (mouse or touch).
- Student clicks "Submit". System POSTs the PNG snapshot.
- The server relays it live to the teacher's WebSocket for that session; it is
  not stored. The teacher sees a live submissions feed.

## 4. Screen sketches (text wireframes)

```
Login                    Dashboard (teacher)        Session (teacher)
+------------------+     +--------------------+     +----------------------+
|  email           |     | + New class        |     | [board canvas]       |
|  password        |     | Class A  [Start]   |     |  pen | eraser | clear|
|  [Login]         |     | Class B  [Start]   |     +----------------------+
+------------------+     +--------------------+     | note: [__________]   |
                                                    | [Send note]  [End]   |
                                                    | submissions (live)   |
                                                    +----------------------+

Dashboard (student)      Session (student)
+--------------------+   +----------------------+
| enter join code    |   |  notes received      |
| Class A  [Join]    |   |  - note text ...     |
|  Class A  [Open]   |   |  [board canvas]      |
+--------------------+   |  pen | eraser | clear|
                         |  [Submit]            |
                         +----------------------+

Admin — Users
+--------------------------------------------------+
| + New user                                       |
| Name        Email           Role      Actions    |
| Alice       a@example.com   teacher   [Edit]     |
| Bob         b@example.com   student   [Edit]     |
+--------------------------------------------------+
```

## 5. Data model (SQLite)

| Table           | Columns                                                                                              |
| --------------- | ---------------------------------------------------------------------------------------------------- |
| `users`         | id, name, email (unique), password_hash, role (`super_admin`\|`admin`\|`teacher`\|`student`), created_at |
| `classes`       | id, name, teacher_id, join_code (unique), created_at                                                 |
| `class_members` | id, class_id, user_id, joined_at                                                                     |
| `sessions`      | id, class_id, status (`scheduled`\|`live`\|`ended`), created_at, started_at, ended_at                |
| `notes`         | id, session_id, teacher_id, body (text), created_at                                                  |

Notes:
- Notes are text-only and persisted per session; late joiners and history read them.
- Student submissions are PNG snapshots (`canvas.toDataURL()`) relayed live to the
  teacher over WebSocket and **not stored** — there is no `submissions` table.
- One super admin is seeded from env on first boot; all other users are created by
  admins or the super admin.

## 6. API + WebSocket

### Auth (JWT in `Authorization: Bearer`)
- `POST /api/auth/login` `{email,password}` → `{token, user}`
- `GET  /api/me` → current user
- No public registration. The single super admin is seeded from
  `SUPERADMIN_EMAIL` / `SUPERADMIN_PASSWORD` on first boot.

### Admin (super admin + admin)
- `GET    /api/admin/users` → list users
- `POST   /api/admin/users` `{name,email,password,role}` → create user
- `PATCH  /api/admin/users/:id` `{name?,role?}` → update user
- `POST   /api/admin/users/:id/password` `{password}` → reset password
- `DELETE /api/admin/users/:id` → delete user
- Admins may only create/manage `teacher` and `student` accounts. Only the super
  admin can create or manage `admin` accounts, and the super admin cannot be deleted.

### Classes
- `POST /api/classes` (teacher) `{name}` → `{class}` (with join_code)
- `GET  /api/classes` → list for current user (teacher: owned; student: joined)
- `POST /api/classes/join` (student) `{code}`
- `GET  /api/classes/:id` → class + members (members visible to teacher)

### Sessions
- `POST /api/classes/:id/sessions` (teacher) → create (status `scheduled`)
- `POST /api/sessions/:id/start` (teacher) → `scheduled` → `live`
- `POST /api/sessions/:id/end` (teacher) → `live` → `ended`
- `GET  /api/sessions/:id` → session + status

### Notes (teacher → students)
- `POST /api/sessions/:id/notes` (teacher) `{body}` → stores + broadcasts
- `GET  /api/sessions/:id/notes` → list

### Submissions (student → teacher, ephemeral)
- `POST /api/sessions/:id/submissions` (student) `{image}`
  → relayed to the teacher's WebSocket for the session; not stored

### WebSocket
- `GET /api/sessions/:id/ws` (upgrade, token via query/header)
- Server → students: `{"type":"note","note":{"id","session_id","body","created_at"}}`
- Server → teacher: `{"type":"submission","submission":{"student_id","student_name","image","created_at"}}`
- Server → teacher: `{"type":"student_joined"}` (optional, TBD)

## 7. Out of scope (v1)

- Live streaming of the teacher board (presented via external screen share).
- Students taking their own private notes.
- Live ink streaming (submissions are explicit submit only; they are relayed live
  but not stored).
- Storing submissions / submission history.
- Automatic session scheduling (sessions are created and started manually).
- Public self-registration.
- Vector/redrawable notes (notes are text-only).
- Email verification, self-service password reset, OAuth.
- Chat, attendance tracking, grades.
- Multi-tenant / LMS features, recurring classes.
- Fine-grained RBAC (v1 uses basic role gates only).

## 8. Tech choices & risks

| Concern  | Choice                                                                       | Risk / mitigation                                                            |
| -------- | ---------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Backend  | Go, `chi`, `gorilla/websocket`, `golang-jwt`, `modernc.org/sqlite`, `bcrypt` | SQLite single-file is fine for MVP; swap to Postgres via an interface later. |
| Frontend | React + Vite + TS, plain `<canvas>` + Pointer Events                         | Touch support handled by Pointer Events (no extra lib).                      |
| Auth     | JWT (short-lived) + bcrypt                                                   | Rotate secret via env; add refresh later if needed.                          |
| Realtime | WebSocket for note push (teacher→students) and submission relay (student→teacher) | Two directions, both session-scoped; submissions are best-effort. |

Risks:
1. **Auth correctness** — gate every endpoint by role; test super-admin/admin/teacher/student.
2. **Canvas on phone** — verify touch drawing and PNG capture on mobile early.
3. **WS auth & reconnect** — authenticate the socket and handle drop/reconnect.
4. **Ephemeral submissions** — if the teacher is offline or disconnected, the
   submission is lost; keep the live feed clearly best-effort.
5. **Super admin seeding** — read credentials from env and never commit a real
   password; seed only when no super admin exists.

## 9. Build order

1. Auth (JWT + roles) + super-admin seed
2. Admin user management
3. Class + join code
4. Session lifecycle (`scheduled` → `live` → `ended`)
5. Teacher board (local) + text notes (WS push)
6. Student board + ephemeral submission relay
7. Polish (history, empty states, errors)

## 10. What's next (v2)?
Planning feature + heavy ai usage

1. Grades
2. Quiz feature, quiz maker, quiz answer, ai generated quiz based on teacher input
3. Student performance
4. note type (capability of integration to anki)
5. postgres database

