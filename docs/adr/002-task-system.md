# ADR-002: Task system architecture

## Status

Accepted

## Date

2026-07-02

## Context

Phase 1.4 delivers the automation layer described in VISION.md: scheduled operations (cron or fixed interval) that execute predefined macros against volumes. This completes the lcloud MVP.

Requirements:

- Tasks persisted in PostgreSQL; volume config stays in `.volume.json` (not task definitions)
- In-process scheduler (gocron v2) — no external CronJob dependency
- Closed macro vocabulary for MVP — changes require a `/spec` cycle
- All disk mutations go through `internal/volume/` — task module orchestrates only
- Dry-run mode on destructive macros — preview without mutation
- Event bus integration: `task.executed`, `task.failed`, `file.moved`, `volume.alert.usage`

## Decision

### Scheduler: gocron v2 in-process

Use `github.com/go-co-op/gocron/v2`. Tasks loaded from PostgreSQL at startup; CRUD re-registers jobs without server restart. Cron expressions interpreted in **UTC**. Minimum schedule granularity: **1 minute** (cron and interval).

### Persistence

- `tasks` table: definition, schedule, macro, parameters (JSONB), scope (volume | global)
- `task_runs` table: execution history (200 runs max per task, purge on insert)

Task definitions are **not** written to `.volume.json`.

### Macro executor

`internal/task/` owns scheduling, CRUD, and macro dispatch. A closed registry maps macro names to executor functions. Parameters validated per macro before save and before run.

### Disk mutations: MacroOps in volume package

`internal/volume/macro_ops.go` provides `DeleteFileInternal` and `MoveFileInternal` — same metadata/index/thumbnail/usage/event logic as user-facing file ops, without JWT claims. The task layer validates ownership at CRUD time; MacroOps trust the caller.

### Dry-run

Destructive macros accept `dry_run: true` in parameters. When set:

- No disk mutation, no per-file events (`file.deleted`, `file.moved`)
- Run status `dry_run`; `affected_count` = would-be count
- Does not acquire the destructive volume mutex

One-shot override: `POST /tasks/:id/run?dry_run=true`.

### Concurrency

| Rule | Behavior |
|---|---|
| Same task | At most one run; concurrent trigger → status `skipped` |
| Same volume, destructive, non-dry-run | Per-volume mutex; second task → `skipped`, `VOLUME_TASK_BUSY` (fail fast) |
| Non-destructive macros | May run concurrently with destructive ops on same volume |
| Dry-run | No volume mutex |

### Event emission

Task service publishes via plugin bridge helpers:

- `task.executed` / `task.failed` after every run
- `file.moved` from `MoveFileInternal` (move macros)
- `volume.alert.usage` from `alert_usage` macro

### Global scope

Admin-only. Limited to `rebuild_index`, `compute_stats`, `alert_usage` — iterates all volumes.

## Macro vocabulary (closed for MVP)

| Category | Macros |
|---|---|
| Cleanup | `delete_old_files`, `delete_large_files`, `clear_cache` |
| Organization | `move_files`, `sort_by_type`, `sort_by_date` |
| Maintenance | `rebuild_index`, `compute_stats` |
| Alerts | `alert_usage` |

## Alternatives considered

### External cron (system crontab, K8s CronJob)

- Rejected: violates "single Docker Compose" deployment model; adds operator burden

### Event-triggered tasks (on `file.uploaded`)

- Rejected for MVP; schedule-only keeps scope bounded

### Task definitions in `.volume.json`

- Rejected: breaks separation between portable volume data and instance automation config

### Queue/wait on volume busy

- Rejected: fail-fast chosen for predictability in homelab use (see SPEC OQ2)

## Consequences

- Phase 1.3 deferred events (`file.moved`, `task.executed`, `volume.alert.usage`) are wired in Phase 1.4
- `rebuild_index` must walk metadata cache after `BleveIndexer.Rebuild` — Rebuild alone clears without re-indexing
- `clear_cache` wipes thumbnails + metadata only, not Bleve index
- New macros require ADR + `/spec` cycle
- MVP completion triggers STARTUP.md deletion (operator confirmation)

## Security rules

- Task CRUD scoped to owner; admin sees all
- Global tasks admin-only
- Macro ops never bypass quota accounting on deletes
- No user-defined script execution
