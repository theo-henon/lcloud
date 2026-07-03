# Spec: Corbeille (Trash / soft-delete)

> **Status:** Draft — pending review  
> **Scope:** Replace permanent file deletion with a recoverable trash flow per volume. User deletes move files to a portable on-disk trash folder; users can restore, permanently delete, or empty trash. Auto-purge after a configurable retention period.  
> **Prerequisite:** Volume explorer shipped (`GET/PATCH/DELETE /api/volumes/:id/files`), file metadata cache, Bleve indexer, task system.  
> **Sources:** [VISION.md](../../../VISION.md) (Principle 1 — ownership & portability), [docs/project.md](../../project.md), [IDEAS.md — Trash / recycle bin](../../../IDEAS.md), [design/DESIGN.md](../../../design/DESIGN.md), explorer spec ([SPEC.md](../../../SPEC.md) — trash explicitly deferred)

---

## Assumptions (correct me now or I proceed)

1. **Per-volume trash, not global** — each volume has its own `.trash/` directory. No cross-volume trash bucket.
2. **Disk-first, portable storage** — trashed blobs and metadata live under `{volume}/.trash/`. Re-importing a volume by path preserves trash contents without PostgreSQL.
3. **Files only in v1** — soft-delete applies to **files** in `userdata/`, consistent with today's explorer delete (folders not trashed). Directory delete remains out of scope.
4. **Trashed files still count toward quota** — `used_bytes` is unchanged on soft-delete; quota is freed only on permanent purge (matches Google Drive / Dropbox behaviour).
5. **Default retention: 30 days** — items older than retention are permanently deleted by a built-in background sweeper (not a user-created task in v1).
6. **Retention is instance-wide in v1** — single env var `LCLOUD_TRASH_RETENTION_DAYS` (default `30`). Per-volume override deferred (would touch `.volume.json`).
7. **`DELETE /files` becomes soft-delete** — the existing endpoint moves to trash instead of permanent removal. No separate « move to trash » endpoint in v1.
8. **Permanent delete is explicit** — only via trash APIs (`DELETE /trash/items/:id`, `DELETE /trash` empty) or auto-purge. Emits `file.deleted` (today's semantics).
9. **New bus events** — `file.trashed` on soft-delete, `file.restored` on restore. `file.deleted` reserved for permanent removal.
10. **Index & search** — trashed files are removed from Bleve; they do not appear in volume search or normal file listing. Restore re-indexes.
11. **Thumbnails preserved on trash** — thumbnail cache entries keyed by file ID are kept until permanent purge (faster trash UI, simpler restore).
12. **Task cleanup macros soft-delete to trash** — `delete_old_files` / `delete_large_files` move matching files to `.trash/` (same as user delete). Permanent purge is via `purge_trash` macro, trash APIs, or the instance sweeper.
13. **UI label « Trash »** — English product copy (project language). French operator docs may say « Corbeille ».
14. **No bulk trash actions in v1** — single-file delete/restore/permanent-delete only; bulk selection is a separate IDEAS entry.

---

## Problem statement

Today, deleting a file in the volume explorer is **immediate and irreversible**:

```310:314:web/src/components/volumes/explorer/VolumeExplorer.tsx
                  onDelete={(entry) => {
                    if (window.confirm("Delete this file?")) {
                      void deleteFile.mutateAsync(entry.path);
                    }
                  }}
```

Backend `FileService.Delete` removes the blob, metadata, index entry, and thumbnail, then emits `file.deleted`:

```444:489:internal/volume/file_service.go
func (s *FileService) Delete(claims *auth.Claims, volumeID uuid.UUID, relPath string) error {
	// ... os.Remove(absPath), metadata/index/thumbnail cleanup, file.deleted event
}
```

**Gap:** Every major cloud (OneDrive, Google Drive, Dropbox, Nextcloud) uses a trash step first. Accidental deletes are unrecoverable — a table-stakes gap for homelab users who expect undo.

---

## Objective

**Who:** Any authenticated user with access to a volume (`admin` or volume owner).

**Why:** Recover from accidental deletes; align with cloud-drive mental model without sacrificing data ownership or volume portability.

**User stories:**

| ID | Story |
|---|---|
| TR1 | As a user, when I delete a file from the explorer, it moves to Trash instead of being permanently removed. |
| TR2 | As a user, I can open a Trash view for a volume and see deleted files with original location and deletion date. |
| TR3 | As a user, I can restore a trashed file to its original path (or a safe fallback if that path is occupied). |
| TR4 | As a user, I can permanently delete a single item from Trash after confirmation. |
| TR5 | As a user, I can empty all of Trash after confirmation. |
| TR6 | As a user, trashed files no longer appear in normal folder listings or search. |
| TR7 | As a user, trashed files still count toward my volume quota until permanently removed. |
| TR8 | As a user, items in Trash older than the retention period are automatically purged. |
| TR9 | As a plugin author, I receive `file.trashed` / `file.restored` / `file.deleted` (permanent) as distinct events. |

**Out of scope (v1):**

- Directory / folder soft-delete
- Bulk select → move to trash / restore / purge
- Per-volume retention in `.volume.json`
- Trash widget on dashboard
- WebDAV/FTP trash semantics (protocol layers may still hard-delete until a follow-up spec)
- Version history (separate IDEAS entry)
- Admin trash audit log beyond existing volume access controls
- Plugin hooks to intercept or block trash operations

### Future scope (explicitly not v1)

- Per-volume `trash_retention_days` in `.volume.json`
- Bulk trash operations (depends on multi-select IDEAS entry)
- Trash-aware `delete_old_files` (skip or include trashed)
- Cross-volume « global trash » admin view
- Restore-to-custom-path picker in UI

---

## Phasing

| Slice | Delivers |
|---|---|
| **Slice 1 — ADR + on-disk layout** | ADR-005; `.trash/` structure; `TrashDir` constant; volume layout bootstrap for new volumes; lazy mkdir on first trash |
| **Slice 2 — Backend trash service** | Soft-delete refactor of `FileService.Delete`; `TrashService` (list, restore, purge one, empty); quota/index/metadata rules |
| **Slice 3 — API + events** | REST endpoints; `file.trashed`, `file.restored`; plugin bridge + SDK constants; `file.deleted` only on permanent ops |
| **Slice 3b — Task macro `purge_trash`** | New `purge_trash` macro (days + dry-run); refactor cleanup macros to soft-delete; ADR-002 macro table update |
| **Slice 4 — Auto-purge** | Background sweeper (startup + periodic); `LCLOUD_TRASH_RETENTION_DAYS` config |
| **Slice 5 — Explorer UI** | Trash sidebar entry; trash list view; restore / delete forever / empty trash; updated delete confirm copy |

Slices 2–4 can partially overlap; Slice 1 must land first. UI (Slice 5) depends on API (Slice 3).

---

## On-disk layout (ADR-005)

New top-level directory per volume (sibling of `userdata/`, `cache/`, `plugins/`):

```
{volume-root}/
├── .volume.json
├── userdata/
├── cache/
├── plugins/
├── logs/
└── .trash/
    └── {file-id}/
        ├── blob                  ← original file bytes (same name as original basename)
        └── meta.json             ← trash metadata (see schema)
```

### `meta.json` schema

```json
{
  "id": "uuid-or-nanoid",
  "name": "photo.jpg",
  "original_path": "photos/2024/photo.jpg",
  "mime_type": "image/jpeg",
  "size_bytes": 1048576,
  "sha256": "...",
  "deleted_at": "2026-07-03T14:00:00Z",
  "modified_at": "2026-06-01T10:00:00Z"
}
```

**Rules:**

| Rule | Rationale |
|---|---|
| `{file-id}` matches pre-trash metadata cache ID | Stable restore, thumbnail reuse |
| `original_path` is relative to `userdata/` | Same path vocabulary as today |
| `.trash/` is never exposed via normal `GET /files?path=` | Trash is a separate API surface |
| Existing volumes get `.trash/` on first soft-delete (lazy) | No migration job required |
| Re-import reads `.trash/` as-is | Portability |

**ADR required:** Yes — `docs/adr/005-volume-trash-layout.md` before implementation. Updates `docs/project.md` volume structure diagram.

---

## Behaviour specification

### Soft-delete (replaces current `FileService.Delete`)

1. Validate auth + path (existing rules).
2. Read metadata record (if any) and file stat.
3. Create `.trash/{file-id}/` and write `meta.json`.
4. Move blob from `userdata/...` → `.trash/{file-id}/{basename}` (`os.Rename`).
5. Remove metadata cache entry for original path (file no longer in userdata).
6. Remove Bleve index entry for original path.
7. **Do not** delete thumbnail cache entry.
8. **Do not** decrement `used_bytes`.
9. Emit `file.trashed`.

### Restore

1. Load `meta.json` for `{file-id}`.
2. Compute target path = `original_path`. If a file already exists at that path → **409 PATH_OCCUPIED** (v1: no auto-rename; user must rename conflicting file first).
3. Move blob back to `userdata/{original_path}` (create parent dirs if needed).
4. Rewrite metadata cache; re-index Bleve.
5. Remove `.trash/{file-id}/` directory.
6. Emit `file.restored`.

### Permanent delete (single item)

1. Remove `.trash/{file-id}/` tree.
2. Delete thumbnail + any orphaned metadata for that ID.
3. Decrement `used_bytes` by `size_bytes`.
4. Emit `file.deleted`.

### Empty trash

1. For each item in `.trash/`, permanent delete (batch).
2. Return count purged.

### Auto-purge

- Runs on server startup and every 24h (same pattern as other background maintenance — implementation detail in plan).
- Permanently deletes items where `deleted_at + retention_days < now()`.
- Logs count per volume at info level.

### Path conflict on restore

| Case | Behaviour |
|---|---|
| Original path free | Restore there |
| File exists at original path | `409 PATH_OCCUPIED` with `{ conflicting_path }` |
| Original parent directory missing | Recreate parent directories |

---

## API

Base path: `/api/volumes/:id/trash` — auth via existing volume ownership checks.

| Method | Path | Body / query | Response | Notes |
|---|---|---|---|---|
| `GET` | `/trash` | `?limit=50&offset=0` | `{ items: TrashEntry[], total: number }` | Sorted by `deleted_at` desc |
| `POST` | `/trash/:fileId/restore` | — | `200 FileEntry` | Restored file metadata |
| `DELETE` | `/trash/:fileId` | — | `204` | Permanent delete one item |
| `DELETE` | `/trash` | — | `204` + `{ purged: number }` | Empty all trash |

**Existing endpoint change:**

| Method | Path | Change |
|---|---|---|
| `DELETE` | `/files?path=` | Soft-delete → trash (was permanent). Response stays `204`. |

### `TrashEntry` JSON

```json
{
  "id": "file-id",
  "name": "photo.jpg",
  "original_path": "photos/2024/photo.jpg",
  "mime_type": "image/jpeg",
  "size_bytes": 1048576,
  "deleted_at": "2026-07-03T14:00:00Z",
  "modified_at": "2026-06-01T10:00:00Z",
  "thumbnail_url": "/api/volumes/{id}/files/thumbnail?path=..."
}
```

Thumbnail URL in trash: serve from cache by file ID (new query param `?id=` on thumbnail endpoint, or internal-only path — **prefer `?id=`** to avoid faking userdata paths).

### Error codes

| Code | HTTP | When |
|---|---|---|
| `NOT_FOUND` | 404 | Unknown file ID or volume |
| `PATH_OCCUPIED` | 409 | Restore blocked by existing file |
| `FORBIDDEN` | 403 | No volume access |
| `PATH_TRAVERSAL` | 422 | Invalid path |

---

## Event bus

| Event | When | Payload (key fields) |
|---|---|---|
| `file.trashed` | Soft-delete | `volume_id`, `file_id`, `original_path`, `size_bytes`, `deleted_at` |
| `file.restored` | Restore | `volume_id`, `file_id`, `restored_path`, `size_bytes` |
| `file.deleted` | Permanent purge | `volume_id`, `relative_path` (original_path), `size_bytes` |

Add to `volume.EventPublisher`, `internal/plugin/bridge.go`, `pkg/pluginsdk/events.go`.

**Breaking change note:** Plugins subscribing to `file.deleted` today receive it on user delete. After this feature, user delete emits `file.trashed` instead. Document in ADR-005 and changelog. Permanent purge still emits `file.deleted`.

---

## Tech Stack

See [docs/project.md — Tech Stack](../../project.md#tech-stack).

| Layer | Change |
|---|---|
| Backend | `internal/volume/trash.go` — trash ops (or extend `file_ops.go`) |
| Backend | `internal/volume/file_service.go` — delegate delete to trash |
| Backend | `internal/volume/layout.go` — optional `.trash/` in new volume layout |
| Backend | `internal/volume/events.go` — `FileTrashedEvent`, `FileRestoredEvent` |
| Backend | `internal/api/trash_handler.go` — trash REST handlers |
| Backend | `internal/api/router.go` — trash routes |
| Backend | `cmd/server/main.go` — trash sweeper goroutine; retention env |
| Backend | `internal/plugin/` — bridge + event constants |
| Backend | `pkg/pluginsdk/` — event constants |
| Frontend | `web/src/components/volumes/trash/` — `TrashView`, `TrashToolbar` |
| Frontend | `web/src/components/volumes/explorer/VolumeExplorer.tsx` — trash mode toggle |
| Frontend | `web/src/components/volumes/explorer/FolderTree.tsx` — Trash nav entry |
| Frontend | `web/src/hooks/useVolumeTrash.ts` — React Query mutations |
| Frontend | `web/src/lib/api.ts` — trash API client methods |
| Config | `.env.example` — `LCLOUD_TRASH_RETENTION_DAYS=30` |

**Dependencies:** No new npm or Go modules.

---

## Commands

### Development

```bash
# Backend
go test ./internal/volume/... -run Trash
go test ./internal/api/... -run Trash

# Frontend
cd web && npm run test
cd web && npm run build

# Full stack smoke
make embed
docker compose up -d --build
# → http://localhost:8080/volumes/:id — delete file, open Trash, restore
```

### Lint / typecheck

```bash
go test ./...
cd web && npm run build
```

---

## UI specification

Read [design/DESIGN.md](../../../design/DESIGN.md) before implementation.

### Navigation

- Add **Trash** entry at the bottom of the explorer sidebar (`FolderTree`), visually separated from folder tree (icon `Trash2`, muted label).
- Selecting Trash switches explorer to trash mode — breadcrumb shows `Trash`, folder tree selection cleared.
- URL: `/volumes/:id?trash=1` (query flag) or `/volumes/:id/trash` — **prefer `?view=trash`** to reuse `VolumeDetailPage` shell.

### Trash view

- Reuse list view columns where applicable: Name, Original location, Deleted, Size.
- Row actions (context menu + toolbar when one item selected): **Restore**, **Delete permanently**.
- Toolbar: **Empty trash** (destructive, confirm dialog with item count).
- Empty state: « Trash is empty » + muted helper text about retention period.
- Delete confirm in normal explorer: « Move to trash? » (not « Delete permanently »).

### Destructive styling

Follow [control-panel-iconography SPEC](../control-panel-iconography/SPEC.md): `Trash2` + `text-accent-rose` for permanent delete and empty trash.

---

## Testing strategy

### Backend (Go + testify)

| Test | Covers |
|---|---|
| Soft-delete moves blob + writes meta; userdata file gone | TR1 |
| Soft-delete keeps `used_bytes` unchanged | TR7 |
| Soft-delete removes index entry; file absent from List | TR6 |
| Restore returns file to original path + re-indexes | TR3 |
| Restore with occupied path → `PATH_OCCUPIED` | Conflict |
| Permanent delete frees quota + emits deleted | TR4 |
| Empty trash purges all items | TR5 |
| Sweeper deletes items past retention | TR8 |
| Lazy `.trash/` creation on old volumes | Migration |
| Auth: other user's volume → 403 | Security |

### API (handler tests)

| Test | Covers |
|---|---|
| `GET /trash` pagination | List |
| `POST /trash/:id/restore` round-trip | Restore |
| `DELETE /files` now soft-deletes | Regression |
| `DELETE /trash/:id` permanent | Purge one |

### Frontend (Vitest)

| Test | Covers |
|---|---|
| Trash sidebar entry renders | Navigation |
| Trash mode shows items from API | TR2 |
| Restore mutation invalidates queries | TR3 |
| Empty trash confirm dialog | TR5 |

Manual smoke:

- [ ] Delete file → appears in Trash → restore → file back in original folder
- [ ] Permanent delete removes item; quota decreases
- [ ] Empty trash clears all
- [ ] Thumbnail visible in trash list for images
- [ ] Plugin receives `file.trashed` on delete (manual or test plugin)

---

## Code style

Follow existing volume module patterns:

- Trash logic in `internal/volume/` — handlers stay thin
- Shared disk ops extracted if overlap with `file_ops.go` (move/rename primitives)
- Frontend hooks mirror `useVolumeFiles.ts` naming: `useVolumeTrash`, `useRestoreTrashItem`, etc.

Handler shape:

```go
func (h *TrashHandler) Restore(c *gin.Context) {
    claims, _ := auth.ClaimsFromContext(c)
    volumeID, fileID := parseParams(c)
    entry, err := h.trash.Restore(claims, volumeID, fileID)
    if err != nil { mapTrashError(c, err); return }
    httputil.JSON(c, http.StatusOK, entry)
}
```

---

## Boundaries

**Always:**

- Keep trash blobs on disk under `.trash/` — never PostgreSQL-only tombstones
- Update ADR + `docs/project.md` volume diagram before merging Slice 1
- Emit distinct events for trashed / restored / permanently deleted
- Validate ownership on every trash endpoint
- Read `design/DESIGN.md` before UI work

**Ask first:**

- Changing `.volume.json` schema for per-volume retention
- Making task macros trash-aware (behaviour change for automation users)
- WebDAV/FTP delete semantics (hard vs soft)
- Adding bulk trash endpoints

**Never:**

- Expose `.trash/` via normal file list API
- Bypass `VolumeIndexer` on restore (must re-index through interface)
- Decrement quota on soft-delete
- Call plugin functions directly from trash service — events only
- Permanent delete without explicit user action or retention sweeper

---

## Success criteria

- [ ] User delete moves file to trash; file recoverable from Trash UI
- [ ] Restore, permanent delete, and empty trash work end-to-end
- [ ] Trashed files excluded from normal listing and search
- [ ] Quota unchanged on soft-delete; decreases on permanent purge
- [ ] Auto-purge removes items older than `LCLOUD_TRASH_RETENTION_DAYS`
- [ ] `file.trashed`, `file.restored`, `file.deleted` wired through plugin bridge
- [ ] ADR-005 accepted; volume structure docs updated
- [ ] `go test ./internal/volume/... ./internal/api/...` pass
- [ ] `cd web && npm run test && npm run build` pass
- [ ] No new npm dependencies

---

## Open questions

| # | Question | Default if unanswered |
|---|---|---|
| OQ1 | Restore conflict: auto-rename (`photo (1).jpg`) vs 409 error? | **409 PATH_OCCUPIED** — explicit user action |
| OQ2 | Retention default 30 days or 60? | **30 days** (Drive default) |
| OQ3 | Show retention warning in Trash UI (« Items deleted after N days »)? | **Yes** — footer helper text |
| OQ4 | Thumbnail in trash: reuse `?path=` with virtual path or new `?id=` param? | **`?id=`** query on thumbnail endpoint |
| OQ5 | Sweeper interval: 24h fixed or configurable? | **24h fixed** in v1 |
| OQ6 | `delete_old_files` macro: document as permanent bypass or change to trash? | **Resolved — soft-delete to trash**; use `purge_trash` for permanent purge |

---

## Relationship to other ideas

| Idea | Relationship |
|---|---|
| Multi-file selection & bulk actions | Future bulk trash/restore builds on this API |
| File versions / history | Complementary — trash = accidental delete; versions = overwrite |
| Volume explorer UX | Replaces « permanent delete with confirm » assumption in explorer spec |
| WebDAV / FTP | Trash semantics undefined until protocol spec updated |
| Dashboard widgets | Future « trash count » widget optional |

---

## ADR needed?

**Yes — ADR-005: Volume trash directory layout and event semantics.** Covers:

- `.trash/` directory contract and `meta.json` schema
- Portability on re-import
- Event bus breaking change (`file.deleted` → `file.trashed` on user delete)
- Quota accounting rules

Update [docs/project.md](../../project.md) volume structure section when ADR is accepted.
