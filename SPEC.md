# Spec: Phase 1.4 — Task system

> **Status:** Draft — pending review
> **Scope:** Task scheduler (gocron), macro executor, predefined macro vocabulary, task CRUD API, task history, Tasks UI. **MVP milestone** — last bootstrap phase.
> **Prerequisite:** Phase 1.1 (volumes, files, metadata cache, Bleve indexer), Phase 1.2 (monitoring, `ComputeStats`, stats cache), Phase 1.3 (event bus, `task.executed` / `task.failed` / `volume.alert.usage` event types defined).
> **Sources:** [STARTUP.md](./STARTUP.md#phase-14--task-system), [VISION.md](./VISION.md), [docs/project.md](./docs/project.md)

---

## Assumptions (correct me now or I proceed)

1. **Phases 1.1–1.3 are shipped** — volume CRUD, file upload/delete, monitoring stats, plugin event bus operational.
2. **`internal/task/` is greenfield** — module referenced in docs but not yet created; Phase 1.4 creates it entirely.
3. **gocron v2 is new** — `github.com/go-co-op/gocron/v2` added during `/build`; not yet in `go.mod`.
4. **Cron schedules use UTC** — server timezone is not exposed in MVP; UI displays "UTC" next to cron helpers.
5. **Manual run is in MVP** — UC4.2 allows triggering a task immediately via API/UI (`POST /tasks/:id/run`), in addition to scheduled runs.
6. **Volume-scoped tasks are the primary path** — UC4.1 creates a task bound to one volume; global scope is admin-only and limited to macros that iterate volumes (see Macro catalog).
7. **Macro file ops bypass JWT but stay in `internal/volume/`** — tasks call new internal methods (`MoveFileInternal`, `DeleteFileInternal`) that reuse metadata/index/thumbnail/event logic; no auth claims required because the task record already validated ownership at CRUD time.
8. **`file.moved` is emitted by move macros** — `move_files`, `sort_by_type`, `sort_by_date` emit `file.moved` per affected file; first emission site in the codebase.
9. **`rebuild_index` re-indexes from metadata cache** — `BleveIndexer.Rebuild` clears the index; macro then walks `./cache/metadata/` and re-indexes all records (metadata is source for rebuild, not a fresh disk walk).
10. **`clear_cache` does not wipe Bleve index** — STARTUP specifies thumbnails + metadata cache only; search index remains (use `rebuild_index` separately if needed).
11. **Task history in PostgreSQL** — last run time, status, affected file count, error message; not written to `{volume}/logs/` in MVP (directory exists but stays empty).
12. **`alert_usage` writes to task run history + event bus** — emits `volume.alert.usage`; no separate on-disk alert log in MVP.
13. **Macro vocabulary is closed** — no new macro names without a `/spec` cycle (AGENTS.md).
14. **ADR required before merge** — task model, scheduler wiring, macro executor boundaries documented in `docs/adr/` via `documentation-and-adrs`.
15. **Dry-run mode in MVP** — destructive macros accept `dry_run: true`; preview impact without touching files (see Macro catalog + OQ1).

---

## Objective

Build the task scheduler and connect it to predefined macros operating on volumes — the automation layer described in VISION.md. Completing Phase 1.4 **reaches the lcloud MVP**.

**Who:** Self-hosters who want recurring maintenance (cleanup, organization, index rebuild) without manual intervention.

**User stories:**

| ID | Story |
|---|---|
| UC4.1 | As a user, I create a task "Delete images older than 90 days" on my Photos volume, scheduled every Sunday at 2:00 UTC. |
| UC4.2 | The task runs (scheduled or manually triggered), deletes old files, and the execution appears in task history with the count of deleted files. |

**Out of scope for Phase 1.4:**

- Custom/user-defined macros or scripting language
- Event-triggered tasks (run on `file.uploaded`, etc.) — schedule-only in MVP
- Task templates / marketplace
- Email, webhook, or push notifications for alerts (event bus only; plugins can subscribe)
- Per-file undo / rollback of macro runs
- Changes to `.volume.json` schema or volume directory structure
- Hot-reload of task schedules without restart (gocron updates in-process on CRUD — no restart needed)
- Kubernetes CronJob / external scheduler

---

## Tech Stack

See [docs/project.md — Tech Stack](./docs/project.md#tech-stack).

Phase 1.4 adds:

| Layer | Addition |
|---|---|
| Backend | `github.com/go-co-op/gocron/v2` — in-process cron + interval scheduler |
| Backend | PostgreSQL models: `Task`, `TaskRun` |
| Backend | `internal/task/` — scheduler, macro registry, executor |
| Backend | `internal/volume/macro_ops.go` — internal move/delete used by macros |
| Frontend | Tasks page, task form, run history |

---

## Commands

### Development

```bash
# Add gocron dependency (during /build)
go get github.com/go-co-op/gocron/v2

# Backend tests (task module focus)
go test ./internal/task/... -cover
go test ./internal/volume/... -run Macro -cover
go test ./... -cover

# Frontend tests
cd web && npm run test

# Local stack
docker compose up -d --build
docker compose logs -f app
```

### Verification (Phase 1.4 done)

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"<ADMIN_EMAIL>","password":"<ADMIN_PASSWORD>"}' \
  | jq -r '.access_token')

VOL_ID="<photos-volume-uuid>"

# 1. Create scheduled cleanup task
TASK=$(curl -s -X POST http://localhost:8080/api/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Delete old photos\",
    \"macro\": \"delete_old_files\",
    \"scope\": \"volume\",
    \"volume_id\": \"$VOL_ID\",
    \"parameters\": {\"days\": 90, \"extensions\": [\".jpg\", \".png\", \".webp\"]},
    \"schedule_type\": \"cron\",
    \"schedule\": \"0 2 * * 0\",
    \"enabled\": true
  }")
TASK_ID=$(echo "$TASK" | jq -r '.id')
echo "$TASK" | jq

# 2. Manual run (UC4.2)
curl -s -X POST "http://localhost:8080/api/tasks/$TASK_ID/run" \
  -H "Authorization: Bearer $TOKEN" | jq

# 3. Task history shows affected count
curl -s "http://localhost:8080/api/tasks/$TASK_ID/runs?limit=5" \
  -H "Authorization: Bearer $TOKEN" | jq

# 4. List tasks — next_run_at populated
curl -s http://localhost:8080/api/tasks \
  -H "Authorization: Bearer $TOKEN" | jq

# 5. Disable task
curl -s -X PATCH "http://localhost:8080/api/tasks/$TASK_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}' | jq
```

---

## Project Structure

See [docs/project.md — Project structure](./docs/project.md#project-structure).

Phase 1.4 creates or extends:

```
lcloud/
├── internal/
│   ├── task/
│   │   ├── model.go              ← GORM Task, TaskRun
│   │   ├── macros.go             ← macro name constants + param schemas
│   │   ├── registry.go           ← macro name → executor func
│   │   ├── executor.go             ← MacroExecutor, dispatches by macro type
│   │   ├── scheduler.go          ← gocron wrapper, load/register jobs
│   │   ├── service.go            ← CRUD, run-now, history
│   │   ├── service_test.go
│   │   ├── executor_test.go
│   │   └── scheduler_test.go
│   ├── volume/
│   │   └── macro_ops.go          ← MoveFileInternal, DeleteFileInternal
│   ├── plugin/
│   │   └── events.go             ← FromTaskExecuted, FromTaskFailed, FromVolumeAlertUsage
│   └── api/
│       ├── task_handler.go
│       └── router.go             ← register task routes
├── cmd/server/main.go            ← wire task service, start scheduler, graceful shutdown
├── web/src/
│   ├── pages/TasksPage.tsx       ← replace PlaceholderPage
│   ├── components/tasks/
│   │   ├── TaskList.tsx
│   │   ├── TaskForm.tsx
│   │   ├── TaskRunHistory.tsx
│   │   ├── MacroParameterFields.tsx
│   │   └── ScheduleFields.tsx
│   └── hooks/useTasks.ts
├── docs/
│   └── adr/
│       └── NNN-task-system.md    ← scheduler + macro vocabulary (created at /build)
└── SPEC.md                       ← this file
```

---

## Architecture

### Execution flow

```
gocron trigger (cron | interval)
    → task.Service.executeTask(taskID)
    → MacroExecutor.Run(ctx, task)
    → registry[macro](ctx, deps, volume, params)
    → volume.MacroOps (move/delete) | monitoring.ComputeStats | indexer.Rebuild+reindex
    → TaskRun persisted (status, affected_count, error)
    → EventPublisher: task.executed | task.failed | file.moved | volume.alert.usage
    → Plugin bus (async, existing pattern)
```

### Module boundaries (AGENTS.md)

- **`internal/task/`** owns scheduling, macro dispatch, task CRUD, run history.
- **`internal/volume/`** owns all disk mutations — macros never write to `userdata/` directly from `task/`.
- **`internal/monitoring/`** — `ComputeStats` called by `compute_stats` macro (existing method).
- **`internal/indexer/`** — `Rebuild` + `Index` called by `rebuild_index` macro via `VolumeIndexer` interface only.
- **Event emission** — task service publishes via extended `EventPublisher` or plugin bridge helpers; domain modules do not import `internal/task/`.
- **No business logic in Gin handlers** — handlers call `task.Service` only.

### Concurrency

| Rule | Behavior |
|---|---|
| Same task | At most one run in progress; concurrent trigger (schedule + manual) skips if already running |
| Same volume | Destructive macros (`delete_*`, `move_*`, `sort_*`, `clear_cache`) acquire a per-volume mutex for the whole run — see **OQ2** for collision policy |
| Non-destructive | `rebuild_index`, `compute_stats`, `alert_usage` may run concurrently with each other and **while** a destructive macro holds the volume lock |
| Dry-run runs | A dry-run (`dry_run: true`) **does not** acquire the volume mutex — read-only scan; safe to overlap with other dry-runs |

### Scheduler lifecycle

1. **Startup:** load enabled tasks from PostgreSQL → register gocron jobs → compute `next_run_at`.
2. **CRUD:** create/update/delete task → unregister old job → register new job (in-process, no restart).
3. **Shutdown:** `scheduler.Shutdown()` on SIGTERM before plugin stop.

---

## Data models

### PostgreSQL — `tasks`

```go
type Task struct {
    ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
    OwnerID      uuid.UUID      `gorm:"type:uuid;not null;index"`
    Name         string         `gorm:"not null"`
    Macro        string         `gorm:"not null;index"` // predefined macro name
    Scope        string         `gorm:"not null"`       // volume | global
    VolumeID     *uuid.UUID     `gorm:"type:uuid;index"` // required when scope=volume
    Parameters   JSON           `gorm:"type:jsonb;not null;default:'{}'"`
    ScheduleType string         `gorm:"not null"` // cron | interval
    Schedule     string         `gorm:"not null"` // cron expr or Go duration e.g. "24h"
    Enabled      bool           `gorm:"not null;default:true"`
    NextRunAt    *time.Time     `gorm:"index"`
    LastRunAt    *time.Time
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Validation rules:**

- `macro` must be in the predefined catalog.
- `scope=volume` → `volume_id` required; owner must own volume (or admin).
- `scope=global` → `volume_id` null; **admin-only** create/update; macro must allow global scope.
- `schedule_type=cron` → validate with standard 5-field cron parser (minute hour dom month dow).
- `schedule_type=interval` → parse as Go duration (`1h`, `24h`, `168h`); minimum `1m`.
- `parameters` validated per macro schema before save.

### PostgreSQL — `task_runs`

```go
type TaskRun struct {
    ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
    TaskID        uuid.UUID  `gorm:"type:uuid;not null;index"`
    Status        string     `gorm:"not null"` // success | failed | skipped | dry_run
    AffectedCount int        `gorm:"not null;default:0"`
    Message       string     // human-readable summary
    Error         string     // set when status=failed
    StartedAt     time.Time  `gorm:"not null;index"`
    FinishedAt    time.Time
    DurationMs    int64
}
```

**Retention:** last **200 runs per task** — delete oldest on insert when exceeded (mirrors plugin log policy).

---

## Macro catalog (closed vocabulary)

All macros return `MacroResult{AffectedCount, Message}`.

### Scope matrix

| Macro | Volume | Global (admin) |
|---|---|---|
| `delete_old_files` | ✅ | ❌ |
| `delete_large_files` | ✅ | ❌ |
| `clear_cache` | ✅ | ❌ |
| `move_files` | ✅ | ❌ |
| `sort_by_type` | ✅ | ❌ |
| `sort_by_date` | ✅ | ❌ |
| `rebuild_index` | ✅ | ✅ (all volumes) |
| `compute_stats` | ✅ | ✅ (all volumes) |
| `alert_usage` | ✅ | ✅ (all volumes) |

Global execution iterates volumes the task owner can access (admin: all volumes; user: own volumes only — global tasks admin-only in MVP so this is admin → all volumes).

### Dry-run mode (MVP)

Destructive macros support an optional **`dry_run`** parameter (boolean, default `false`). When `true`:

| Behavior | Detail |
|---|---|
| Disk | **No mutation** — no delete, move, or cache wipe |
| Events | **No emission** — no `file.deleted`, `file.moved`, `task.executed` side-effects on individual files (task-level `task.executed` still fires with status `dry_run`) |
| History | Run status = `dry_run`; `affected_count` = number of files **that would** be affected |
| Message | Human-readable summary, e.g. `"Would delete 12 files older than 90 days"`; optional `preview_paths` (max 20) in run detail API |
| Mutex | Does **not** take the destructive volume lock |

**Macros supporting `dry_run`:** `delete_old_files`, `delete_large_files`, `clear_cache`, `move_files`, `sort_by_type`, `sort_by_date`.

**Not supported:** `rebuild_index`, `compute_stats`, `alert_usage` (non-destructive or read-only already).

**UI:** checkbox "Simulation only (dry-run)" in task form; "Run simulation" button on Run-now dialog when macro is destructive. Scheduled runs honour the task's saved `dry_run` parameter — a scheduled dry-run task never mutates files.

**One-shot override:** `POST /tasks/:id/run?dry_run=true` runs a simulation even if the saved task has `dry_run: false` (useful for preview before enabling a destructive schedule).

### Cleanup

#### `delete_old_files`

Delete files in `./userdata/` whose **modification date** is older than N days.

| Parameter | Type | Required | Default |
|---|---|---|---|
| `days` | int | yes | — |
| `extensions` | []string | no | all files |
| `dry_run` | bool | no | `false` |

- Match against `FileMetadataRecord.ModifiedAt` from metadata cache; fallback to `os.Stat` if metadata missing.
- Uses `DeleteFileInternal` per match (skipped when `dry_run: true`).
- Does not delete directories.

#### `delete_large_files`

Delete files larger than N megabytes.

| Parameter | Type | Required | Default |
|---|---|---|---|
| `min_size_mb` | int | yes | — |
| `extensions` | []string | no | all files |
| `dry_run` | bool | no | `false` |

- Compare `SizeBytes >= min_size_mb * 1024 * 1024`.

#### `clear_cache`

Wipe `./cache/thumbnails/` and `./cache/metadata/` for the volume.

| Parameter | Type | Required | Default |
|---|---|---|---|
| `dry_run` | bool | no | `false` |

- Remove all files in both dirs; recreate empty dirs (skipped when `dry_run: true`).
- Invalidate stats cache (`StatsCache.Invalidate`).
- Does **not** remove `./cache/index/` (Bleve).
- `AffectedCount` = number of metadata records removed.

### Organization

#### `move_files`

Move files matching a glob pattern to a target subfolder under `./userdata/`.

| Parameter | Type | Required | Default |
|---|---|---|---|
| `pattern` | string | yes | — |
| `target_subfolder` | string | yes | — |
| `dry_run` | bool | no | `false` |

- Glob matched against relative path from userdata root (e.g. `*.jpg`, `vacation/*.png`).
- `target_subfolder` sanitized via `PathResolver` (no `..`).
- Creates target directory if missing.
- Skips if destination file already exists (count as skipped, not error).
- Uses `MoveFileInternal`; emits `file.moved` per file.

#### `sort_by_type`

| Parameter | Type | Required | Default |
|---|---|---|---|
| `dry_run` | bool | no | `false` |

Auto-sort files into type-based subfolders at userdata root:

| Folder | MIME rule (reuse `monitoring.CategoryForMIME`) |
|---|---|
| `images/` | `image/*` |
| `documents/` | documents category |
| `videos/` | `video/*` |
| `audio/` | `audio/*` |
| `other/` | everything else |

- Only moves files **directly in userdata root** (not recursive into existing subfolders) — avoids reshuffling organized trees.
- Archives category maps to `other/` (no `archives/` folder in STARTUP list).

#### `sort_by_date`

| Parameter | Type | Required | Default |
|---|---|---|---|
| `dry_run` | bool | no | `false` |

Auto-sort files from userdata root into `YYYY/MM/` subfolders based on **modification date**.

- Same non-recursive rule as `sort_by_type`.
- Example: file modified 2026-03-15 → `2026/03/filename.jpg`.

### Maintenance

#### `rebuild_index`

Force full Bleve reindex.

1. `VolumeIndexer.Rebuild(volumePath)` — clears index files.
2. Walk metadata cache (`ListAll`) → `Index` each record.
3. `AffectedCount` = records re-indexed.

#### `compute_stats`

Force recalculation of volume statistics.

- Calls existing `monitoring.Service.ComputeStats(vol)`.
- `AffectedCount` = file count from computed stats.

### Alerts

#### `alert_usage`

Check if volume usage exceeds threshold; log and emit event when triggered.

| Parameter | Type | Required | Default |
|---|---|---|---|
| `threshold_percent` | int | yes | — |

- Usage = `vol.UsedBytes / vol.QuotaBytes * 100` (skip if `QuotaBytes == 0`).
- When exceeded: emit `volume.alert.usage` with `{usage_percent, quota_bytes, used_bytes}`.
- `AffectedCount` = 1 when alert fired, 0 otherwise.
- Does not block or delete files.

---

## Internal volume operations

New file `internal/volume/macro_ops.go` — used **only** by `internal/task/` (same module, unexported helpers or `MacroOps` struct).

```go
type MacroOps struct {
    volumes      *Service
    paths        *PathResolver
    metadata     *MetadataCache
    thumbnails   *ThumbnailGenerator
    indexManager *indexer.IndexManager
    statsCache   StatsInvalidator
    events       EventPublisher
}

// DeleteFileInternal removes a file by relative path without auth claims.
// Reuses the same steps as FileService.Delete (metadata, thumbnail, index, usage, events).
func (m *MacroOps) DeleteFileInternal(ctx context.Context, vol *Volume, relPath string) error

// MoveFileInternal moves within userdata; updates metadata path, re-indexes, emits file.moved.
func (m *MacroOps) MoveFileInternal(ctx context.Context, vol *Volume, fromRel, toRel string) error
```

**Security:** callers must validate task ownership before invoking. Macro ops trust the task layer.

**Events added to `volume.EventPublisher`:**

```go
FileMoved(ctx context.Context, event FileMovedEvent)
// FileMovedEvent: VolumeID, FromPath, ToPath, SizeBytes
```

Plugin bridge maps to `file.moved` (already defined in `internal/plugin/events.go`).

---

## Event catalog (Phase 1.4 emissions)

| Event | Emitted when | Payload highlights |
|---|---|---|
| `task.executed` | Task run completes with `success` | `task_id`, `macro`, `volume_id`, `affected_count`, `duration_ms` |
| `task.failed` | Task run completes with `failed` | `task_id`, `macro`, `volume_id`, `error` |
| `file.moved` | After each successful `MoveFileInternal` | `volume_id`, `from_path`, `to_path`, `size_bytes` |
| `volume.alert.usage` | `alert_usage` threshold exceeded | `volume_id`, `usage_percent`, `quota_bytes`, `used_bytes` |

Emit via plugin bridge after `TaskRun` persisted — same async fire-and-forget pattern as file events.

---

## API (Phase 1.4)

Base path: `/api`. All endpoints require JWT.

### Tasks

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/tasks` | JWT | List tasks (scoped to owner; admin sees all) |
| `POST` | `/tasks` | JWT | Create task |
| `GET` | `/tasks/:id` | JWT | Task detail + last run summary |
| `PATCH` | `/tasks/:id` | JWT | Update task (owner or admin) |
| `DELETE` | `/tasks/:id` | JWT | Delete task + unregister schedule |
| `POST` | `/tasks/:id/run` | JWT | Trigger immediate run |
| `GET` | `/tasks/:id/runs` | JWT | Execution history |

Optional query on `GET /tasks`: `volume_id` filter.

#### `POST /tasks` — request example

```json
{
  "name": "Delete old photos",
  "macro": "delete_old_files",
  "scope": "volume",
  "volume_id": "660e8400-e29b-41d4-a716-446655440001",
  "parameters": {
    "days": 90,
    "extensions": [".jpg", ".png", ".webp"]
  },
  "schedule_type": "cron",
  "schedule": "0 2 * * 0",
  "enabled": true
}
```

#### `GET /tasks` — response shape

```json
{
  "tasks": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Delete old photos",
      "macro": "delete_old_files",
      "scope": "volume",
      "volume_id": "660e8400-e29b-41d4-a716-446655440001",
      "volume_name": "Photos",
      "parameters": {"days": 90, "extensions": [".jpg", ".png", ".webp"]},
      "schedule_type": "cron",
      "schedule": "0 2 * * 0",
      "schedule_description": "Every Sunday at 02:00 UTC",
      "enabled": true,
      "next_run_at": "2026-07-06T02:00:00Z",
      "last_run_at": "2026-06-29T02:00:01Z",
      "last_run_status": "success",
      "created_at": "2026-06-01T10:00:00Z"
    }
  ]
}
```

#### `GET /tasks/:id/runs`

| Param | Type | Description |
|---|---|---|
| `limit` | int | Max entries (default 20, max 100) |

```json
{
  "runs": [
    {
      "id": "...",
      "status": "success",
      "affected_count": 12,
      "message": "Deleted 12 files older than 90 days",
      "error": "",
      "started_at": "2026-06-29T02:00:01Z",
      "finished_at": "2026-06-29T02:00:03Z",
      "duration_ms": 2100
    }
  ]
}
```

#### `POST /tasks/:id/run`

**Query parameters:**

| Param | Type | Description |
|---|---|---|
| `dry_run` | bool | Optional one-shot override — `true` forces simulation even if task params have `dry_run: false` |

Returns `202 Accepted` with `{ "run_id": "..." }` when execution starts asynchronously.

Poll `GET /tasks/:id/runs` or refresh task detail for result.

**Error codes:**

| Code | Condition |
|---|---|
| `TASK_NOT_FOUND` | Unknown task id |
| `INVALID_MACRO` | Unknown macro name |
| `INVALID_PARAMETERS` | Params fail schema validation |
| `INVALID_SCHEDULE` | Bad cron or interval |
| `VOLUME_REQUIRED` | scope=volume without volume_id |
| `GLOBAL_ADMIN_ONLY` | Non-admin creates global task |
| `VOLUME_NOT_FOUND` | volume_id invalid or not accessible |
| `TASK_ALREADY_RUNNING` | Manual run while execution in progress |
| `TASK_DISABLED` | Manual run on a disabled task |
| `VOLUME_TASK_BUSY` | Volume locked by another destructive macro (if OQ2 = fail fast) |
| `FORBIDDEN` | Accessing another user's task |

---

## Frontend (Phase 1.4)

Follow [design/DESIGN.md](./design/DESIGN.md): dark cards, electric yellow primary actions, status badges.

### Routes

| Path | Access | Content |
|---|---|---|
| `/tasks` | Protected | Task manager (replaces placeholder) |

Remove "Soon" badge from Tasks sidebar item when Phase 1.4 ships.

### `/tasks` — Layout

**Section 1 — Task list (`TaskList`):**

- Table: name, volume (linked), macro label, schedule (human-readable + UTC), enabled toggle, next run, last run status badge.
- Actions: Run now, Edit, Delete.
- Empty state: "No tasks yet — create one to automate volume maintenance."
- Filter by volume (dropdown of user's volumes).

**Section 2 — Create/Edit dialog (`TaskForm`):**

- Fields: name, volume picker (hidden for global macros), macro selector (grouped: Cleanup / Organization / Maintenance / Alerts).
- Dynamic parameter fields per macro (`MacroParameterFields`).
- Schedule: radio cron vs interval; cron helper presets ("Every Sunday 2am UTC", "Daily midnight UTC"); interval dropdown (1h, 6h, 24h, 7d).
- Enabled checkbox (default on).

**Section 3 — Run history (`TaskRunHistory`):**

- Shown inline expanded row or slide-over when selecting a task.
- Columns: started, duration, status, affected count, message/error.
- Success → emerald badge; failed → rose; skipped → muted; dry_run → yellow/muted "Simulation".

### Data fetching

- `useTasks(volumeId?)` — key `['tasks', volumeId]`, staleTime 15s.
- `useTask(id)` — key `['tasks', id]`.
- `useTaskRuns(id)` — key `['tasks', id, 'runs']`.
- `useCreateTask`, `useUpdateTask`, `useDeleteTask`, `useRunTask` — mutations with query invalidation.

### Macro labels (UI copy)

| Macro | Label |
|---|---|
| `delete_old_files` | Delete old files |
| `delete_large_files` | Delete large files |
| `clear_cache` | Clear cache |
| `move_files` | Move matching files |
| `sort_by_type` | Sort by file type |
| `sort_by_date` | Sort by date |
| `rebuild_index` | Rebuild search index |
| `compute_stats` | Recompute statistics |
| `alert_usage` | Usage alert |

---

## Code Style

### Go — macro registration

```go
// internal/task/registry.go
type MacroFunc func(ctx context.Context, exec *Executor, vol *volume.Volume, params map[string]any) (MacroResult, error)

func DefaultRegistry() map[string]MacroFunc {
    return map[string]MacroFunc{
        "delete_old_files":  execDeleteOldFiles,
        "delete_large_files": execDeleteLargeFiles,
        // ...
    }
}
```

- Parameter parsing in dedicated `parseDeleteOldFilesParams(params)` functions — validate before side effects.
- Return partial success: if 3 of 5 deletes fail, status `failed`, `AffectedCount=2`, `Error` summarizes first failure.

### Go — scheduler registration

```go
func (s *Scheduler) Register(task Task) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.unregisterLocked(task.ID)
    if !task.Enabled {
        return nil
    }
    jobFn := func() { s.service.executeTask(context.Background(), task.ID) }
    switch task.ScheduleType {
    case "cron":
        _, err := s.cron.NewJob(gocron.CronJob(task.Schedule, false), gocron.NewTask(jobFn))
        return err
    case "interval":
        d, err := time.ParseDuration(task.Schedule)
        if err != nil { return err }
        _, err = s.cron.NewJob(gocron.DurationJob(d), gocron.NewTask(jobFn))
        return err
    }
    return ErrInvalidSchedule
}
```

### TypeScript

- Macro selector drives conditional form fields — no single mega-form with booleans.
- Cron preset buttons fill the schedule input; advanced users can edit raw cron.
- Confirm dialog before Run now on destructive macros (when `dry_run` is false).
- Dry-run checkbox visible only for destructive macros; helper text: "Aucun fichier ne sera modifié — prévisualisation uniquement."

---

## Testing Strategy

See [docs/project.md — Coverage targets](./docs/project.md#coverage-targets).

| Layer | Focus | Location |
|---|---|---|
| Backend unit | Parameter validation per macro | `internal/task/macros_test.go` |
| Backend unit | Cron/interval parsing, next_run computation | `internal/task/scheduler_test.go` |
| Backend unit | Macro registry dispatches correctly | `internal/task/executor_test.go` |
| Backend integration | `delete_old_files` on temp volume dir | `internal/task/executor_test.go` |
| Backend integration | `MoveFileInternal` updates metadata + index | `internal/volume/macro_ops_test.go` |
| Backend integration | Task API CRUD + authorization | `internal/api/task_handler_test.go` |
| Backend integration | Manual run creates TaskRun, emits events (mock publisher) | `internal/task/service_test.go` |
| Frontend unit | TaskList, TaskForm macro switching, ScheduleFields | `web/src/components/tasks/*.test.tsx` |

**Coverage target:** 80% minimum on `internal/task/`.

**Verify before ship:**

```bash
go test ./internal/task/... -cover
go test ./...
cd web && npm run test
# Manual UC4.1 – UC4.2 via UI + curl script above
```

---

## Boundaries

### Always

- All disk mutations go through `internal/volume/` — macros in `task/` orchestrate, never `os.Remove` on userdata directly.
- `rebuild_index` uses `VolumeIndexer` interface only — never import Bleve from `task/`.
- Task CRUD validates macro name against closed catalog before save.
- Scheduled runs and manual runs share the same `executeTask` path.
- Emit `task.executed` / `task.failed` on every completed run.
- Write ADR in `docs/adr/` before merge.
- Follow [design/DESIGN.md](./design/DESIGN.md) for Tasks UI.
- On task delete, unregister gocron job and cancel in-flight run gracefully if possible.

### Ask first

- Adding macros outside the predefined list.
- Event-triggered tasks (non-schedule triggers).
- Macro ability to delete directories or entire volume trees.
- Changing task history retention policy significantly.
- Exposing server local timezone for cron.

### Never

- User-defined script/code execution in tasks.
- Direct Bleve calls from outside `internal/indexer/`.
- Bypass volume quota checks during macro deletes (usage must stay consistent).
- Store task definitions in `.volume.json` — PostgreSQL only (volume portability unaffected).
- Block plugin event delivery on task failure.
- Change `.volume.json` schema or volume directory structure.

---

## Success Criteria

Phase 1.4 is **done** when all of the following pass — **MVP reached**:

- [ ] **SC4.1** gocron v2 integrated; enabled tasks loaded and scheduled at startup
- [ ] **SC4.2** Task CRUD API with owner scoping; admin sees all tasks
- [ ] **SC4.3** Cron and interval schedule types validated; `next_run_at` exposed in API
- [ ] **SC4.4** All 9 predefined macros implemented with parameter validation
- [ ] **SC4.5** `delete_old_files` deletes by age; respects optional extension filter (UC4.1)
- [ ] **SC4.6** Manual run via `POST /tasks/:id/run` works (UC4.2)
- [ ] **SC4.6b** Dry-run on destructive macro returns status `dry_run`, correct `affected_count`, zero disk changes
- [ ] **SC4.6c** `POST /tasks/:id/run?dry_run=true` one-shot override works
- [ ] **SC4.7** Task history records status, affected count, duration, error message
- [ ] **SC4.8** Enable/disable toggles gocron registration without server restart
- [ ] **SC4.9** `move_files`, `sort_by_type`, `sort_by_date` emit `file.moved` per file
- [ ] **SC4.10** `rebuild_index` rebuilds Bleve from metadata cache
- [ ] **SC4.11** `compute_stats` refreshes monitoring stats cache
- [ ] **SC4.12** `clear_cache` wipes thumbnails + metadata only (not Bleve)
- [ ] **SC4.13** `alert_usage` emits `volume.alert.usage` when threshold exceeded
- [ ] **SC4.14** `task.executed` and `task.failed` emitted on bus after each run
- [ ] **SC4.15** UI: create task on a volume with cron schedule (UC4.1)
- [ ] **SC4.16** UI: run history shows deleted file count after execution (UC4.2)
- [ ] **SC4.17** UI: Tasks sidebar active without "Soon" badge
- [ ] **SC4.18** ADR documents task system and macro vocabulary
- [ ] **SC4.19** `go test ./...` and `cd web && npm run test` pass

---

## Implementation order (preview for `/plan`)

Suggested vertical slices — `/plan` will expand into `tasks/plan.md`:

1. **ADR draft** — task model, gocron choice, macro vocabulary contract
2. **PostgreSQL models** — Task, TaskRun; migrate in main
3. **Macro parameter schemas + validation**
4. **volume.MacroOps** — DeleteFileInternal, MoveFileInternal + tests
5. **EventPublisher extension** — FileMoved, task events in plugin bridge
6. **Macro executor + registry** — implement all 9 macros
7. **Task service** — CRUD, executeTask, history, concurrency locks
8. **gocron scheduler** — register/unregister, startup/shutdown wiring
9. **Task API handlers** — routes in router
10. **main.go wiring** — task service deps, scheduler start
11. **Frontend TasksPage** — list, form, history, run now
12. **Integration** — Docker end-to-end UC4.1–UC4.2

---

## Decisions log

| # | Decision | Status |
|---|---|---|
| D1 | gocron v2 in-process scheduler | ✅ Proposed |
| D2 | Task definitions in PostgreSQL only | ✅ Proposed |
| D3 | Cron in UTC | ✅ Proposed |
| D4 | Internal MacroOps in volume package for disk mutations | ✅ Proposed |
| D5 | `rebuild_index` sources files from metadata cache | ✅ Proposed |
| D6 | `clear_cache` excludes Bleve index | ✅ Proposed |
| D7 | Sort macros only move files at userdata root (non-recursive) | ✅ Proposed |
| D8 | Global scope admin-only; limited macros | ✅ Proposed |
| D9 | Manual run in MVP (async 202) | ✅ Proposed |
| D10 | Task run retention: 200 per task | ✅ Proposed |
| D11 | `file.moved` first emitted by Phase 1.4 move macros | ✅ Proposed |
| D12 | Dry-run mode on destructive macros in MVP | ✅ Resolved (OQ1 — operator choice) |
| D13 | Volume busy: fail fast (OQ2) | ✅ Resolved |
| D14 | Minimum schedule interval: 1 minute (OQ3) | ✅ Resolved |

---

## Open Questions

| # | Question | Status |
|---|---|---|
| OQ1 | **Destructive macro dry-run mode** | ✅ Resolved — included in MVP |
| OQ2 | **Volume busy policy** | ✅ Resolved — fail fast (`VOLUME_TASK_BUSY`, status `skipped`) |
| OQ3 | **Minimum interval for scheduled tasks** | ✅ Resolved — 1 minute minimum |

### OQ1 — Dry-run mode ✅ Resolved

**Decision:** Dry-run **included in MVP** (operator request).

Destructive macros accept `dry_run: true`. See [Dry-run mode (MVP)](#dry-run-mode-mvp) for full behavior.

---

### OQ2 — Volume busy policy

**Contexte — pourquoi cette question existe**

Une macro destructive (suppression, déplacement, vidage cache) lit et modifie des fichiers pendant plusieurs secondes ou minutes. Si deux macros de ce type s'exécutent **en même temps sur le même volume**, elles peuvent se marcher dessus : double suppression, déplacement vers un chemin déjà pris, compteur d'espace disque incohérent.

lcloud pose donc un **verrou par volume** pendant toute la durée d'une exécution destructive réelle (`dry_run: false`). La question est : que fait-on quand une deuxième tâche destructive arrive alors que le verrou est déjà pris ?

**Ce qui ne change pas (quel que soit le choix)**

| Situation | Comportement |
|---|---|
| Deux tâches sur **des volumes différents** | Aucun conflit — exécution en parallèle |
| Tâche destructive + tâche non destructive (`compute_stats`, etc.) sur le **même** volume | Pas de blocage — les macros maintenance/alerte peuvent tourner en parallèle |
| Deux dry-runs sur le même volume | Pas de verrou — lectures seulement |
| Même tâche déclenchée deux fois (cron + clic « Exécuter ») | La deuxième est ignorée — statut `skipped`, message « already running » |

**Scénarios concrets à trancher**

| Scénario | Option A — Échec immédiat | Option B — File d'attente |
|---|---|---|
| Dimanche 2h00 : tâche planifiée « supprimer vieux fichiers » démarre. À 2h01 tu cliques « Exécuter maintenant » sur une autre tâche « tri par date » sur le **même** volume | La 2e run se termine tout de suite avec statut `skipped` et message « volume occupé ». Tu relances manuellement plus tard. | La 2e run **attend** que la 1re finisse, puis s'exécute automatiquement (peut démarrer à 2h05 si la 1re a duré 4 min). |
| Deux tâches planifiées au **même** créneau cron sur le même volume | La 2e est `skipped` ; visible dans l'historique avec horodatage. | La 2e démarre dès que la 1re libère le verrou — ordre d'arrivée respecté. |
| Tâche longue (tri de 10 000 fichiers) + 3e déclenchement pendant l'attente (option B) | N/A | Risque de **chaîne d'attente** : la 3e attend la 2e qui attend la 1re ; une run peut démarrer très tard. |

**Impact utilisateur**

| | Option A — Échec immédiat | Option B — File d'attente |
|---|---|---|
| Prévisibilité | ✅ Tu sais tout de suite que ça n'a pas tourné | ⚠️ Tu peux croire que rien ne se passe alors qu'une run est en attente |
| Simplicité technique | ✅ Pas de goroutine bloquée, pas de timeout à gérer | ⚠️ Faut définir un timeout max d'attente (ex. 30 min) sinon run fantôme |
| Cas d'usage homelab | Rare d'avoir 2 tâches destructives qui se chevauchent | Utile si tu empiles plusieurs automatisations sur un gros volume |

**Recommandation :** Option A — échec immédiat (`VOLUME_TASK_BUSY`, statut `skipped`). Les chevauchements sont rares en homelab ; un échec explicite dans l'historique est plus clair qu'une exécution retardée de plusieurs minutes.

---

### OQ3 — Intervalle minimum entre exécutions planifiées

**Contexte — pourquoi cette question existe**

En plus du cron (ex. « tous les dimanches à 2h »), lcloud accepte des tâches en **intervalle fixe** (ex. « toutes les 24h »). gocron peut théoriquement aller jusqu'à **1 seconde**. Sur du matériel self-hosted (NAS, Raspberry Pi, VPS modeste), un intervalle trop court peut :

- solliciter le disque en continu (`compute_stats`, `rebuild_index`) ;
- rallonger les uploads pendant qu'une macro tourne ;
- consommer CPU/RAM si plusieurs volumes ont chacun une tâche fréquente.

Cette question ne concerne **que** `schedule_type: interval`. Le cron a sa propre granularité (minimum **1 minute** entre deux déclenchements si l'expression le permet — ex. `*/1 * * * *`).

**Macros les plus sensibles à un intervalle court**

| Macro | Charge si intervalle agressif |
|---|---|
| `rebuild_index` | 🔴 Élevée — parcourt tout le metadata cache + réécrit Bleve |
| `compute_stats` | 🟠 Moyenne — lit tous les metadata |
| `delete_old_files` / `sort_*` | 🟠 Variable — dépend du nombre de fichiers |
| `alert_usage` | 🟢 Faible — une lecture de quota |

**Options**

| Option | Minimum autorisé | Conséquence |
|---|---|---|
| **A — 1 minute** | `schedule: "1m"` | Tu peux rafraîchir les stats quasi en temps réel ; doc avertit sur la charge disque |
| **B — 15 minutes** | `schedule: "15m"` | Limite les abus accidentels (« j'ai mis 1m au lieu de 1h ») ; stats moins fraîches |
| **C — 1 heure** | `schedule: "1h"` | Très conservateur ; alertes usage et stats au plus hourly |

**Presets UI proposés (indépendamment du minimum)**

| Preset | Valeur | Usage typique |
|---|---|---|
| Hourly | `1h` | `alert_usage`, `compute_stats` |
| Daily | `24h` | maintenance légère |
| Weekly | `168h` | nettoyage |

Le minimum s'applique si l'utilisateur saisit un intervalle custom (ex. `5m` refusé si minimum = 15m → erreur `INVALID_SCHEDULE` avec message explicite).

**Recommandation :** Option A — **1 minute**. Cible homelab expérimentée ; l'avertissement dans l'UI suffit. Cron et intervalle partagent la même granularité minimale (1 min) pour rester cohérent.

---

## Next step

After spec approval → `/plan` to produce `tasks/plan.md` with ordered, verifiable tasks.

**Note:** Phase 1.4 completes the MVP defined in STARTUP.md. After ship, confirm deletion of STARTUP.md before starting post-MVP feature work.
