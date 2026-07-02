# ADR-003: Volume explorer file operations and preview extension

## Status

Accepted

## Date

2026-07-02

## Context

The MVP volume detail page exposes list/upload/delete/download only. The volume explorer redesign (SPEC.md) requires:

- Web API for move and rename (files and directory rename with metadata cascade)
- Event bus semantics distinguishing move vs rename
- In-browser preview for PDF, text, and images via `Content-Disposition: inline`
- A frontend `FileOpener` registry extensible by plugins post-v1 (e.g. FFmpeg video preview)

`MoveFileInternal` already exists in `macro_ops.go` for task macros but is not exposed to the web UI.

## Decision

### Shared file operations: `internal/volume/file_ops.go`

Extract `MoveFile` and `RenameEntry` into `file_ops.go`. Both `MacroOps` and `FileService` delegate to this module — no duplicated disk/metadata/index logic.

### Event semantics

| Operation | Event | Condition |
|---|---|---|
| Move file | `file.moved` | Destination directory changes (`from_path` → `to_path`) |
| Rename file or directory | `file.renamed` | Parent directory unchanged; last segment changes |

Add `FileRenamedEvent` to `volume.EventPublisher` and `FromFileRenamed` in the plugin bridge.

### Web API

| Method | Path | Body |
|---|---|---|
| `PATCH` | `/api/volumes/:id/files/move` | `{ "from_path", "to_path" }` |
| `PATCH` | `/api/volumes/:id/files/rename` | `{ "path", "new_name" }` |

Auth via existing `FileService` + `volumes.Get(claims, ...)`. Move applies to **files only** in v1.

### Directory rename

`RenameEntry` on a directory:

1. `os.Rename` on the directory path
2. Walk metadata cache; rewrite `RelativePath` for all records under the old prefix
3. Re-index each affected file in Bleve
4. Emit single `file.renamed` with `entry_type: directory`

Directory **move** (drag folder) is out of v1 scope.

### Content disposition

Extend `GET /files/content` with optional query `disposition=inline|attachment` (default `attachment`). Preview uses `inline`; download keeps `attachment`.

### FileOpener registry (frontend)

Core ships a priority-sorted registry (`registerFileOpener`, `resolveFileOpener`) with built-in handlers for images, text (512 KiB cap), and PDF. Plugins will register additional openers via a future bus/SDK contract — not in this ADR's implementation scope.

## Consequences

- Positive: One code path for macro and web file mutations; explorer UX matches cloud-drive expectations
- Positive: Plugin video preview can extend registry without explorer layout changes
- Negative: Directory rename may be slow on volumes with many indexed files under the tree
- Migration: No schema or volume structure changes

## Alternatives considered

| Alternative | Rejected because |
|---|---|
| Separate rename API only (no move) | Drag-move and macro move both need move primitive |
| Emit `file.moved` for all renames | Plugins distinguishing rename vs move would break |
| Server-side layout prefs in v1 | Operator deferred; localStorage sufficient for homelab |
