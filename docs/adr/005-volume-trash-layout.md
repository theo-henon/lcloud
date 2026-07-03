# ADR-005: Volume trash directory layout and event semantics

## Status

Accepted

## Date

2026-07-03

## Context

The volume explorer permanently deletes files on `DELETE /files`. Homelab users expect a recoverable trash step (Google Drive, Dropbox, Nextcloud). The corbeille spec requires:

- Per-volume `.trash/` directory (portable, disk-first)
- Soft-delete preserves quota until permanent purge
- Distinct plugin events: `file.trashed`, `file.restored`, `file.deleted` (permanent only)
- Task macros `delete_old_files` / `delete_large_files` must soft-delete to trash (OQ6 resolved)
- New `purge_trash` macro for permanent purge of aged trash items

## Decision

### On-disk layout

New top-level directory per volume (sibling of `userdata/`, `cache/`, `plugins/`):

```
{volume-root}/
└── .trash/
    └── {file-id}/
        ├── {original-basename}   ← blob bytes
        └── meta.json             ← trash metadata (id, name, original_path, mime_type, size_bytes, sha256, deleted_at, modified_at)
```

- `{file-id}` matches pre-trash metadata cache ID when available
- `original_path` is relative to `userdata/`
- `.trash/` is never exposed via `GET /files?path=`
- New volumes get `.trash/` at creation; existing volumes get lazy `MkdirAll` on first soft-delete
- Re-import preserves trash contents without PostgreSQL

### Quota accounting

| Operation | `used_bytes` |
|---|---|
| Soft-delete (trash) | Unchanged |
| Restore | Unchanged |
| Permanent purge | Decremented by item `size_bytes` |

### Event bus breaking change

| User action (before) | Event (before) | Event (after) |
|---|---|---|
| Explorer delete | `file.deleted` | `file.trashed` |
| Restore from trash | — | `file.restored` |
| Permanent purge / empty trash / sweeper / `purge_trash` | `file.deleted` | `file.deleted` (unchanged) |

Plugins subscribing to `file.deleted` for user deletes must migrate to `file.trashed`.

### Task macro semantic change (breaking for operators)

| Macro | Before | After |
|---|---|---|
| `delete_old_files` | Permanent delete in `userdata/` | Soft-delete to `.trash/` |
| `delete_large_files` | Permanent delete in `userdata/` | Soft-delete to `.trash/` |
| `purge_trash` (new) | — | Permanent purge of trash items older than `days` (dry-run supported) |

Operators who relied on `delete_old_files` for immediate space reclamation should schedule `purge_trash` (or rely on the instance sweeper via `LCLOUD_TRASH_RETENTION_DAYS`).

### Thumbnails

Thumbnail cache entries keyed by file ID are preserved on soft-delete and removed on permanent purge. Trash UI serves thumbnails via `GET /files/thumbnail?id={file-id}`.

### Restore conflict

If `original_path` is occupied → `409 PATH_OCCUPIED` (no auto-rename in v1).

## Consequences

- Positive: Accidental deletes recoverable; volume remains portable
- Positive: Quota behaviour matches major cloud drives
- Negative: `delete_old_files` no longer frees disk space immediately — operators must use `purge_trash` or wait for sweeper
- Negative: Plugin `file.deleted` subscribers need migration for user-delete flows
- Migration: No DB schema change; lazy `.trash/` creation for existing volumes

## Alternatives considered

| Alternative | Rejected because |
|---|---|
| PostgreSQL tombstones only | Breaks portability contract |
| Global cross-volume trash | Out of scope; per-volume matches volume ownership model |
| Keep cleanup macros as permanent bypass | OQ6 resolved — inconsistent UX and surprise data loss |
| Auto-rename on restore conflict | Deferred; explicit 409 keeps behaviour predictable |
