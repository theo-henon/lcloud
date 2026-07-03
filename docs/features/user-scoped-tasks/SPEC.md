# Spec: User-scoped tasks

> **Status:** Implemented  
> **Scope:** Clarify and complete per-user task ownership — each user sees and manages only their own scheduled tasks; admins see all instance tasks with owner attribution.  
> **Prerequisite:** Task system shipped ([ADR-002](../../adr/002-task-system.md)); user accounts + volume ownership ([user-management-ui](../user-management-ui/SPEC.md), [user-owned-volumes](../user-owned-volumes/SPEC.md)).  
> **Sources:** [VISION.md](../../../VISION.md), [docs/project.md](../../project.md), [design/DESIGN.md](../../../design/DESIGN.md), [ADR-002](../../adr/002-task-system.md)

---

## Assumptions (correct me now or I proceed)

1. **Ownership is per task, not per volume** — `tasks.owner_id` = the user who created the task. A volume can have tasks from different owners only if volume sharing exists later (out of scope v1).
2. **Scheduler runs all enabled tasks** — gocron executes every enabled task in PostgreSQL regardless of viewer; scoping applies to **API + UI visibility**, not background execution.
3. **Global-scope tasks stay admin-only** — macros like `compute_stats` across all volumes remain reserved for `admin`; they are still owned by the creating admin (`owner_id`).
4. **Admin is operator, not impersonator** — admin can view, edit, run, and delete any user's task, but tasks are not reassigned to admin on create.
5. **No task sharing v1** — unlike future shared volumes, users cannot delegate task visibility to another user.
6. **Existing rows are valid** — pre-multi-user tasks created by the bootstrap admin keep `owner_id` = that admin; no migration beyond optional UI clarity.

---

## Problem statement

Operators report that the **Tasks** page feels « global »: every scheduled job appears in one list without clear ownership. That is confusing now that multiple `user` accounts exist.

**Reality today (code audit):**

| Layer | Behaviour |
|---|---|
| **PostgreSQL** | `tasks.owner_id` set on create (`claims.UserID`) |
| **`GET /api/tasks`** | Non-admin: `WHERE owner_id = ?`. Admin: all tasks |
| **CRUD / run / runs** | `authorizeTask`: admin bypass; user must match `owner_id` |
| **Volume binding** | Non-admin can only attach tasks to volumes they own |
| **Global scope** | `403` for non-admin (`ErrGlobalAdminOnly`) |
| **UI** | No owner label; same copy for all roles; `owner_id` in JSON but unused |

**Gap:** permissions are mostly correct server-side, but **product clarity is missing** — especially for admin, who sees every user's tasks mixed together with no « Created by » indicator (same class of issue as volumes before `owner_email`).

---

## Objective

**User stories:**

| ID | Story |
|---|---|
| UST1 | As a `user`, I see only tasks I created. |
| UST2 | As a `user`, I can create volume-scoped tasks only on **my** volumes. |
| UST3 | As a `user`, I cannot create or see global (all-volumes) tasks. |
| UST4 | As an `admin`, I see **all** tasks on the instance, with the owner's email on each row. |
| UST5 | As an `admin`, I can run, edit, or delete any user's task (operator maintenance). |
| UST6 | As a `user`, page copy says « your tasks » — not « all volumes on the server ». |

**Out of scope v1:**

- Filter tasks by owner (admin dropdown) — nice follow-up, not required for first ship
- Reassign task ownership (`PATCH owner_id`)
- Tasks on **shared** volumes (depends on [IDEAS.md — In-instance volume sharing](../../../IDEAS.md))
- Per-user task quotas / rate limits
- Notification when another user's task fails

---

## Permissions

| Action | `user` | `admin` |
|---|---|---|
| `GET /api/tasks` | Own tasks only | All tasks |
| `POST /api/tasks` (volume scope) | Own volumes only | Any volume |
| `POST /api/tasks` (global scope) | **403** | Allowed (maintenance macros) |
| `GET/PATCH/DELETE /api/tasks/:id` | Own task only | Any task |
| `POST /api/tasks/:id/run` | Own task only | Any task |
| `GET /api/tasks/:id/runs` | Own task only | Any task |
| Response field `owner_email` | **Omitted** | Present on list + get |

**Global vs volume scope (unchanged from ADR-002):**

| Scope | Meaning | Who can create |
|---|---|---|
| `volume` | Runs macro on one volume | Owner of that volume (or admin) |
| `global` | Runs macro across all volumes (`rebuild_index`, `compute_stats`, `alert_usage`) | Admin only |

UI label for global tasks: prefer **« Instance-wide »** or **« All volumes (admin) »** instead of bare « All volumes », to avoid sounding like « everyone's shared list ».

---

## API changes

### List / Get enrichment (admin only)

Mirror volume `owner_email` pattern:

```
GET /api/tasks
GET /api/tasks/:id
```

When `claims.Role == admin`, resolve `owner_id` → `users.email` (batch on list via `OwnerEmailsByIDs` or equivalent on `volume.Service` / `auth.Service`).

Response (admin list excerpt):

```json
{
  "tasks": [
    {
      "id": "...",
      "owner_id": "...",
      "owner_email": "user@example.com",
      "name": "Weekly cleanup",
      "scope": "volume",
      "volume_name": "Photos",
      ...
    }
  ]
}
```

Non-admin responses **must not** include `owner_email` (redundant — always self).

### No schema change

`tasks.owner_id` already exists. No ADR required unless we add owner filter query param later (`?owner_id=`).

---

## UI changes

### `/tasks` — TasksPage

| Element | `user` | `admin` |
|---|---|---|
| Header description | « Automate recurring jobs on **your** volumes. » | « Manage scheduled jobs across the instance. » |
| Volume filter dropdown | Only user's volumes (already via `useVolumes`) | All volumes (unchanged) |
| Section title | « Your scheduled tasks » | « All scheduled tasks » |
| Empty state | « No tasks yet — create one on your volumes. » | « No tasks on this instance yet. » |

### TaskList card

For **admin only**, show a muted line under the task name (same pattern as `VolumeCard`):

```
Created by user@example.com
```

Use `owner_email` from API; icon `User` (Lucide), `text-sm text-muted`.

### TaskForm

No change to ownership rules. Keep global macro picker **admin-only** (already gated by `isAdmin`).

---

## Testing strategy

| ID | Test | Package |
|---|---|---|
| T-UST1 | User A creates task; User B `GET /api/tasks` does not return it | `internal/api` |
| T-UST2 | User cannot `POST` task on volume owned by another user | `internal/api` |
| T-UST3 | Admin `GET /api/tasks` includes `owner_email` for another user's task | `internal/api` |
| T-UST4 | User `GET /api/tasks` omits `owner_email` field | `internal/api` |
| T-UST5 | Existing `TestTaskGlobalForbiddenForUser` still passes | `internal/api` |

Frontend: optional snapshot/manual — admin sees owner line; user does not.

**Commands:**

```bash
go test ./internal/task/... ./internal/api/...
cd web && npm run build
```

---

## Tech stack

See [docs/project.md](../../project.md#tech-stack).

---

## Project structure (touch points)

```
internal/task/service.go       ← List / authorizeTask (already scoped)
internal/api/task_handler.go   ← admin owner_email enrichment
internal/api/task_handler_test.go
web/src/lib/api.ts             ← owner_email?: string on TaskRecord
web/src/components/tasks/TaskList.tsx
web/src/pages/TasksPage.tsx
```

---

## Boundaries

- **Always:** enforce ownership in `authorizeTask`; never bypass MacroOps volume checks for non-admin
- **Ask first:** new query params, new macros, scheduler behaviour changes
- **Never:** expose other users' tasks to non-admin; let users create global-scope tasks

---

## Success criteria

- [x] SC-UST1 Non-admin list returns only `owner_id = self`
- [x] SC-UST2 Non-admin cannot CRUD another user's task (`403`)
- [x] SC-UST3 Admin list/get includes `owner_email`
- [x] SC-UST4 TaskList shows « Created by … » for admin only
- [x] SC-UST5 TasksPage copy is role-aware
- [x] SC-UST6 `go test ./internal/api/...` and `cd web && npm run build` pass

---

## Future scope

| Item | Trigger |
|---|---|
| Admin filter « Owner » dropdown on `/tasks` | Many users per instance |
| Tasks on shared volumes (read/write permission) | In-instance volume sharing spec |
| `owner_email` on global search / command palette | Unified search spec |
| Transfer task ownership | Operator workflow request |

---

## Relationship to other docs

- **[ADR-002](../../adr/002-task-system.md)** — scheduler, macros, global scope rules (architecture; this spec is product scoping + UI)
- **[user-owned-volumes](../user-owned-volumes/SPEC.md)** — volume `owner_id` gates which volumes a user can attach tasks to
- **[IDEAS.md — In-instance volume sharing](../../../IDEAS.md)** — when shipped, extend `validateTaskInput` to allow tasks on volumes shared with the user (not only `vol.OwnerID == claims.UserID`)
