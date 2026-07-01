# Phase 0 — Todo

> Detailed plan: [plan.md](./plan.md) · Spec: [SPEC.md](../SPEC.md)

**Branch:** `feature/phase-0-foundations`

## Phase 1: Foundation

- [x] **Task 1** — Repository scaffold
- [x] **Task 2** — Config loader and HTTP helpers
- [x] **Task 3** — Docker Compose and health endpoint stub

### Checkpoint 1
- [x] `docker compose up` + `/api/health` OK

## Phase 2: Backend

- [x] **Task 4** — GORM models and AutoMigrate
- [x] **Task 5** — Auth service (login, JWT, refresh, logout)
- [x] **Task 6** — Auth middleware, handlers, and router
- [x] **Task 7** — Admin seed and user creation endpoint
- [x] **Task 8** — Auth integration tests

### Checkpoint 2
- [x] Auth API complete via curl · `go test ./...` green

## Phase 3: Frontend

- [x] **Task 9** — Vite + React + TypeScript scaffold
- [x] **Task 10** — Tailwind design tokens and shadcn/ui
- [x] **Task 11** — API client, auth store, and silent refresh
- [x] **Task 12** — Login page and route guards

### Checkpoint 3
- [x] Login flow works in dev mode

## Phase 4: Integration

- [x] **Task 13** — App shell, sidebar, and placeholder pages
- [x] **Task 14** — Embed SPA in Go server
- [x] **Task 15** — Frontend unit tests
- [x] **Task 16** — E2E verification and README sync

### Checkpoint 4 — Phase 0 done
- [x] SC0.1–SC0.13 all pass
- [ ] Ready for `/review` → `/ship`
