# Spec: Phase 0 — Technical foundations

> **Status:** Implemented on `feature/phase-0-foundations` — pending `/review`
> **Scope:** Bootstrap the full stack from zero to a running Docker deployment with auth and UI skeleton.
> **Sources:** [STARTUP.md](./STARTUP.md#phase-0--technical-foundations), [VISION.md](./VISION.md), [docs/project.md](./docs/project.md)

---

## Assumptions (correct me now or I proceed)

1. **Greenfield repo** — only documentation exists today; all code is created in this phase.
2. **No public signup** — "register" in STARTUP means an **admin-only** user-creation endpoint, not self-service registration ([docs/project.md](./docs/project.md#tech-stack)).
3. **Initial admin from env** — on first startup, if no users exist, seed one admin from `ADMIN_EMAIL` / `ADMIN_PASSWORD` in `.env`.
4. **JWT transport** — Bearer access token in `Authorization` header; refresh token stored separately (httpOnly cookie or secure storage — see Auth flow). Access token 24h, refresh token 7 days, both configurable via env.
5. **Single Docker image** — Go server serves the built React SPA as static files on the same port (`8080`). No separate frontend container for MVP bootstrap.
6. **Volume model is schema-only** — GORM model + migration exist; no volume CRUD or disk operations until Phase 1.1.
7. **GORM AutoMigrate** — used for Phase 0 development migrations; production SQL migrations documented later per [docs/project.md](./docs/project.md#database-migrations).
8. **Local URL** — `http://localhost:8080` ([docs/project.md](./docs/project.md#urls)).
9. **Design system** — UI follows [design/DESIGN.md](./design/DESIGN.md) (dark canvas `#0a0a0a`, electric yellow `#faff69`, Inter + JetBrains Mono).
10. **Reverse proxy / TLS** — out of scope; deployer handles HTTPS in front of Docker ([STARTUP.md](./STARTUP.md#open-assumptions)).

---

## Objective

Set up the complete technical environment before any visible functional development (volumes, files, plugins, tasks).

**Who:** Experienced self-hosters who clone the repo and run `docker compose up`.

**User stories:**

| ID | Story |
|---|---|
| UC0.1 | As a deployer, I clone the repo, copy `.env.example` to `.env`, run `docker compose up`, open `http://localhost:8080`, and see the login page styled per the design system. |
| UC0.2 | As an admin, I log in with credentials from `.env`, land on an empty dashboard shell, and can navigate placeholder routes without errors. |
| UC0.3 | As an admin, I can create additional user accounts via API (admin-only); regular users cannot self-register. |

**Out of scope for Phase 0:**

- Volume CRUD, file upload/download, indexing, plugins, tasks, monitoring
- User management UI (API only for admin user creation)
- Production-grade migration tooling, CI pipeline, reverse proxy setup
- Password reset, email verification, OAuth

---

## Tech Stack

See [docs/project.md — Tech Stack](./docs/project.md#tech-stack).

Phase 0 uses only the subset needed for bootstrap:

| Layer | Choices |
|---|---|
| Backend | Go 1.25+, Gin, GORM, PostgreSQL 17, golang-jwt/jwt v5 |
| Frontend | React 19, Vite 8, TypeScript, Tailwind CSS, shadcn/ui, React Router, TanStack Query, Zustand |
| Infra | Docker Compose (Go service + PostgreSQL 17) |

---

## Commands

### Prerequisites

```bash
# Required on host for local dev (optional if Docker-only)
go version    # 1.25+
node -v       # 20+
docker compose version
```

### Bootstrap (first time)

```bash
git clone https://github.com/<username>/lcloud
cd lcloud
cp .env.example .env
# Edit .env — set POSTGRES_*, JWT_SECRET, ADMIN_EMAIL, ADMIN_PASSWORD, STORAGE_BASE_PATH
docker compose up -d --build
```

### Runtime

```bash
# Full stack (production-like)
docker compose up -d
docker compose logs -f app

# Tear down
docker compose down
```

### Backend (local dev, outside Docker)

```bash
go mod tidy
go run ./cmd/server
go test ./...
go test ./... -cover
```

### Frontend (local dev with hot reload)

```bash
cd web
npm install
npm run dev          # Vite dev server (proxies API to backend)
npm run build        # Production build → embedded by Go server
npm run test         # Vitest
npm run lint         # ESLint (if configured)
```

### Verification (Phase 0 done)

```bash
# 1. Stack healthy
docker compose ps
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/          # expect 200 (SPA)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/health  # expect 200

# 2. Login
curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"<ADMIN_EMAIL>","password":"<ADMIN_PASSWORD>"}'
# expect 200 + JWT

# 3. Protected route
curl -s http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer <token>"
# expect 200 + user payload
```

---

## Project Structure

See [docs/project.md — Project structure](./docs/project.md#project-structure).

Phase 0 creates these paths (empty dirs get `.gitkeep` where needed):

```
lcloud/
├── cmd/server/main.go              ← entry: DB connect, migrate, seed admin, Gin, static SPA
├── internal/
│   ├── auth/
│   │   ├── model.go                ← User GORM model
│   │   ├── service.go              ← login, create user, password hash
│   │   ├── middleware.go           ← JWT validation, role checks
│   │   └── service_test.go
│   ├── volume/
│   │   └── model.go                ← Volume GORM model (schema only)
│   ├── api/
│   │   ├── router.go               ← route registration
│   │   ├── auth_handler.go
│   │   └── admin_handler.go        ← admin user creation
│   └── config/
│       └── config.go               ← env loading
├── pkg/
│   └── httputil/                   ← shared JSON helpers, error responses
├── web/
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── pages/
│   │   │   ├── LoginPage.tsx
│   │   │   └── DashboardPage.tsx   ← empty shell
│   │   ├── components/
│   │   │   ├── layout/             ← AppShell, Sidebar, Header
│   │   │   └── ui/                 ← shadcn components
│   │   ├── hooks/
│   │   ├── lib/
│   │   │   ├── api.ts              ← fetch wrapper + auth header
│   │   │   └── utils.ts
│   │   └── store/
│   │       └── auth.ts             ← Zustand auth store
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   └── tsconfig.json
├── plugins/.gitkeep
├── docs/adr/.gitkeep
├── docker-compose.yml
├── Dockerfile                      ← multi-stage: build web → build Go → runtime
├── go.mod
├── go.sum
├── .env.example
└── .gitignore
```

Modules **not** created in Phase 0 (stub folders optional, no implementation): `monitoring/`, `indexer/`, `plugin/`, `task/`.

---

## Data models

### User (`internal/auth/model.go`)

| Field | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | Generated on create |
| `email` | string, unique, not null | Login identifier |
| `password_hash` | string, not null | bcrypt |
| `role` | enum: `admin` \| `user` | Default `user` |
| `created_at` | timestamp | Auto |
| `updated_at` | timestamp | Auto |

### Volume (`internal/volume/model.go`)

Base schema only — aligned with future `.volume.json` fields ([docs/project.md](./docs/project.md#volume-structure-on-disk)):

| Field | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | Matches future volume UUID |
| `name` | string | Display name |
| `owner_id` | UUID (FK → users) | Owner reference |
| `disk_path` | string | Physical path (unused until Phase 1.1) |
| `quota_bytes` | int64 | Default 0 |
| `created_at` | timestamp | Auto |
| `updated_at` | timestamp | Auto |

### RefreshToken (`internal/auth/model.go`)

| Field | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | |
| `user_id` | UUID (FK → users) | |
| `token_hash` | string | SHA-256 of opaque token |
| `expires_at` | timestamp | Default now + 7 days |
| `revoked_at` | timestamp, nullable | Set on logout or rotation |
| `created_at` | timestamp | Auto |

No volume API endpoints in Phase 0.

---

## API (Phase 0)

Base path: `/api`

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | Public | `{ "status": "ok" }` — liveness |
| `POST` | `/auth/login` | Public | `{ email, password }` → `{ access_token, refresh_token, user }` |
| `POST` | `/auth/refresh` | Public | `{ refresh_token }` → `{ access_token, refresh_token? }` (rotation optional) |
| `POST` | `/auth/logout` | JWT or refresh | Invalidates refresh token server-side |
| `GET` | `/auth/me` | JWT (access) | Current user profile |
| `POST` | `/admin/users` | JWT + admin | `{ email, password, role? }` → created user (no password in response) |

**Error shape (all endpoints):**

```json
{ "error": "human-readable message", "code": "INVALID_CREDENTIALS" }
```

**Access token (JWT):** claims `sub`, `email`, `role`, `exp` — default **24h**, configurable via `JWT_EXPIRY_HOURS`.

**Refresh token:** opaque random token (not a JWT), persisted hashed in PostgreSQL (`refresh_tokens` table: `user_id`, `token_hash`, `expires_at`, `revoked_at`). Default lifetime **7 days**, configurable via `REFRESH_TOKEN_EXPIRY_DAYS`. Rotated on each `/auth/refresh` call (old token revoked, new one issued).

---

## Frontend (Phase 0)

### Routes

| Path | Access | Content |
|---|---|---|
| `/login` | Public | Login form (email + password), yellow primary CTA per design system |
| `/` | Protected | Redirect to `/dashboard` |
| `/dashboard` | Protected | Empty shell — sidebar + header + "Welcome" placeholder |
| `/volumes` | Protected | Placeholder ("Phase 1.1") |
| `/monitoring` | Protected | Placeholder ("Phase 1.2") |
| `/plugins` | Protected | Placeholder ("Phase 1.3") |
| `/tasks` | Protected | Placeholder ("Phase 1.4") |
| `/settings` | Protected | Placeholder (system preferences) |
| `/settings/users` | Protected, admin | Placeholder ("Phase 0 — API only") |
| `*` | — | 404 page within app shell |

### Sidebar navigation (Phase 0 skeleton)

Grouped layout aligned with MVP phases — all non-dashboard items show a "Soon" badge until their phase ships:

| Group | Item | Route | Phase |
|---|---|---|---|
| — | Dashboard | `/dashboard` | 0 (active) |
| Storage | Volumes | `/volumes` | 1.1 |
| Storage | Monitoring | `/monitoring` | 1.2 |
| Extend | Plugins | `/plugins` | 1.3 |
| Extend | Tasks | `/tasks` | 1.4 |
| System | Settings | `/settings` | 0 (placeholder) |
| System | Users *(admin only)* | `/settings/users` | 0 (API only) |

Logout action in sidebar footer (not a route).

### Auth flow

1. Unauthenticated access to protected routes → redirect `/login`
2. Login success → store `access_token` (Zustand + sessionStorage) and `refresh_token` (httpOnly cookie preferred; fallback secure sessionStorage for Phase 0 if cookie setup deferred) → navigate `/dashboard`
3. TanStack Query fetches `/api/auth/me` on app load when access token present
4. On 401 from API → silent refresh via `POST /auth/refresh` → retry request; if refresh fails → logout
5. Logout calls `POST /auth/logout`, clears tokens, redirects `/login`

### Design integration

- Tailwind theme tokens mapped from [design/DESIGN.md](./design/DESIGN.md): `canvas`, `primary`, `surface-card`, `hairline`, etc.
- shadcn/ui initialized in dark mode; components use design tokens
- Fonts: Inter (UI), JetBrains Mono (code snippets if any)
- No feature UI beyond login + empty dashboard shell

---

## Docker & environment

### `docker-compose.yml`

Services:

| Service | Image / build | Ports | Notes |
|---|---|---|---|
| `app` | `Dockerfile` build | `8080:8080` | Go + embedded SPA |
| `postgres` | `postgres:17` | internal only | Persistent volume |

`app` depends on `postgres` (healthcheck before start).

### `.env.example` variables

| Variable | Required | Description |
|---|---|---|
| `POSTGRES_USER` | yes | DB user |
| `POSTGRES_PASSWORD` | yes | DB password |
| `POSTGRES_DB` | yes | Database name |
| `DATABASE_URL` | auto | Composed in compose; documented for local dev |
| `JWT_SECRET` | yes | Access token signing key (min 32 chars) |
| `JWT_EXPIRY_HOURS` | no | Access token lifetime; default `24` |
| `REFRESH_TOKEN_EXPIRY_DAYS` | no | Refresh token lifetime; default `7` |
| `ADMIN_EMAIL` | yes | Seed admin email |
| `ADMIN_PASSWORD` | yes | Seed admin password (min 8 chars) |
| `STORAGE_BASE_PATH` | yes | Host path mounted for future volumes (e.g. `./data/storage`) |
| `APP_PORT` | no | Default `8080` |
| `GIN_MODE` | no | Default `release` in Docker |

### Dockerfile (multi-stage)

1. **web-builder** — `npm ci && npm run build`
2. **go-builder** — `go build` with embedded `web/dist`
3. **runtime** — Alpine + ca-certs; non-root user; expose 8080

---

## Code Style

### Go

Follow standard Go conventions. Handlers are thin; business logic lives in services.

```go
// internal/api/auth_handler.go — handler calls service, maps errors to HTTP
func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        httputil.BadRequest(c, "invalid request body")
        return
    }
    result, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        httputil.Unauthorized(c, "invalid credentials")
        return
    }
    httputil.JSON(c, http.StatusOK, result)
}
```

- Package naming: lowercase, no underscores
- Errors: sentinel errors in service layer; never leak internal details in HTTP responses
- Context passed as first arg to service methods

### TypeScript / React

- Functional components only; named exports for pages
- Colocate component-specific types in the same file unless shared
- API calls through `lib/api.ts`; no raw fetch in page components
- Tailwind utility classes; avoid inline styles

```tsx
// web/src/pages/LoginPage.tsx — pattern
export function LoginPage() {
  const login = useAuthStore((s) => s.login);
  // form state, submit → api.login → store token → navigate
  return (
    <main className="flex min-h-screen items-center justify-center bg-canvas">
      {/* shadcn Card + Input + Button variant="default" (yellow CTA) */}
    </main>
  );
}
```

---

## Testing Strategy

See [docs/project.md — Coverage targets](./docs/project.md#coverage-targets).

| Layer | Framework | Location | Phase 0 focus |
|---|---|---|---|
| Backend unit | Go testing + testify | `internal/*/*_test.go` | Auth service: hash, login, JWT issue/validate, admin seed |
| Backend integration | Go testing + testify | `internal/api/*_test.go` | Login + `/me` + admin create user (test DB or sqlite stub if feasible) |
| Frontend unit | Vitest + Testing Library | `web/src/**/*.test.tsx` | LoginPage render, auth store, protected route redirect |

**Coverage target for Phase 0 critical path (auth):** 80% on `internal/auth/`.

**Verify before ship:**

```bash
go test ./... -cover
cd web && npm run test
docker compose up -d --build && manual UC0.1 + UC0.2
```

---

## Boundaries

### Always

- Run `go test ./...` and `npm run test` before marking phase complete
- Handlers call services only — no business logic in `internal/api/`
- Reference [docs/project.md](./docs/project.md) for stack/structure — do not duplicate into other docs
- Follow [design/DESIGN.md](./design/DESIGN.md) for any UI work
- Keep `.env` out of git; maintain `.env.example` with all variables documented

### Ask first

- Adding dependencies not listed in [docs/project.md](./docs/project.md#approved-third-party-libraries)
- Changing JWT transport (cookie vs Bearer) or token lifetime defaults
- Adding CI/CD pipeline configuration
- Any change to Volume model fields that would conflict with future `.volume.json` schema

### Never

- Commit secrets or real `.env` files
- Public self-registration endpoint
- Volume CRUD or file operations (Phase 1.1)
- Direct Bleve, go-plugin, or gocron setup (later phases)
- Business logic in Gin handlers
- Duplicating tech stack content outside `docs/project.md`

---

## Success Criteria

Phase 0 is **done** when all of the following pass:

- [ ] **SC0.1** `docker compose up -d --build` starts `app` + `postgres` without manual steps beyond `.env`
- [ ] **SC0.2** `GET http://localhost:8080/` serves the React app; login page matches design system (dark canvas, yellow CTA)
- [ ] **SC0.3** `GET /api/health` returns 200
- [ ] **SC0.4** Admin seed: first boot creates admin from `ADMIN_EMAIL` / `ADMIN_PASSWORD` if no users exist
- [ ] **SC0.5** `POST /api/auth/login` returns access token (24h) + refresh token (7d)
- [ ] **SC0.6** `GET /api/auth/me` with access token returns user profile; without token returns 401
- [ ] **SC0.6b** `POST /api/auth/refresh` with valid refresh token returns new access token; expired/revoked token returns 401
- [ ] **SC0.6c** `POST /api/auth/logout` revokes refresh token — subsequent refresh fails
- [ ] **SC0.7** `POST /api/admin/users` creates a user when called by admin; returns 403 for non-admin
- [ ] **SC0.8** GORM AutoMigrate creates `users` and `volumes` tables in PostgreSQL
- [ ] **SC0.9** Frontend: login → empty dashboard; sidebar navigates to placeholder routes without crash
- [ ] **SC0.10** Unauthenticated access to `/dashboard` redirects to `/login`
- [ ] **SC0.11** `.env.example` documents every required variable with comments
- [ ] **SC0.12** `go test ./...` and `cd web && npm run test` pass
- [ ] **SC0.13** README local setup instructions match actual behavior

---

## Implementation order (preview for `/plan`)

Suggested vertical slices — not tasks yet; `/plan` will expand:

1. **Scaffold** — `go mod init`, directory tree, `.gitignore`, `.env.example`
2. **Docker** — `Dockerfile`, `docker-compose.yml`, health endpoint, postgres wiring
3. **Auth backend** — User model, bcrypt, JWT middleware, login + admin create + seed
4. **Volume model** — GORM model + migrate only
5. **Frontend scaffold** — Vite + React + Tailwind + shadcn + design tokens
6. **Auth frontend** — Login page, auth store, API client, route guards
7. **Dashboard shell** — Layout, sidebar, placeholder routes
8. **Integration** — Embed SPA in Go, end-to-end Docker verification

---

## Decisions log

| # | Decision | Status |
|---|---|---|
| D1 | Access token lifetime: **24h** (`JWT_EXPIRY_HOURS`, default `24`) | ✅ Validated |
| D2 | Refresh token lifetime: **7 days** (`REFRESH_TOKEN_EXPIRY_DAYS`, default `7`) | ✅ Validated |
| D3 | Sidebar: grouped MVP nav with "Soon" badges (see Sidebar navigation) | ✅ Validated |
| D4 | Docker runtime image: **Alpine** for Phase 0; Distroless before MVP (ADR) | ✅ Validated |

## Open Questions

| # | Question | Default if unanswered |
|---|---|---|
| OQ1 | Docker runtime: Alpine vs Distroless? | See comparison in this spec + user choice |
| OQ2 | Refresh token storage client-side: httpOnly cookie vs sessionStorage? | httpOnly cookie (more secure) |
| OQ3 | GitHub repo URL for README clone command? | Placeholder `<username>/lcloud` until repo published |

### Docker runtime image — Alpine vs Distroless

| | **Alpine** | **Distroless** (Google) |
|---|---|---|
| **Taille image** | ~5–15 Mo de base + binaire | ~2–5 Mo de base + binaire — plus léger |
| **Sécurité** | Contient un shell (`sh`) et parfois `apk` — surface d'attaque plus large | Pas de shell, pas de package manager — très difficile à exploiter si compromis |
| **Debug en prod** | `docker exec -it … sh` — inspecter logs, fichiers, réseau facilement | Pas de shell : debug via logs Docker, métriques, ou image debug séparée |
| **Compatibilité Go** | Parfois besoin de `libc` musl vs glibc — rares surprises avec Go statique | Go compilé statiquement (`CGO_ENABLED=0`) — excellent fit |
| **Courbe d'apprentissage** | Standard, documenté partout | Moins intuitif au début ; Dockerfile légèrement plus exigeant |
| **Recommandation Phase 0** | **Pratique pour bootstrap** — tu itères vite, tu débugues facilement | **Meilleur choix long terme** — aligné avec "production security" dans project.md |

**Suggestion:** Alpine en Phase 0 pour itérer sans friction ; basculer vers Distroless avant MVP ou documenter la migration en ADR.

---

## Next step

After spec approval → `/plan` to produce `tasks/plan.md` with ordered, verifiable tasks.
