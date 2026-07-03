# Spec: Dashboard configurable

> **Status:** Approved  
> **Scope:** Replace the Phase 0 `/dashboard` placeholder with a personal, widget-based home page. Each user arranges a grid of mini-views (storage, tasks, plugins, quick actions) and persists that layout server-side so it follows them across devices.  
> **Prerequisite:** MVP modules shipped (monitoring, volumes, tasks, plugins, auth).  
> **Sources:** [VISION.md](../../../VISION.md) (Stage 1 — *"a dashboard that shows what's happening"*), [docs/project.md](../../project.md), [IDEAS.md — Dashboard home — configurable widget layout](../../../IDEAS.md), [design/DESIGN.md](../../../design/DESIGN.md)

---

## Assumptions (correct me now or I proceed)

1. **Per-user layout, not per-role template** — each account (`admin` or `user`) has their own saved layout. Admins do not push a global layout to other users.
2. **Server persistence in PostgreSQL** — layout JSON stored per `user_id`, not `localStorage`. Same dashboard on phone, laptop, or after clearing browser data.
3. **Widgets are mini-views, not duplicates** — widgets surface a **summary** of existing modules and link to full pages (`/monitoring`, `/tasks`, etc.). They reuse existing APIs and hooks; no parallel backend for widget data in v1.
4. **Role-aware catalog, not role-aware layout** — the widget picker hides entries the user cannot use (e.g. admin-only widgets). APIs already scope data by role; widgets inherit that behaviour.
5. **Fixed widget sizes in v1** — each widget type declares a default grid footprint (`w` × `h`). Users add, remove, and reposition widgets; **no drag-resize handles** in v1 (avoids `react-grid-layout` dependency; `@dnd-kit` already in the project).
6. **12-column responsive grid** — layout coordinates use abstract grid units. CSS renders the grid; breakpoints collapse gracefully on narrow viewports (single column stack).
7. **Default layout on first visit** — new users (or users with no saved layout) see a curated default. No empty dashboard.
8. **Edit mode toggle** — customization happens in an explicit « Customize dashboard » mode; view mode is read-only and loads data normally.
9. **Monitoring visual upgrade is a soft dependency** — widgets work with today's monitoring API (yellow bars, snapshot stats). Colored quota thresholds and health pills from [IDEAS.md — Monitoring dashboard — visual upgrade](../../../IDEAS.md) can land later and automatically improve storage widgets without a dashboard spec change.
10. **No plugin-authored widgets in v1** — widget catalog is a closed, core list. Third-party widget plugins deferred.
11. **No shared layouts or templates** — users cannot export/import layouts or inherit an admin's arrangement.
12. **GORM AutoMigrate** — new `user_dashboard_layouts` table via model + AutoMigrate in dev; production SQL note in implementation PR.
13. **Activity / notifications widget deferred** — depends on [IDEAS.md — In-app notifications](../../../IDEAS.md); not in v1 catalog.

---

## Problem statement

Today `/dashboard` is a Phase 0 shell:

```16:17:web/src/pages/DashboardPage.tsx
            Welcome{user ? `, ${user.email}` : ""}. This is the empty dashboard shell for Phase 0.
```

Meanwhile, full module pages exist and work:

| Page | Route | Backend |
|---|---|---|
| Monitoring | `/monitoring` | `GET /api/monitoring/overview`, volume stats |
| Volumes | `/volumes` | `GET /api/volumes` |
| Tasks | `/tasks` | `GET /api/tasks`, runs |
| Plugins | `/plugins` | `GET /api/plugins`, logs |

**Gap:** After login, users land on an empty page. VISION Stage 1 expects a cockpit that shows *what's happening* — aggregated, personal, at a glance — without duplicating entire module UIs.

---

## Objective

**Who:** Any authenticated user (`admin` or `user`) on a self-hosted lcloud instance.

**Why:** A personal home page — arrange the summaries you care about, in the order you want, on any device.

**User stories:**

| ID | Story |
|---|---|
| DH1 | As a user, I land on `/dashboard` and see a useful default layout (not a placeholder). |
| DH2 | As a user, I enter « Customize dashboard » mode, add widgets from a catalog, remove widgets, and drag to reposition them. |
| DH3 | As a user, I save my layout; when I log in from another browser, the same arrangement appears. |
| DH4 | As a user, I reset my layout to the instance default. |
| DH5 | As a user, each widget shows a compact summary and links to the relevant full page. |
| DH6 | As a `user`, I only see widgets and data scoped to my account (volumes, tasks). |
| DH7 | As an `admin`, I see admin-only widgets (e.g. pending volume deletions) in the catalog. |
| DH8 | As a user, I cannot add a widget type my role does not have access to (server rejects invalid layout). |
| DH9 | As a user, if a widget's data fails to load, other widgets still render (isolated error states). |

**Out of scope (v1):**

- Drag-resize handles / free-form widget dimensions
- Third-party or plugin-provided widgets
- Shared layouts, layout templates, admin-pushed defaults per role
- Widget-level configuration beyond type (e.g. « pick which volume ») — deferred per-widget `config` JSON field
- In-app notifications / activity feed widget
- Recent files / starred files widgets ([IDEAS.md — Favorites & recent](../../../IDEAS.md))
- Unified Spotlight search on the dashboard
- Mobile-native widget SDK
- Real-time WebSocket updates (widgets poll/refetch via React Query like today)

### Future scope (explicitly not v1)

- Per-widget config (selected volume, task filter, row limit)
- Colored quota thresholds in storage widgets (when monitoring upgrade ships)
- Activity / alerts widget (when notification center ships)
- `recent-files` and `starred-files` widgets
- Optional layout export/import JSON for backup

---

## Phasing

IDEAS.md proposed Phase A (fixed curated home) then Phase B (full customization). **This spec ships both in one release** — the default layout covers Phase A; edit mode + persistence covers Phase B. Rationale: the backend model and widget registry are required for either; shipping customization without persistence would force a breaking migration later.

| Slice | Delivers |
|---|---|
| **Slice 1 — Backend layout API** | `user_dashboard_layouts` table, GET/PATCH endpoints, validation, default layout seed |
| **Slice 2 — Widget registry + default home** | Replace placeholder; render default layout in view mode |
| **Slice 3 — Edit mode** | Customize toggle, add/remove, drag reposition, save, reset |
| **Slice 4 — Widget polish** | Loading/error/empty states, deep links, admin-only widget |

---

## Tech Stack

See [docs/project.md — Tech Stack](../../project.md#tech-stack).

This feature extends:

| Layer | Change |
|---|---|
| Backend | `internal/dashboard/` (new) — layout model, validation, default layout |
| Backend | `internal/api/dashboard_handler.go` — GET/PATCH layout |
| Backend | `internal/api/router.go` — `/api/users/me/dashboard` |
| Database | PostgreSQL `user_dashboard_layouts` |
| Frontend | `web/src/pages/DashboardPage.tsx` — grid shell, edit mode |
| Frontend | `web/src/components/dashboard/` — widget components, grid, picker |
| Frontend | `web/src/hooks/useDashboardLayout.ts` — React Query |
| Frontend | `web/src/lib/dashboard/` — widget registry, types, default layout |

**Dependencies:** No new npm packages. Reuse `@dnd-kit/core` (already used in volume explorer) for drag reposition in edit mode. CSS Grid for layout rendering.

---

## Commands

### Development

```bash
# Backend
go test ./internal/dashboard/...
go test ./internal/api/... -run Dashboard

# Frontend
cd web && npm run test
cd web && npm run build

# Full stack smoke
make embed
docker compose up -d --build
# → http://localhost:8080/dashboard (admin + user accounts)
```

### Lint / typecheck

```bash
go test ./...
cd web && npm run build
```

---

## Data model

### Table: `user_dashboard_layouts`

| Column | Type | Notes |
|---|---|---|
| `user_id` | `uuid` PK, FK → `users.id` | One row per user |
| `layout` | `jsonb` not null | Validated widget layout (see schema below) |
| `updated_at` | `timestamp` | Set on PATCH |

**Bootstrap:** No row until first PATCH or lazy create on first GET (implementation choice — prefer **lazy create on GET** returning default without insert, insert on first PATCH).

### Layout JSON schema

```jsonc
{
  "version": 1,
  "widgets": [
    {
      "id": "w-uuid-or-nanoid",       // unique within layout
      "type": "storage-summary",      // catalog key
      "x": 0,                         // grid column start (0–11)
      "y": 0,                         // grid row start (0-based)
      "w": 6,                         // width in columns (1–12)
      "h": 2                          // height in row units
    }
  ]
}
```

**Validation rules (server):**

| Rule | Error |
|---|---|
| `version` must be `1` | `400 INVALID_LAYOUT` |
| `widgets` length 1–20 | `400 INVALID_LAYOUT` |
| Each `type` exists in catalog | `400 UNKNOWN_WIDGET` |
| Each `type` allowed for caller's role | `403 WIDGET_NOT_ALLOWED` |
| Unique `id` within layout | `400 INVALID_LAYOUT` |
| `w` ∈ [1, 12], `x` + `w` ≤ 12 | `400 INVALID_LAYOUT` |
| `h` ∈ [1, 4], `x` ≥ 0, `y` ≥ 0 | `400 INVALID_LAYOUT` |
| `w`/`h` match catalog defaults for `type` (v1 fixed sizes) | `400 INVALID_LAYOUT` |
| No overlapping widgets (optional v1 — **reject overlaps**) | `400 INVALID_LAYOUT` |

---

## Widget catalog (v1)

Fixed footprints on a 12-column grid. All users:

| `type` | Title | Default `w×h` | Data source | Deep link |
|---|---|---|---|---|
| `storage-summary` | Storage | 6×2 | `GET /api/monitoring/overview` | `/monitoring` |
| `volumes-at-risk` | Volumes at risk | 6×2 | overview volumes where `used_bytes / quota_bytes ≥ 0.85` | `/monitoring` |
| `tasks-overview` | Tasks | 6×2 | `GET /api/tasks` + latest run per task | `/tasks` |
| `plugins-status` | Plugins | 4×2 | `GET /api/plugins` | `/plugins` |
| `quick-actions` | Quick actions | 4×1 | static links | various |
| `welcome` | Welcome | 4×1 | `GET /api/auth/me` | — |

Admin only (`role === admin`):

| `type` | Title | Default `w×h` | Data source | Deep link |
|---|---|---|---|---|
| `pending-deletions` | Pending deletions | 6×2 | `GET /api/admin/volume-deletion-requests` | `/volumes` |

**Widget behaviour notes:**

- **`storage-summary`** — Compact row: total disks, aggregate free space, volume count. Reuse `DiskOverviewCard` visual language at reduced density (max 2 disks shown; « +N more » link).
- **`volumes-at-risk`** — List up to 5 volumes over 85% quota with progress bars. Empty state: « All volumes healthy ».
- **`tasks-overview`** — Show count enabled/disabled, last failed run (if any), next 3 enabled tasks by schedule. Empty state: CTA to create first task.
- **`plugins-status`** — Count by status (`running` / `stopped` / `error`). Highlight plugins in `error`.
- **`quick-actions`** — Buttons: « New volume », « Monitoring », « Tasks », « Plugins » (role-appropriate).
- **`welcome`** — Greeting + instance tagline; optional in default layout.
- **`pending-deletions`** — Admin-only; list open deletion requests with dismiss link to volumes page.

**Default layout (new users):**

```
Row 0: [ welcome 4×1 ] [ quick-actions 4×1 ] [ plugins-status 4×2 spans row 1 ]
Row 1: [ storage-summary 6×2 ] [ volumes-at-risk 6×2 ]
Row 2: [ tasks-overview 6×2 ] [ plugins-status continued ]
```

(Exact coordinates computed at implementation; admin default adds `pending-deletions` replacing `welcome` or as extra row.)

---

## API

### `GET /api/users/me/dashboard`

**Auth:** JWT (any authenticated user).

**Response `200`:**

```json
{
  "layout": {
    "version": 1,
    "widgets": [ /* ... */ ]
  },
  "catalog": [
    {
      "type": "storage-summary",
      "title": "Storage",
      "description": "Disk space and volume usage summary",
      "default_w": 6,
      "default_h": 2,
      "admin_only": false
    }
  ],
  "updated_at": "2026-07-03T12:00:00Z"
}
```

- If no saved row: return **default layout** for role; `updated_at` null.
- `catalog` filtered by caller role (admin sees all entries).

### `PATCH /api/users/me/dashboard`

**Auth:** JWT.

**Body:**

```json
{
  "layout": {
    "version": 1,
    "widgets": [ /* full replacement */ ]
  }
}
```

**Response `200`:** Same shape as GET (saved layout + catalog).

**Errors:**

| Status | Code | When |
|---|---|---|
| 400 | `INVALID_LAYOUT` | Schema / bounds / overlap |
| 400 | `UNKNOWN_WIDGET` | Unknown `type` |
| 403 | `WIDGET_NOT_ALLOWED` | Admin-only widget for `user` role |

**Design choice:** Full layout replacement on PATCH (not per-widget delta). Simpler validation; layout payloads are small (&lt; 20 widgets).

### `DELETE /api/users/me/dashboard` (reset)

**Auth:** JWT.

**Behaviour:** Delete saved row; subsequent GET returns default layout.

**Response `204`:** No content.

---

## Frontend architecture

### Widget registry

```typescript
// web/src/lib/dashboard/registry.ts
export type WidgetType = "storage-summary" | "volumes-at-risk" | /* ... */;

export type WidgetDefinition = {
  type: WidgetType;
  title: string;
  component: React.ComponentType<WidgetProps>;
  defaultW: number;
  defaultH: number;
  adminOnly?: boolean;
};

export const WIDGET_REGISTRY: Record<WidgetType, WidgetDefinition> = { /* ... */ };
```

Each widget component:
- Receives no layout props beyond `type` + `id`
- Fetches its own data via existing hooks (`useMonitoringOverview`, `useTasks`, `usePlugins`, etc.)
- Renders loading skeleton, error card, empty state independently
- Includes a header with title + optional « Open → » link

### Dashboard page modes

| Mode | UI |
|---|---|
| **View** | Grid of widgets, no drag handles, data live |
| **Edit** | « Done » / « Cancel » in header; drag handles; « + Add widget » opens picker sheet; × remove on each widget; « Reset to default » |

**Edit flow:**

1. Enter edit mode → clone layout to local state (Zustand or `useState`)
2. Drag reposition → update `x`/`y` in local state (`@dnd-kit`)
3. Add widget → append with default `w`/`h`, auto-place at next free row
4. Remove widget → filter from local state
5. « Done » → `PATCH` full layout → exit edit mode
6. « Cancel » → discard local state
7. « Reset » → `DELETE` then refetch default

### Grid rendering

- CSS `display: grid; grid-template-columns: repeat(12, 1fr);`
- Each widget: `grid-column: span w; grid-row: span h` + explicit placement via `grid-column-start` / `grid-row-start` from `x`/`y`
- Mobile (`< md`): force single column; ignore `x`/`w`, stack by `y` order

### Header

Replace current static description with:

- **View mode:** « Your personal cockpit » (or role-aware subtitle)
- **Edit mode:** « Customize dashboard » + actions (Done, Cancel, Reset)
- **Customize** button (primary outline) visible in view mode

Consult [design/DESIGN.md](../../../design/DESIGN.md) — dark cards (`surface-card`), yellow accents on CTAs and stat numbers, `hairline` borders.

---

## Permissions

| Action | `user` | `admin` |
|---|---|---|
| `GET /api/users/me/dashboard` | Own layout + user catalog | Own layout + full catalog |
| `PATCH /api/users/me/dashboard` | Own layout; no admin-only widgets | Own layout; all widgets |
| `DELETE /api/users/me/dashboard` | Reset own layout | Reset own layout |
| Widget data | Scoped by existing APIs | Full instance where applicable |

No new authorization beyond layout validation. Widget components call the same hooks as module pages.

---

## Testing strategy

### Backend (`internal/dashboard/`)

| Test | Covers |
|---|---|
| `ValidateLayout` — valid default | Happy path |
| `ValidateLayout` — unknown type | `UNKNOWN_WIDGET` |
| `ValidateLayout` — admin widget as user | `WIDGET_NOT_ALLOWED` |
| `ValidateLayout` — overlap / out of bounds | `INVALID_LAYOUT` |
| `GetLayout` — no row returns default | Default seed per role |
| `PatchLayout` — persists and returns | Round-trip |
| `ResetLayout` — deletes row | GET returns default after |

### API (`internal/api/`)

| Test | Covers |
|---|---|
| `GET /api/users/me/dashboard` — auth required | 401 |
| `PATCH` + `GET` round-trip | 200 |
| User PATCH with `pending-deletions` | 403 |

### Frontend (Vitest + Testing Library)

| Test | Covers |
|---|---|
| Default layout renders widget slots | No placeholder text |
| Edit mode shows add/remove controls | Mode toggle |
| Widget registry maps all catalog types | No missing components |
| `QuickActionsWidget` renders links | Smoke |

Manual smoke:

- [ ] Admin: customize, save, reload — layout persists
- [ ] User: same flow on separate browser/incognito
- [ ] User: catalog does not show `pending-deletions`
- [ ] Reset restores default
- [ ] Widget error isolation (stop monitoring container → other widgets still show)

---

## Code style

Follow existing patterns:

- Backend: thin handler → `dashboard.Service` → GORM; errors in `internal/dashboard/errors.go`
- Frontend: hooks in `web/src/hooks/`, types in `web/src/lib/dashboard/types.ts`
- Widget components colocated in `web/src/components/dashboard/widgets/`

Handler example (shape only):

```go
func (h *DashboardHandler) GetMe(c *gin.Context) {
    claims := auth.ClaimsFromContext(c)
    result, err := h.service.GetLayout(c.Request.Context(), claims)
    if err != nil { /* ... */ }
    httputil.JSON(c, http.StatusOK, result)
}
```

---

## Boundaries

**Always:**

- Validate widget types server-side — never trust client-only catalog filtering
- Reuse existing API hooks inside widgets — no duplicate fetch logic in a god hook
- Keep widget components small; shared formatting in `lib/utils` / existing monitoring components
- Read `design/DESIGN.md` before UI work

**Ask first:**

- Adding `react-grid-layout` or other layout dependencies
- Per-widget config JSON (schema change)
- Storing layout in `.volume.json` or any on-disk user preference outside PostgreSQL

**Never:**

- Plugin-extensible widget registry in v1
- `localStorage` as source of truth for layout
- Business logic in Gin handlers beyond request parsing
- Bypass role scoping in widget data « for convenience »

---

## Success criteria

- [ ] `/dashboard` no longer shows Phase 0 placeholder text
- [ ] New user sees role-appropriate default layout without configuration
- [ ] User can add, remove, reposition widgets and save; layout persists across sessions
- [ ] User can reset to default
- [ ] Server rejects invalid layouts and admin-only widgets for `user` role
- [ ] Each v1 widget type renders with loading, error, and empty states
- [ ] `go test ./internal/dashboard/...` and dashboard API tests pass
- [ ] `cd web && npm run test && npm run build` pass
- [ ] No new npm dependencies

---

## Open questions

| # | Question | Default if unanswered |
|---|---|---|
| OQ1 | Should admin default layout differ from user default (include `pending-deletions`)? | **Yes** — admin default includes `pending-deletions`; user default does not |
| OQ2 | Auto-place algorithm on add widget — next row vs find first gap? | **Next row** (simpler) |
| OQ3 | Overlap validation — strict reject or auto-resolve on save? | **Strict reject** with clear error; client should prevent |
| OQ4 | Include `welcome` widget in default or rely on header greeting only? | **Include** `welcome` in default (keeps personal touch) |
| OQ5 | Ship `volumes-at-risk` threshold at 85% hardcoded or wait for monitoring upgrade constants? | **85% hardcoded** in widget; extract shared constant when monitoring upgrade lands |

---

## Relationship to other ideas

| Idea | Relationship |
|---|---|
| Monitoring visual upgrade | Soft dependency — storage widgets improve when colored bars ship; no dashboard rework |
| In-app notifications | Future `activity-alerts` widget |
| Favorites & recent | Future `recent-files` / `starred-files` widgets |
| Unified Spotlight | Complementary — search from anywhere; dashboard is glanceable home |

---

## ADR needed?

**No ADR required** for v1 — layout in PostgreSQL per user is a straightforward product preference, not a volume-portability or plugin-architecture decision. If per-widget config or plugin-provided widgets are added later, revisit with `documentation-and-adrs`.
