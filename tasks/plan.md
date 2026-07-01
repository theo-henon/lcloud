# Implementation Plan: Phase 0 — Technical foundations

> **Source spec:** [SPEC.md](../SPEC.md)
> **Branch suggestion:** `feature/phase-0-foundations`
> **Estimated tasks:** 16 vertical slices + 4 checkpoints

---

## Overview

Bootstrap lcloud from documentation-only to a running Docker stack: Go API with JWT auth (access 24h + refresh 7d), PostgreSQL with User/Volume/RefreshToken models, React SPA with design system, login flow, and a grouped sidebar skeleton. Single Alpine image serves API + embedded frontend on port 8080.

---

## Architecture decisions (from spec)

| Decision | Choice |
|---|---|
| Access token | JWT Bearer, 24h (`JWT_EXPIRY_HOURS`) |
| Refresh token | Opaque, hashed in PostgreSQL, 7d, rotated on refresh |
| Admin bootstrap | Seed from `ADMIN_EMAIL` / `ADMIN_PASSWORD` if DB empty |
| Volume model | GORM schema only — no API |
| Frontend delivery | Embedded `web/dist` served by Gin |
| Docker runtime | Alpine (Phase 0); Distroless deferred to pre-MVP ADR |
| Sidebar | Grouped nav + "Soon" badges on future MVP routes |

---

## Dependency graph

```
Task 1 (scaffold)
    └── Task 2 (config + httputil)
            └── Task 3 (Docker stub + health)
                    └── Task 4 (GORM models)
                            └── Task 5 (auth service)
                                    └── Task 6 (handlers + router)
                                            └── Task 7 (admin seed + create user)
                                                    └── Task 8 (auth tests)
                                                            │
                    Task 9 (Vite scaffold) ─────────────────┤  ← can start after Task 3
                            └── Task 10 (Tailwind + shadcn) │
                                    └── Task 11 (API client + store) ← needs Task 6 API contract
                                            └── Task 12 (login + guards)
                                                    └── Task 13 (shell + sidebar + placeholders)
                                                            └── Task 14 (static embed)
                                                                    └── Task 15 (frontend tests)
                                                                            └── Task 16 (E2E + README)
```

---

## Task list

### Phase 1: Foundation

---

## Task 1: Repository scaffold

**Description:** Initialize the Go module, directory tree, gitignore, and environment template. No runtime logic yet — just the skeleton the rest of Phase 0 builds on.

**Acceptance criteria:**
- [ ] `go.mod` exists with module path `github.com/<org>/lcloud` (or chosen path)
- [ ] Directories created per [SPEC.md](../SPEC.md#project-structure): `cmd/server/`, `internal/{auth,volume,api,config}/`, `pkg/httputil/`, `web/`, `plugins/`, `docs/adr/`
- [ ] `.gitignore` covers `.env`, `web/node_modules/`, `web/dist/`, `plugins/*.`, build artifacts
- [ ] `.env.example` lists all variables from spec with comments

**Verification:**
- [ ] `go mod tidy` succeeds (empty main ok)
- [ ] Manual: directory tree matches spec

**Dependencies:** None

**Files likely touched:**
- `go.mod`, `.gitignore`, `.env.example`
- `plugins/.gitkeep`, `docs/adr/.gitkeep`
- `cmd/server/main.go` (stub)

**Maps to:** SC0.11

**Estimated scope:** S (4–6 files)

---

## Task 2: Config loader and HTTP helpers

**Description:** Centralize environment loading and shared JSON/error response utilities used by all handlers.

**Acceptance criteria:**
- [ ] `internal/config/config.go` loads all env vars from spec with defaults (`JWT_EXPIRY_HOURS=24`, `REFRESH_TOKEN_EXPIRY_DAYS=7`, `APP_PORT=8080`, `GIN_MODE=release`)
- [ ] Validates required vars at startup; fails fast with clear message if missing
- [ ] `pkg/httputil/` provides `JSON`, `BadRequest`, `Unauthorized`, `Forbidden`, `InternalError` helpers matching spec error shape

**Verification:**
- [ ] Unit test: config defaults applied when env vars absent
- [ ] `go build ./...` succeeds

**Dependencies:** Task 1

**Files likely touched:**
- `internal/config/config.go`, `internal/config/config_test.go`
- `pkg/httputil/response.go`

**Estimated scope:** S (3 files)

---

## Task 3: Docker Compose and health endpoint stub

**Description:** Wire PostgreSQL 17 + app service in Docker Compose. Minimal Go server responds on `/api/health` so the stack is verifiable before auth exists.

**Acceptance criteria:**
- [ ] `docker-compose.yml`: services `app` + `postgres:17`, postgres healthcheck, `app` depends on healthy postgres
- [ ] `Dockerfile` multi-stage skeleton (web-builder, go-builder, Alpine runtime) — web stage can copy placeholder until Task 14
- [ ] `GET /api/health` returns `{ "status": "ok" }`
- [ ] Server connects to PostgreSQL (ping on startup, log success)

**Verification:**
- [ ] `docker compose up -d --build` → both services healthy
- [ ] `curl http://localhost:8080/api/health` → 200

**Dependencies:** Task 2

**Files likely touched:**
- `docker-compose.yml`, `Dockerfile`
- `cmd/server/main.go`, `internal/api/router.go` (health only)

**Maps to:** SC0.1, SC0.3

**Estimated scope:** M (5 files)

---

### Checkpoint 1: Infrastructure

- [ ] `docker compose up` works
- [ ] `/api/health` returns 200
- [ ] Postgres reachable from app
- [ ] Review before auth work

---

### Phase 2: Backend auth and data

---

## Task 4: GORM models and AutoMigrate

**Description:** Define User, RefreshToken, and Volume models; run AutoMigrate on startup.

**Acceptance criteria:**
- [ ] `User` model: UUID PK, email unique, password_hash, role enum (`admin`|`user`), timestamps
- [ ] `RefreshToken` model: UUID PK, user_id FK, token_hash, expires_at, revoked_at nullable, created_at
- [ ] `Volume` model: fields per spec (schema only)
- [ ] AutoMigrate runs in `main.go` before seed/handlers

**Verification:**
- [ ] `docker compose up` → tables `users`, `refresh_tokens`, `volumes` exist in postgres
- [ ] `\dt` in psql confirms schema

**Dependencies:** Task 3

**Files likely touched:**
- `internal/auth/model.go`, `internal/volume/model.go`
- `cmd/server/main.go`

**Maps to:** SC0.8

**Estimated scope:** S (3 files)

---

## Task 5: Auth service (login, JWT, refresh, logout)

**Description:** Core auth business logic — password hashing, access JWT issuance, opaque refresh token lifecycle with rotation.

**Acceptance criteria:**
- [ ] Bcrypt password hashing on user create/login verify
- [ ] `Login(email, password)` → access_token (JWT, claims: sub, email, role, exp) + refresh_token (opaque)
- [ ] Refresh token stored as SHA-256 hash in DB with expiry from `REFRESH_TOKEN_EXPIRY_DAYS`
- [ ] `Refresh(refreshToken)` → new access_token + rotated refresh_token (old revoked)
- [ ] `Logout(refreshToken)` → sets `revoked_at`
- [ ] `ValidateAccessToken(token)` → user claims
- [ ] Expired/revoked refresh returns sentinel error (not internal details)

**Verification:**
- [ ] `go test ./internal/auth/... -v` covers login, refresh rotation, logout revoke, expired token
- [ ] Coverage on `internal/auth/` ≥ 80%

**Dependencies:** Task 4

**Files likely touched:**
- `internal/auth/service.go`, `internal/auth/service_test.go`

**Maps to:** SC0.5, SC0.6b, SC0.6c

**Estimated scope:** M (2 files, dense logic)

---

## Task 6: Auth middleware, handlers, and router

**Description:** Thin Gin handlers wired to auth service; JWT middleware for protected routes; public auth endpoints.

**Acceptance criteria:**
- [ ] `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `GET /api/auth/me`
- [ ] JWT middleware extracts Bearer token, validates, sets user in context
- [ ] `RequireAdmin` middleware checks role
- [ ] Handlers call service only — no business logic in handlers
- [ ] Error responses match spec JSON shape

**Verification:**
- [ ] Manual curl: login → me with token → 401 without token
- [ ] Manual curl: refresh → new access token; logout → refresh fails

**Dependencies:** Task 5

**Files likely touched:**
- `internal/auth/middleware.go`
- `internal/api/auth_handler.go`, `internal/api/router.go`

**Maps to:** SC0.6, SC0.6b, SC0.6c

**Estimated scope:** M (4 files)

---

## Task 7: Admin seed and user creation endpoint

**Description:** Bootstrap first admin from env; expose admin-only user creation API.

**Acceptance criteria:**
- [ ] On startup: if `users` count = 0, create admin from `ADMIN_EMAIL` / `ADMIN_PASSWORD`
- [ ] `POST /api/admin/users` — admin only, body `{ email, password, role? }`, returns user without password
- [ ] Non-admin JWT → 403
- [ ] Duplicate email → 409 or 400 with clear error

**Verification:**
- [ ] Fresh DB → admin login works with env credentials (SC0.4)
- [ ] Admin creates user via curl; new user can login
- [ ] Regular user cannot call `/api/admin/users` (SC0.7)

**Dependencies:** Task 6

**Files likely touched:**
- `internal/api/admin_handler.go`
- `cmd/server/main.go` (seed call)

**Maps to:** SC0.4, SC0.7

**Estimated scope:** S (2–3 files)

---

## Task 8: Auth integration tests

**Description:** HTTP-level tests for the full auth API contract using test database or httptest with mocked service where needed.

**Acceptance criteria:**
- [ ] Tests cover: login success/failure, me authorized/unauthorized, refresh valid/expired/revoked, logout, admin create user, admin forbidden for user role
- [ ] `go test ./...` passes

**Verification:**
- [ ] `go test ./... -cover` — auth package ≥ 80%

**Dependencies:** Task 7

**Files likely touched:**
- `internal/api/auth_handler_test.go`, `internal/api/admin_handler_test.go`

**Maps to:** SC0.12

**Estimated scope:** M (2 files)

---

### Checkpoint 2: Backend complete

- [ ] All auth endpoints work via curl
- [ ] Admin seed on fresh DB
- [ ] `go test ./...` green, auth coverage ≥ 80%
- [ ] Review before frontend integration

---

### Phase 3: Frontend scaffold and auth UI

> **Note:** Tasks 9–10 can start after Checkpoint 1 in parallel with Tasks 4–8 if using two agents. Task 11 requires Task 6 (API contract stable).

---

## Task 9: Vite + React + TypeScript scaffold

**Description:** Initialize the frontend app with React Router, TanStack Query, and Zustand dependencies per [docs/project.md](../docs/project.md).

**Acceptance criteria:**
- [ ] `web/package.json` with React 19, Vite 8, TypeScript, React Router, TanStack Query, Zustand
- [ ] `vite.config.ts` proxies `/api` → `localhost:8080` in dev
- [ ] `main.tsx`, `App.tsx` render without errors
- [ ] `npm run dev` and `npm run build` succeed

**Verification:**
- [ ] `cd web && npm run build` → `web/dist/` created

**Dependencies:** Task 1 (directory exists); parallel with Task 4+

**Files likely touched:**
- `web/package.json`, `web/vite.config.ts`, `web/tsconfig.json`, `web/index.html`
- `web/src/main.tsx`, `web/src/App.tsx`

**Estimated scope:** M (6–8 files)

---

## Task 10: Tailwind design tokens and shadcn/ui

**Description:** Map [design/DESIGN.md](../design/DESIGN.md) colors/typography to Tailwind; init shadcn/ui in dark mode with Button, Input, Card.

**Acceptance criteria:**
- [ ] Tailwind theme: `canvas`, `primary`, `surface-card`, `hairline`, `body`, `muted`, etc. from DESIGN.md
- [ ] Fonts: Inter (UI), JetBrains Mono (mono)
- [ ] shadcn/ui configured; `Button` default variant = yellow CTA (`#faff69` on dark)
- [ ] Global dark background `#0a0a0a`

**Verification:**
- [ ] Manual: dev page shows yellow button on dark canvas
- [ ] `npm run build` succeeds

**Dependencies:** Task 9

**Files likely touched:**
- `web/tailwind.config.ts`, `web/src/index.css`
- `web/src/components/ui/button.tsx`, `input.tsx`, `card.tsx`
- `web/src/lib/utils.ts`

**Maps to:** SC0.2 (visual foundation)

**Estimated scope:** M (5–7 files)

---

## Task 11: API client, auth store, and silent refresh

**Description:** Centralized fetch wrapper with Bearer header; Zustand store for access token; automatic refresh on 401.

**Acceptance criteria:**
- [ ] `lib/api.ts`: base URL, JSON helpers, attaches `Authorization: Bearer`
- [ ] On 401: call `/api/auth/refresh`, update tokens, retry once
- [ ] Zustand store: `accessToken`, `refreshToken`, `user`, `login`, `logout`, `setTokens`
- [ ] Tokens persisted: access in sessionStorage; refresh in sessionStorage (Phase 0) — document httpOnly cookie as post-Phase-0 hardening
- [ ] TanStack Query provider wraps app

**Verification:**
- [ ] Unit test: store login/logout clears state
- [ ] Unit test: api client retries after mock refresh

**Dependencies:** Task 6 (API exists), Task 10

**Files likely touched:**
- `web/src/lib/api.ts`, `web/src/store/auth.ts`
- `web/src/App.tsx`

**Estimated scope:** M (3–4 files)

---

## Task 12: Login page and route guards

**Description:** Login form wired to API; protected routes redirect unauthenticated users.

**Acceptance criteria:**
- [ ] `LoginPage`: email + password, yellow submit, error display, loading state
- [ ] Success → store tokens → navigate `/dashboard`
- [ ] `ProtectedRoute` wrapper: no token → redirect `/login`
- [ ] `/` redirects to `/dashboard` when authed, `/login` when not
- [ ] Routes registered for all paths in spec (placeholders can be stub components initially)

**Verification:**
- [ ] Manual dev: unauthenticated `/dashboard` → `/login` (SC0.10)
- [ ] Manual dev: login with admin → lands on dashboard

**Dependencies:** Task 11

**Files likely touched:**
- `web/src/pages/LoginPage.tsx`
- `web/src/components/ProtectedRoute.tsx`
- `web/src/App.tsx`

**Maps to:** SC0.2, SC0.10

**Estimated scope:** M (3–4 files)

---

### Checkpoint 3: Auth UI works (dev mode)

- [ ] Login flow works against local backend (`npm run dev` + `go run`)
- [ ] Protected routes enforce auth
- [ ] Design system visible on login page
- [ ] Review before shell polish

---

### Phase 4: Dashboard shell and Docker integration

---

## Task 13: App shell, sidebar, and placeholder pages

**Description:** Layout with grouped sidebar navigation, "Soon" badges, admin-only Users link, logout footer, and placeholder pages for all MVP routes.

**Acceptance criteria:**
- [ ] `AppShell`: sidebar + header + main content area
- [ ] Sidebar groups: Dashboard | Storage (Volumes, Monitoring) | Extend (Plugins, Tasks) | System (Settings, Users admin-only)
- [ ] Non-dashboard items show "Soon" badge
- [ ] `Users` link visible only when `user.role === 'admin'`
- [ ] Placeholder pages for `/volumes`, `/monitoring`, `/plugins`, `/tasks`, `/settings`, `/settings/users`
- [ ] `DashboardPage`: welcome placeholder
- [ ] `NotFoundPage` within shell
- [ ] Logout button calls API + clears store

**Verification:**
- [ ] Manual: navigate all sidebar links without crash (SC0.9)
- [ ] Admin sees Users; regular user does not

**Dependencies:** Task 12

**Files likely touched:**
- `web/src/components/layout/AppShell.tsx`, `Sidebar.tsx`, `Header.tsx`
- `web/src/pages/DashboardPage.tsx`, `PlaceholderPage.tsx`, `NotFoundPage.tsx`
- `web/src/App.tsx`

**Maps to:** SC0.9

**Estimated scope:** M (5–8 files)

---

## Task 14: Embed SPA in Go server

**Description:** Production build embedded/served by Gin with SPA fallback routing; finalize Dockerfile web build stage.

**Acceptance criteria:**
- [ ] `go:embed` or `embed.FS` serves `web/dist` on `/`
- [ ] Unknown non-API paths return `index.html` (client-side routing)
- [ ] `/api/*` never caught by static handler
- [ ] Dockerfile web-builder runs `npm ci && npm run build`; final image runs single binary

**Verification:**
- [ ] `docker compose up -d --build`
- [ ] `curl http://localhost:8080/` → 200 HTML (SC0.1, SC0.2)
- [ ] Direct navigation to `/dashboard` after login works (SPA fallback)

**Dependencies:** Task 13, Task 8

**Files likely touched:**
- `cmd/server/main.go`, `internal/api/router.go`
- `Dockerfile`

**Maps to:** SC0.1, SC0.2

**Estimated scope:** M (3 files)

---

## Task 15: Frontend unit tests

**Description:** Vitest tests for critical auth UI paths.

**Acceptance criteria:**
- [ ] `LoginPage` renders form fields and submit button
- [ ] Auth store: login sets tokens, logout clears
- [ ] `ProtectedRoute`: redirects when unauthenticated

**Verification:**
- [ ] `cd web && npm run test` passes (SC0.12)

**Dependencies:** Task 13

**Files likely touched:**
- `web/src/pages/LoginPage.test.tsx`
- `web/src/store/auth.test.ts`
- `web/vitest.config.ts` (if missing)

**Maps to:** SC0.12

**Estimated scope:** S (3–4 files)

---

## Task 16: End-to-end verification and README sync

**Description:** Run full success criteria checklist; align README with actual setup behavior.

**Acceptance criteria:**
- [ ] All SC0.1–SC0.13 pass (see checklist below)
- [ ] README commands match working flow
- [ ] No secrets in repo

**Verification:**
- [ ] Full curl script from [SPEC.md](../SPEC.md#verification-phase-0-done)
- [ ] Manual UC0.1 + UC0.2 walkthrough
- [ ] `go test ./...` + `cd web && npm run test`

**Dependencies:** Task 14, Task 15

**Files likely touched:**
- `README.md` (minor fixes if needed)

**Maps to:** SC0.1–SC0.13, UC0.1, UC0.2

**Estimated scope:** S (1 file + manual)

---

### Checkpoint 4: Phase 0 complete

- [ ] `docker compose up -d --build` → login → dashboard → sidebar navigation
- [ ] All success criteria green
- [ ] Ready for `/review` then `/ship`

---

## Success criteria traceability

| Criterion | Task(s) |
|---|---|
| SC0.1 Docker up | 3, 14, 16 |
| SC0.2 Login page design | 10, 12, 14 |
| SC0.3 Health | 3 |
| SC0.4 Admin seed | 7 |
| SC0.5 Login tokens | 5, 6 |
| SC0.6 /me | 6 |
| SC0.6b Refresh | 5, 6 |
| SC0.6c Logout | 5, 6 |
| SC0.7 Admin users | 7 |
| SC0.8 Migrations | 4 |
| SC0.9 Sidebar nav | 13 |
| SC0.10 Route guard | 12 |
| SC0.11 .env.example | 1 |
| SC0.12 Tests | 8, 15, 16 |
| SC0.13 README | 16 |

---

## Risks and mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Go 1.25 not available locally | Medium | Dockerfile pins Go version; document minimum in README |
| shadcn/ui v3 + Tailwind v4 setup friction | Medium | Follow official init; lock versions in package.json |
| Refresh token in sessionStorage (XSS) | Medium | Document as Phase 0 trade-off; httpOnly cookie in hardening ADR |
| Alpine musl + CGO | Low | Build with `CGO_ENABLED=0` static binary |
| Docker build time (npm + go) | Low | Multi-stage cache layers; acceptable for Phase 0 |

---

## Parallelization

| Parallel track A (backend) | Parallel track B (frontend) |
|---|---|
| Tasks 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 | Tasks 9 → 10 (after Task 3) |
| | Task 11 waits for Task 6 |
| | Tasks 12 → 13 → 15 |
| **Merge at Task 14** | |

Single-agent default order: **1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10 → 11 → 12 → 13 → 14 → 15 → 16**

---

## Next step

Run `/build` on **Task 1** (create branch `feature/phase-0-foundations` first if on `main`).
