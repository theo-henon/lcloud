# Spec: In-app notifications (alert center)

> **Status:** Approved  
> **Scope:** User-facing notification layer in the web UI — persistent inbox per user, bell + panel, optional toast on new alerts. Core subscribers: quota usage alerts, failed tasks, volume deletion requests (phase B of user-owned volumes).  
> **Prerequisite:** Event bus shipped (`internal/plugin/`), task system (ADR-002), user-owned volumes phase A (deletion-request queue).  
> **Sources:** [VISION.md](../../../VISION.md) (Stage 2 — notifications as platform capability), [docs/project.md](../../project.md), [IDEAS.md — In-app notification system](../../../IDEAS.md), [design/DESIGN.md](../../../design/DESIGN.md), [user-owned-volumes](../user-owned-volumes/SPEC.md), [dashboard-configurable](../dashboard-configurable/SPEC.md), [ADR-002](../../adr/002-task-system.md)

---

## Assumptions (correct me now or I proceed)

1. **PostgreSQL inbox, not volume-portable** — notifications are instance-level UX state tied to a user account. They do not live in `.volume.json` or on disk under a volume. Re-importing a volume does not restore notification history (acceptable — same class as dashboard layout).
2. **Web UI only in v1** — no email, push, SMS, or webhook delivery (those stay plugin / [IDEAS — Webhook notifications macro](../../../IDEAS.md) territory).
3. **Polling, not WebSocket/SSE in v1** — frontend polls `GET /api/notifications/unread-count` on an interval (e.g. 30 s) and on window focus. Real-time push deferred.
4. **Built-in event sources only in v1** — map a closed set of bus events to notifications. Plugin custom events → notifications is a follow-up (requires recipient policy).
5. **Recipient = one user per notification** — no shared/group inbox. Admin-targeted alerts are rows where `user_id` = admin user(s).
6. **Deletion-request alerts go to all admins** — when a `user` submits a volume deletion request, every `admin` account gets a notification (small instance assumption: 1–3 admins).
7. **Task failure alerts go to task owner only** — consistent with [user-scoped-tasks](../user-scoped-tasks/SPEC.md) (admin is not auto-notified when another user's task fails).
8. **Usage alerts go to volume owner** — `volume.alert.usage` notifies the volume's `owner_id`. Admins who do not own the volume are not notified unless they are the owner.
9. **No toast for historical items on login** — toasts fire only for notifications created **after** the session started (tracked client-side by `lastSeenAt` or comparing `created_at` to session start). Inbox still shows full history.
10. **Dismiss = soft-delete** — `DELETE /api/notifications/:id` sets `dismissed_at`; dismissed items hidden from default list. No hard purge UI in v1 (background retention job optional).
11. **English product copy** — titles and bodies in English (project language), same as rest of UI.
12. **Bell lives in the sidebar chrome** — placed in the sidebar header area (control panel), visible on every authenticated page. No separate `/notifications` page in v1 (panel-only).

---

## Problem statement

lcloud emits meaningful events on the internal bus — `volume.alert.usage`, `task.failed`, volume lifecycle — but **nothing surfaces them in the UI** except task run history (buried per task) and the admin deletion-request panel on `/volumes`.

**Gap:** Operators miss quota alerts and failed automations unless they actively check Tasks or Monitoring. User-owned volumes phase A added a deletion queue without a proactive signal — admins must open Volumes to notice pending requests.

**Today (code audit):**

| Signal | Emitted | User-visible today |
|---|---|---|
| Quota threshold exceeded | `volume.alert.usage` via `alert_usage` macro | Task run message only |
| Task run failed | `task.failed` | Task run history on Tasks page |
| User requests volume delete | DB row in `volume_deletion_requests` | Admin panel on Volumes + dashboard widget |
| Task succeeded | `task.executed` | Not needed in v1 (noise) |
| File ops | `file.*` | Not needed in v1 (noise) |

---

## Objective

**Who:** Any authenticated user — each sees **their own** inbox. Admin-only notification types are still stored per admin user row.

**Why:** Close the loop between automation/events and human attention without requiring a custom plugin or digging through logs.

**User stories:**

| ID | Story |
|---|---|
| N1 | As a volume owner, when my volume crosses the usage threshold, I receive a notification I can open from the bell. |
| N2 | As a task owner, when my scheduled task fails, I receive a notification with a link to that task. |
| N3 | As an admin, when a user requests volume deletion, I receive a notification linking to the volumes page / request queue. |
| N4 | As a user, I see an unread badge on the bell and can mark items read individually or all at once. |
| N5 | As a user, I can dismiss a notification I no longer care about. |
| N6 | As a user, when a **new** notification arrives during my session, I see a brief toast (non-blocking). |
| N7 | As a user, notifications I don't have access to anymore (e.g. volume deleted) still display safely with stale context — no 500 on open. |

**Out of scope (v1):**

- Email, push, SMS, Slack/Discord
- Webhook macro ([IDEAS.md](../../../IDEAS.md))
- Plugin-emitted custom events → notifications (recipient rules undefined)
- Dedicated `/notifications` full-page history with filters
- Per-user notification preferences / mute per type
- Notify admin when **another user's** task fails
- `task.executed` success toasts
- File upload/delete/move notifications
- Activity / alerts dashboard widget ([dashboard-configurable](../dashboard-configurable/SPEC.md) — depends on this feature)
- Mobile PWA push

### Future scope (explicitly not v1)

- SSE or WebSocket for instant delivery
- Plugin SDK helper `NotifyUser(userID, title, body, link)`
- Notification preferences (mute `volume.alert.usage`, digest mode)
- `activity-alerts` dashboard widget
- Cross-user admin feed (« all failed tasks on instance »)
- i18n of notification templates

---

## Phasing

| Slice | Delivers |
|---|---|
| **Slice 1 — Domain + persistence** | `notifications` table; `internal/notification/` service; GORM model; retention helper |
| **Slice 2 — Event subscribers** | Bus handlers for `volume.alert.usage`, `task.failed`; hook on volume deletion-request create |
| **Slice 3 — REST API** | List, unread count, mark read, mark all read, dismiss |
| **Slice 4 — UI shell** | Sidebar bell + dropdown panel; React Query hooks; unread badge |
| **Slice 5 — Toasts** | Session-scoped toast on new items; deep links from panel rows |

Slices 1→2→3 are sequential. Slice 4 can start once API exists; Slice 5 after polling works.

---

## Data model

PostgreSQL table `notifications`:

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `user_id` | UUID FK → users | Recipient |
| `type` | varchar(64) | Stable machine type (see below) |
| `title` | varchar(200) | Short headline |
| `body` | text | One or two sentences |
| `link_path` | varchar(500) nullable | In-app route, e.g. `/volumes`, `/tasks?task=:id` |
| `volume_id` | UUID nullable | Context |
| `task_id` | UUID nullable | Context |
| `source_event_id` | varchar(64) nullable | Bus event `id` for idempotency |
| `read_at` | timestamptz nullable | |
| `dismissed_at` | timestamptz nullable | Soft delete |
| `created_at` | timestamptz | |

**Indexes:** `(user_id, dismissed_at, created_at DESC)`, `(user_id, read_at)` where dismissed null.

**Retention (v1):** background job or lazy trim — keep last **200** notifications per user OR delete dismissed rows older than **90 days** (whichever is simpler to implement; prefer **200 cap** mirroring `task_runs`).

### Notification types (v1)

| `type` | Trigger | Recipient | Title template | `link_path` |
|---|---|---|---|---|
| `volume.usage_alert` | `volume.alert.usage` | Volume `owner_id` | « {volumeName} quota alert » | `/volumes/{volumeId}` |
| `task.failed` | `task.failed` | Task `owner_id` | « Task failed: {taskName} » | `/tasks` (select task client-side or `?highlight={taskId}`) |
| `volume.deletion_requested` | `POST …/deletion-request` | Each admin | « Deletion requested: {volumeName} » | `/volumes` |

**Body examples:**

- Usage: « Storage is at 92% of quota (4.6 GB / 5 GB). »
- Task failed: « Macro `delete_old_files` failed: {error} »
- Deletion: « {userEmail} requested deletion of volume « {volumeName} ». »

Volume/task names resolved at creation time from DB; store denormalized strings in `title`/`body` so rows stay readable if entity is renamed later.

---

## Event → notification mapping

Subscribe inside `internal/notification/` (or wire from `cmd/server/main.go` after plugin bus is created):

```
plugin.EventVolumeAlertUsage  → create volume.usage_alert for owner
plugin.EventTaskFailed        → create task.failed for task owner
(volume deletion request)     → create volume.deletion_requested for each admin
```

**Idempotency:** if `source_event_id` already exists for `(user_id, type)`, skip insert (protects against duplicate bus delivery or restarts).

**Deduplication (usage alerts):** optional v1 enhancement — if an unread `volume.usage_alert` for the same `volume_id` exists with `created_at` within **24 h**, update `body`/`created_at` instead of inserting (reduces spam on recurring `alert_usage` task). **Default if OQ unresolved:** insert every time (simpler).

**Access control on create:**

- Resolve `owner_id` from `volumes` table; skip if volume gone.
- Resolve task owner from `tasks`; skip if task gone.
- Admins: query `users WHERE role = admin`.

Do **not** subscribe to `*` for plugins — only explicit types above.

---

## API

Base: `/api/notifications` — JWT required. Handlers in `internal/api/notification_handler.go`; logic in `internal/notification/service.go`.

| Method | Path | Query / body | Response | Notes |
|---|---|---|---|---|
| `GET` | `/notifications` | `?limit=30&offset=0` | `{ notifications: Notification[], total: number }` | Excludes dismissed; newest first |
| `GET` | `/notifications/unread-count` | — | `{ count: number }` | For badge |
| `PATCH` | `/notifications/:id/read` | — | `200` notification | 404 if not recipient |
| `POST` | `/notifications/read-all` | — | `{ updated: number }` | Mark all unread for current user |
| `DELETE` | `/notifications/:id` | — | `204` | Dismiss (soft) |

**Authorization:** every query scoped to `claims.UserID`. No admin override to read others' inboxes.

### `Notification` JSON

```json
{
  "id": "uuid",
  "type": "volume.usage_alert",
  "title": "Photos quota alert",
  "body": "Storage is at 92% of quota (4.6 GB / 5 GB).",
  "link_path": "/volumes/uuid",
  "volume_id": "uuid",
  "task_id": null,
  "read_at": null,
  "created_at": "2026-07-03T14:00:00Z"
}
```

---

## Tech Stack

See [docs/project.md — Tech Stack](../../project.md#tech-stack).

| Layer | Change |
|---|---|
| Backend | `internal/notification/model.go`, `service.go`, `subscribers.go` |
| Backend | `internal/api/notification_handler.go` |
| Backend | `internal/api/router.go` — notification routes |
| Backend | `internal/volume/service.go` — emit notification hook on deletion request (or subscriber on service event) |
| Backend | `cmd/server/main.go` — AutoMigrate `Notification`; register bus subscribers |
| Frontend | `web/src/components/notifications/NotificationBell.tsx`, `NotificationPanel.tsx` |
| Frontend | `web/src/hooks/useNotifications.ts` |
| Frontend | `web/src/lib/api.ts` — notification client methods |
| Frontend | `web/src/components/layout/Sidebar.tsx` — mount bell |
| Frontend | Toast — custom Zustand store + `ToastHost` (no npm dependency) |

**Dependencies:** No new Go modules. No new frontend npm dependencies for toasts (custom implementation).

---

## Commands

### Development

```bash
# Backend
go test ./internal/notification/...
go test ./internal/api/... -run Notification

# Frontend
cd web && npm run test
cd web && npm run build

# Smoke
make embed
docker compose up -d --build
# → trigger alert_usage / failed task / deletion request; bell badge updates
```

### Lint / typecheck

```bash
go test ./...
cd web && npm run build
```

---

## UI specification

Read [design/DESIGN.md](../../../design/DESIGN.md) before implementation.

### Placement

- **Bell button** in sidebar header block (below « Control panel », above `GlobalSearch`), right-aligned or full-width row with icon + badge.
- Icon: `Bell` from lucide (consistent with [control-panel-iconography](../control-panel-iconography/SPEC.md)).
- **Unread badge:** small pill using `primary` background, `on-primary` text; cap display at `9+`.

### Panel (dropdown)

- Use existing `DropdownMenu` or `Popover` from shadcn patterns.
- Width ~360px; max-height ~420px scrollable.
- Header row: « Notifications » + link-style **Mark all read**.
- Row: title (semibold), body (muted, 2-line clamp), relative time (`created_at`).
- Unread rows: subtle `surface-elevated` background or hairline left accent in `primary`.
- Click row: navigate to `link_path`, mark read.
- Row action: dismiss (icon button, `X` or `Trash2` muted — dismiss is not destructive delete; prefer `X`).
- Empty state: « No notifications » + muted helper text.
- Footer optional: « Dismissed items are hidden » (muted caption).

### Toast

- Bottom-right stack, dark `surface-elevated` card, `title` + truncated `body`.
- Auto-dismiss ~5 s; click navigates + marks read.
- Severity: usage alert → `warning` accent border; task failed → `error` accent; deletion request → `primary` accent.

### Deep links

| Type | Navigate |
|---|---|
| `volume.usage_alert` | `/volumes/:id` |
| `task.failed` | `/tasks` with task pre-selected (extend Tasks page to read `?task=` query or use existing selection state) |
| `volume.deletion_requested` | `/volumes` (admin panel visible) |

---

## Testing strategy

### Backend (Go + testify)

| Test | Covers |
|---|---|
| Subscriber creates usage alert for volume owner | N1 |
| Subscriber creates task.failed for task owner | N2 |
| Deletion request notifies all admins | N3 |
| Idempotent on duplicate `source_event_id` | dedupe |
| List/count scoped to user; other user's rows 404 on patch/delete | auth |
| Dismiss hides from list; unread count decreases | N5 |

### Frontend (Vitest + Testing Library)

| Test | Covers |
|---|---|
| Bell shows badge when unread count > 0 | N4 |
| Mark all read clears badge | N4 |
| Panel renders empty state | UI |

### Manual smoke

1. Create `alert_usage` task; exceed threshold → owner sees notification.
2. Force task failure → owner sees notification + toast.
3. User requests volume delete → admin bell increments.
4. Dismiss + mark read persist after refresh.

---

## Code style

Follow existing module boundaries:

```go
// internal/notification/service.go — pattern
func (s *Service) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Notification, int64, error)
func (s *Service) CreateFromUsageAlert(ctx context.Context, event plugin.Event) error
```

Handlers stay thin:

```go
func (h *NotificationHandler) List(c *gin.Context) {
    claims := auth.MustClaims(c)
    items, total, err := h.notifications.List(c.Request.Context(), claims.UserID, limit, offset)
    // ...
}
```

Frontend: TanStack Query keys `["notifications"]`, `["notifications", "unread-count"]`; invalidate on mark read / dismiss.

---

## Boundaries

- **Always:** Scope all notification queries to authenticated user; validate UUIDs; run `go test` + `npm run build` before PR.
- **Ask first:** New npm dependency for toasts; changing retention cap; notifying admins on all task failures; plugin notification API.
- **Never:** Store notifications in `.volume.json`; bypass event bus for external integrations in this feature; expose other users' notifications to non-admins.

---

## Success criteria

- [x] SC-N1 `volume.alert.usage` creates a notification for the volume owner with correct link
- [x] SC-N2 `task.failed` creates a notification for the task owner
- [x] SC-N3 Volume deletion request creates one notification per admin
- [x] SC-N4 Bell badge reflects unread count; mark read / mark all read work
- [x] SC-N5 Dismiss removes item from list
- [x] SC-N6 New notification during session shows toast (not on initial history load)
- [x] SC-N7 `go test ./internal/...` and `cd web && npm run build` pass

---

## Open questions

| # | Question | Default if unanswered |
|---|---|---|
| OQ1 | Usage alert dedupe within 24 h (update vs new row)? | **Insert every time** |
| OQ2 | Should `task.failed` include dry-run / skipped runs? | **No** — only `task.failed` bus event (real failures) |
| OQ3 | Toast implementation: add `sonner` vs minimal custom component? | **Resolved — minimal custom** (Zustand + `ToastHost`, no new dep) |
| OQ4 | Poll interval for unread count? | **30 s** + refetch on window focus |
| OQ5 | Notify all admins or only « primary » admin? | **Resolved — all admins by default**, configurable via instance setting `notify_admins_on_deletion_request` (toggle in admin Settings) |

---

## Relationship to other features

| Feature | Relationship |
|---|---|
| [user-owned-volumes](../user-owned-volumes/SPEC.md) | Phase B — deletion-request notifications (this spec) |
| [dashboard-configurable](../dashboard-configurable/SPEC.md) | Future `activity-alerts` widget consumes same API |
| [user-scoped-tasks](../user-scoped-tasks/SPEC.md) | Task failure recipient = task owner |
| Monitoring visual upgrade | Complementary — colored quota bars + notifications |
| Webhook notifications macro | External channel; out of scope here |

---

## ADR needed?

**No ADR required for v1** — PostgreSQL per-user inbox is standard product persistence, not a volume-portability or plugin-architecture contract change. If v2 adds plugin `NotifyUser` API or changes event bus semantics, revisit with `documentation-and-adrs`.

---

## Project structure (this feature)

```
internal/notification/     ← new domain module
internal/api/notification_handler.go
web/src/components/notifications/
web/src/hooks/useNotifications.ts
docs/features/notifications/SPEC.md   ← this file
```

See [docs/project.md — Project structure](../../project.md#project-structure) for the full tree.
