# Study Companion — Design Doc (v1 draft)

> Status: DRAFT — review and edit before writing code. This document is the source
> of truth for what we build. Features are intentionally small; see "Out of scope".

## 1. Problem & audience

A lightweight study companion for a teacher and their students.

- **Teacher** needs a live digital board to present from, a way to push class notes
  to students, and a way to collect students' written work.
- **Student** needs to join a class, follow along, receive notes, and hand in
  written answers from PC or phone.

One teacher, many students, short-lived class sessions. Not a full LMS.

## 2. Glossary

- **Class** — a durable group created by a teacher. Has a join code.
- **Session** — a live, time-boxed instance of a class. Teacher starts/ends it.
- **Note** — a snapshot the teacher sends to students during a session.
- **Submission** — a snapshot a student hands in during a session (non-realtime).
- **Board** — a drawing canvas. Teacher board is *local only*; student board is
  used to compose a submission.

## 3. User flows

### 1. Login / register
- User registers with name, email, password, role (teacher | student).
- User logs in with email + password, gets a JWT.
- `/me` returns identity + role. Role gates all UI and API access.

### 2. Teacher creates a class
- Teacher clicks "Create class", enters a name.
- System generates a short join code (e.g. `XK4-9TQ`).
- Teacher sees the class with its join code and member list.

### 3. Student joins a class
- Student enters a join code.
- System adds them to `class_members`. Class appears in their list.

### 4. Teacher starts a session
- Teacher opens a class, clicks "Start session".
- System creates a session (status `live`), so students can join/watch/submit.
- Teacher clicks "End session" to close it (status `ended`).

### 5. Teacher presents on the board (local, nothing sent)
- Teacher draws on a blank canvas. **Nothing is saved or sent** — it is a
  presentation surface shared via external screen share (Zoom/Meet).
- Tools: pen, eraser, clear. Color/size TBD.

### 6. Teacher sends a class note
- Teacher, optionally with the board showing, clicks "Send note".
- System captures the current board as a PNG and pushes it to all students in the
  live session over WebSocket. It is also stored so late joiners / history can see it.

### 7. Student submits written work
- Student opens the "Board" screen, draws their answer (mouse or touch).
- Student clicks "Submit". System POSTs the PNG snapshot (non-realtime).
- Teacher sees a list of submissions per session.

## 4. Screen sketches (text wireframes)

```
Login                    Dashboard (teacher)        Session (teacher)
+------------------+     +--------------------+     +----------------------+
|  email           |     | + New class        |     | [board canvas]       |
|  password        |     | Class A  [Start]   |     |  pen | eraser | clear|
|  [Login] [Reg]   |     | Class B  [Start]   |     | [Send note] [End]    |
+------------------+     +--------------------+     +----------------------+
                                                    | submissions (list)   |

Dashboard (student)      Session (student)
+--------------------+   +----------------------+
| enter join code    |   |  notes received      |
| Class A  [Join]    |   |  [board canvas]      |
|  Class A  [Open]   |   |  pen | eraser | clear|
+--------------------+   |  [Submit]            |
                         +----------------------+
```

## 5. Data model (SQLite)

| Table | Columns |
|---|---|
| `users` | id, name, email (unique), password_hash, role, created_at |
| `classes` | id, name, teacher_id, join_code (unique), created_at |
| `class_members` | id, class_id, user_id, joined_at |
| `sessions` | id, class_id, status (`live`\|`ended`), started_at, ended_at |
| `notes` | id, session_id, teacher_id, image (PNG base64/blob), created_at |
| `submissions` | id, session_id, student_id, image (PNG base64/blob), created_at |

Notes:
- Whiteboard content = PNG snapshot (`canvas.toDataURL()`). Simple, works for
  PC + phone. Vector format (SVG) is a later option.
- Images stored as base64 TEXT for MVP; move to files/BLOB if size becomes a problem.

## 6. API + WebSocket

### Auth (JWT in `Authorization: Bearer`)
- `POST /api/auth/register` `{name,email,password,role}`
- `POST /api/auth/login` `{email,password}` → `{token, user}`
- `GET  /api/me` → current user

### Classes
- `POST /api/classes` (teacher) `{name}` → `{class}` (with join_code)
- `GET  /api/classes` → list for current user (teacher: owned; student: joined)
- `POST /api/classes/join` (student) `{code}`
- `GET  /api/classes/:id` → class + members (members visible to teacher)

### Sessions
- `POST /api/classes/:id/sessions` (teacher) → start
- `POST /api/sessions/:id/end` (teacher)
- `GET  /api/sessions/:id` → session + status

### Notes (teacher → students)
- `POST /api/sessions/:id/notes` (teacher) `{image}` → stores + broadcasts
- `GET  /api/sessions/:id/notes` → list

### Submissions (student → teacher)
- `POST /api/sessions/:id/submissions` (student) `{image}`
- `GET  /api/sessions/:id/submissions` (teacher) → list

### WebSocket
- `GET /api/sessions/:id/ws` (upgrade, token via query/header)
- Server → students: `{"type":"note", "note": {...}}` (live push of a sent note)
- Server → teacher: `{"type":"student_joined"}` (optional, TBD)

## 7. Out of scope (v1)

- Live streaming of the teacher board (presented via external screen share).
- Students taking their own private notes.
- Realtime student submissions / live ink (submissions are explicit submit only).
- Vector/redrawable notes (PNG snapshots only).
- Email verification, password reset, OAuth.
- Chat, attendance tracking, grades.
- Multi-tenant / LMS features, scheduling, recurring classes.

## 8. Tech choices & risks

| Concern | Choice | Risk / mitigation |
|---|---|---|
| Backend | Go, `chi`, `gorilla/websocket`, `golang-jwt`, `modernc.org/sqlite`, `bcrypt` | SQLite single-file is fine for MVP; swap to Postgres via an interface later. |
| Frontend | React + Vite + TS, plain `<canvas>` + Pointer Events | Touch support handled by Pointer Events (no extra lib). |
| Auth | JWT (short-lived) + bcrypt | Rotate secret via env; add refresh later if needed. |
| Realtime | WebSocket only for teacher→student note push | Single-direction, low complexity. |

Risks:
1. **Auth correctness** — gate every endpoint by role; test teacher-only vs student-only.
2. **Canvas on phone** — verify touch drawing and PNG capture on mobile early.
3. **WS auth & reconnect** — authenticate the socket and handle drop/reconnect.

## 9. Build order

1. Auth (JWT + roles)
2. Class + join code
3. Session lifecycle
4. Teacher board (local) + send note (WS push)
5. Student board + submit
6. Polish (history, empty states, errors)
