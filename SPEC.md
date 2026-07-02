# Spec: Volume file explorer — UX redesign

> **Status:** Draft — pending review  
> **Scope:** Replace the MVP file browser on `/volumes/:id` with a familiar cloud-drive layout (sidebar tree, toolbar, list/grid views, inline rename, drag-and-drop upload + internal move, in-explorer file preview for PDF/text/images via an extensible opener registry). Expose move/rename file operations via the web API.  
> **Prerequisite:** Phase 1.1 (volumes + file API), Phase 1.4 (`MoveFileInternal`, `file.moved` event bus wiring).  
> **Sources:** [VISION.md](./VISION.md), [docs/project.md](./docs/project.md), [IDEAS.md — Volume file explorer UX redesign](./IDEAS.md), [design/DESIGN.md](./design/DESIGN.md)

---

## Assumptions (correct me now or I proceed)

1. **MVP file API is shipped** — list, upload (single file), create directory, download, delete, thumbnail.
2. **`MoveFileInternal` exists** — in `internal/volume/macro_ops.go`, files only, no JWT; used by task macros today, not exposed to the web UI.
3. **`file.renamed` is defined on the plugin bus** — constant exists in `internal/plugin/events.go`, but no `FileRenamed` publisher in `volume.EventPublisher` yet; rename will add it.
4. **No new volume directory structure** — explorer redesign is UI + web API only; `.volume.json` unchanged.
5. **Layout preferences stay client-side in v1** — view mode, columns, widths stored in `localStorage`; server-side prefs deferred until post-ship validation (IDEAS.md).
6. **Target UX = OneDrive / Google Drive / Dropbox** — not a custom admin-table layout; operator decisions in IDEAS.md are binding for v1.
7. **Directory drag-move is out of v1** — drag-and-drop move applies to **files** onto folder targets; folder relocation deferred (metadata path cascade is heavier).
8. **Inline rename applies to files and folders** — Windows Explorer mental model (F2, context menu, slow double-click on name).
9. **Multi-file upload is in v1** — drop external files onto the contents pane uploads all dropped files with per-file progress (IDEAS competitive gap; natural fit once the permanent upload banner is removed).
10. **Multi-select / bulk delete-download is v1.1** — not blocking the layout redesign; can ship immediately after if time allows.
11. **Trash / soft-delete is out of scope** — separate IDEAS entry; explorer keeps permanent delete with confirm dialog for v1.
12. **Copy / duplicate is out of scope** — separate IDEAS entry; no copy API in v1.
13. **No new backend search/tree endpoint in v1** — lazy folder tree uses existing `GET /files?path=` per expanded node.
14. **`@dnd-kit/core` approved for drag-and-drop** — `@dnd-kit/core` + `@dnd-kit/utilities` (OQ1 ✅).
15. **Double-click opens preview when a registered opener matches** — v1 built-in openers: **PDF**, **text**, **images**; other types fall back to download.
16. **File opener registry is the extension point** — v1 ships core + built-in openers only; plugins register additional MIME handlers later (e.g. video → FFmpeg transcode stream) without redesigning the explorer.

---

## Objective

Replace the minimal MVP explorer (`VolumeDetailPage` + flat `FileBrowser` + permanent `UploadZone` + always-visible "New folder" form) with a **standard cloud file manager** that homelab users already know how to use.

**Who:** Self-hosters managing files daily in the web UI — same audience as VISION.md Principle 3 (technical, expects familiar tooling).

**Pain points addressed (operator feedback, IDEAS.md):**

| Today (MVP) | Target (v1) |
|---|---|
| Permanent dashed upload banner above the list | Upload via toolbar button + drop **onto** the contents pane |
| Always-visible "New folder" text field | **New ▾** menu → "New folder" (inline input or small dialog) |
| Flat table, click folder name to navigate | **Left folder tree** (lazy) + breadcrumb + main pane |
| List only, fixed columns | **List + grid** toggle; customizable list columns |
| No rename | **Inline rename** (files + folders) |
| No internal drag-move | **Drag file → folder** (tree or list row) to move |
| Single-file upload (`files[0]`) | **Multi-file** drop / file picker |

**User stories:**

| ID | Story |
|---|---|
| UX1 | As a user, I open a volume and see a folder tree on the left and files on the right — like Google Drive. |
| UX2 | As a user, I drop files from my desktop onto the file list and they upload into the current folder, with progress per file. |
| UX3 | As a user, I click **New → Folder**, type a name, and the folder appears without a permanent form taking space. |
| UX4 | As a user, I press F2 (or use the context menu) to rename a file or folder in place. |
| UX5 | As a user, I drag a file onto a folder in the tree or list to move it. |
| UX6 | As a user, I switch between list and grid view; my choice persists on this browser. |
| UX7 | As a user, I customize which columns appear in list view (name always visible); widths and order persist locally. |
| UX8 | As a user, I double-click a PDF, text file, or image and it opens in a preview panel inside the explorer — without leaving the page. |
| UX9 | As a user, I double-click an unsupported file type and it downloads (same as today). |

**Out of scope for v1:**

- FTP / WebDAV UI, sync/conflict UI, in-browser **text editor** (read-only text preview only)
- **Video / audio preview** — deferred; opener registry + plugin contract defined so a plugin can add them (e.g. FFmpeg transcode for editors)
- Trash / soft-delete, file copy/duplicate, version history
- Directory drag-move (move entire folders by drag)
- Bulk multi-select actions (planned v1.1 — see [Follow-up slice](#follow-up-slice-v11))
- Server-side layout preference storage
- Changes to `.volume.json` or volume directory structure
- Full-text content search in the explorer (Bleve filename search stays in sidebar `GlobalSearch`)

---

## Tech Stack

See [docs/project.md — Tech Stack](./docs/project.md#tech-stack).

This feature adds or extends:

| Layer | Addition |
|---|---|
| Backend | `FileService.Move`, `FileService.Rename` (auth-aware wrappers over shared move/rename logic) |
| Backend | `PATCH /api/volumes/:id/files/move`, `PATCH /api/volumes/:id/files/rename` |
| Backend | `FileRenamedEvent` on `EventPublisher` + plugin bridge `FromFileRenamed` |
| Backend | `RenameDirectoryInternal` — filesystem rename + metadata/index path cascade for folders |
| Frontend | Explorer shell: tree, toolbar, contents pane, context menu, inline rename |
| Frontend | `localStorage` layout prefs store |
| Frontend | Multi-file upload queue with per-file progress |
| Frontend | `@dnd-kit/core` + `@dnd-kit/utilities` for drag-and-drop |
| Frontend | **File opener registry** + built-in openers (PDF, text, image) + `PreviewPanel` |
| Frontend | shadcn **Context Menu** + **Dropdown Menu** components (add to `web/src/components/ui/`) |
| Backend | `GET /files/content?disposition=inline` for in-browser preview (attachment remains default) |

---

## Commands

### Development

```bash
# Backend — file move/rename focus
go test ./internal/volume/... -run 'Move|Rename' -cover
go test ./internal/api/... -run File -cover
go test ./...

# Frontend
cd web && npm run test
cd web && npm run build

# Local stack
docker compose up -d --build
docker compose logs -f app
```

### Verification (explorer redesign done)

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"<ADMIN_EMAIL>","password":"<ADMIN_PASSWORD>"}' \
  | jq -r '.access_token')

VOL_ID="<volume-uuid>"

# 1. List root (unchanged)
curl -s "http://localhost:8080/api/volumes/$VOL_ID/files?path=." \
  -H "Authorization: Bearer $TOKEN" | jq

# 2. Create folder via API (unchanged)
curl -s -X POST "http://localhost:8080/api/volumes/$VOL_ID/files/directories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"path":"photos/2024"}' | jq

# 3. Move file into folder
curl -s -X PATCH "http://localhost:8080/api/volumes/$VOL_ID/files/move" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"from_path":"vacation.jpg","to_path":"photos/2024/vacation.jpg"}' | jq

# 4. Rename file in place
curl -s -X PATCH "http://localhost:8080/api/volumes/$VOL_ID/files/rename" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"path":"photos/2024/vacation.jpg","new_name":"summer.jpg"}' | jq

# 5. Rename folder
curl -s -X PATCH "http://localhost:8080/api/volumes/$VOL_ID/files/rename" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"path":"photos/2024","new_name":"2024-summer"}' | jq
```

---

## Project Structure

See [docs/project.md — Project structure](./docs/project.md#project-structure).

This feature creates or extends:

```
lcloud/
├── internal/
│   ├── volume/
│   │   ├── file_ops.go              ← shared MoveFile/RenameFile/RenameDirectory (used by FileService + MacroOps)
│   │   ├── file_ops_test.go
│   │   ├── file_service.go          ← Move, Rename methods
│   │   ├── events.go                ← FileRenamedEvent + publisher method
│   │   └── macro_ops.go             ← delegate to shared file_ops (no duplication)
│   ├── plugin/
│   │   └── events.go                ← FromFileRenamed bridge
│   └── api/
│       ├── file_handler.go          ← Move, Rename handlers
│       └── router.go                ← register PATCH routes
├── web/src/
│   ├── pages/
│   │   └── VolumeDetailPage.tsx     ← thin shell; compose explorer layout
│   ├── components/volumes/explorer/
│   │   ├── VolumeExplorer.tsx       ← layout orchestrator
│   │   ├── FolderTree.tsx           ← lazy sidebar tree
│   │   ├── ExplorerToolbar.tsx      ← New, Upload, view toggle, column picker
│   │   ├── ContentsPane.tsx         ← drop target wrapper
│   │   ├── FileListView.tsx         ← resizable/reorderable columns
│   │   ├── FileGridView.tsx         ← thumbnail grid
│   │   ├── FileContextMenu.tsx
│   │   ├── InlineRename.tsx
│   │   ├── UploadQueue.tsx          ← multi-file progress
│   │   ├── PreviewPanel.tsx         ← in-explorer preview shell
│   │   └── explorer.test.tsx
│   ├── lib/fileOpeners/
│   │   ├── registry.ts              ← FileOpener interface + register/resolve
│   │   ├── imageOpener.ts
│   │   ├── textOpener.ts
│   │   ├── pdfOpener.ts
│   │   └── registry.test.ts
│   ├── hooks/
│   │   ├── useVolumeFiles.ts        ← add useMoveFile, useRenameFile
│   │   ├── useExplorerPrefs.ts      ← localStorage read/write
│   │   └── useFilePreview.ts        ← resolve opener + panel state
│   └── store/
│       └── explorerPrefs.ts         ← optional Zustand mirror of prefs
├── docs/
│   └── adr/
│       └── NNN-volume-explorer-file-ops.md   ← web move/rename API (created at /build)
└── SPEC.md                          ← this file
```

**Removed / deprecated after ship:**

- `web/src/components/volumes/UploadZone.tsx` — logic absorbed into `ContentsPane` + `UploadQueue` (file may remain exported for tests until deleted)
- `web/src/components/volumes/FileBrowser.tsx` — replaced by `FileListView` / `FileGridView`

---

## Architecture

### Layout (target)

```
┌─────────────────────────────────────────────────────────────────┐
│ Header — volume name, quota, "Back to volumes"                  │
├──────────────┬──────────────────────────────────────────────────┤
│ FolderTree   │ ExplorerToolbar — [New▾] [Upload] [≡|▦] [Columns]│
│ (lazy,       ├──────────────────────────────────────────────────┤
│  resizable)  │ BreadcrumbNav                                    │
│              ├──────────────────────────────────────────────────┤
│              │ ContentsPane — list OR grid                      │
│              │  · external drop → upload queue                  │
│              │  · internal drag → move onto folder highlight  │
│              │  · context menu, inline rename                   │
└──────────────┴──────────────────────────────────────────────────┘
```

### Module boundaries (AGENTS.md)

- **`internal/volume/`** owns all disk mutations — handlers call `FileService` only.
- **`MoveFileInternal` logic consolidated** — extract shared implementation in `file_ops.go`; `MacroOps` and `FileService` both call it (no duplicated `os.Rename` paths).
- **`VolumeIndexer` interface only** — index updates inside `file_ops.go` via `indexManager`; never import Bleve from `api/`.
- **Event bus** — `file.moved` when destination directory changes; `file.renamed` when parent directory unchanged (name change only). Plugin bridge maps both.
- **No business logic in Gin handlers** — validate JSON, call service, map errors.

### Lazy folder tree

- Tree root = volume root (`path=.`), fetch via `GET /files?path=<node>`.
- Client filters `entries` where `type === "directory"`.
- Expand node → fetch children; cache in TanStack Query (`staleTime: 30s`).
- Selecting a tree node sets `currentPath` → contents pane refetches same path.
- Tree stays in sync on create/move/rename/delete via query invalidation on `["volumes", volumeId, "files"]`.

### Drag-and-drop semantics

| Drag source | Drop target | Action |
|---|---|---|
| External files (OS) | Contents pane (not on a folder row) | Upload into `currentPath` |
| External files | Folder row / tree node | Upload into that folder's path |
| Internal file row | Folder row / tree node | `PATCH .../move` |
| Internal file row | Contents pane background | No-op (or move to current folder = no-op) |

**Visual feedback:** valid folder target highlights with `border-primary` + `bg-surface-elevated` per DESIGN.md.

**Upload vs move detection:** `dataTransfer.types` — if `"Files"` from external, upload; if internal drag uses `application/x-lcloud-path` custom type, move.

### File opener registry (preview on double-click)

Double-click (or context menu **Open**) resolves the first matching opener from a **priority-sorted registry**. Unmatched types → download fallback.

#### `FileOpener` interface (frontend contract)

```typescript
export type FileOpenerContext = {
  volumeId: string;
  entry: FileEntry;
  contentUrl: string; // authenticated content URL
};

export interface FileOpener {
  id: string;
  label: string;
  priority: number;
  canOpen(entry: FileEntry): boolean;
  open(ctx: FileOpenerContext, panel: PreviewPanelApi): void;
}
```

- **`registerFileOpener(opener)`** — append to registry (built-ins register at module init).
- **`resolveFileOpener(entry)`** — highest `priority` wins among matchers.
- **`PreviewPanel`** — shared shell: header (filename, close, download), body slot rendered by opener.

#### Built-in openers (v1)

| Opener | Match rule | Render |
|---|---|---|
| `core.image` | `mime_type` starts with `image/` | `<img>` via `contentUrl` (`disposition=inline`) |
| `core.text` | `text/*` or ext `.txt`, `.md`, `.json`, `.yaml`, `.yml`, `.log`, `.csv` | Fetch text (max **512 KiB**); show in scrollable `<pre>` (JetBrains Mono); truncate with notice if larger |
| `core.pdf` | `application/pdf` or ext `.pdf` | `<iframe>` or `<embed>` via inline content URL |

**Not in v1:** `video/*`, `audio/*` — no built-in opener; double-click downloads. Registry slot reserved for plugins.

#### Plugin extension (designed now, implemented post-v1)

Aligns with VISION extensibility — plugins must not patch explorer React code.

**Phase B (follow-up, separate `/spec` or plugin SDK ADR):**

1. Plugins declare preview capabilities at registration, e.g. `{ "mime_prefixes": ["video/"], "opener_id": "ffmpeg.transcode" }`.
2. On open, core emits **`file.open.requested`** on the event bus with `{ volume_id, path, mime_type }`.
3. Plugin responds (RPC or cached artifact) with a **preview URL** or **transcoded stream path** — e.g. FFmpeg plugin writes `{volume}/cache/previews/{file-id}.mp4` and returns URL.
4. Core registers a dynamic opener that delegates to the plugin response.

**Example (future):** video editor plugin — user double-clicks `clip.mov` → plugin transcodes to H.264 preview → explorer plays inline. No codec support required in core.

**v1 requirement:** registry API + `PreviewPanel` must accept externally registered openers without layout changes — only `registerFileOpener()` calls.

#### Preview backend

Extend existing content endpoint (no new route required):

```
GET /api/volumes/:id/files/content?path=...&disposition=inline|attachment
```

| Param | Default | Behavior |
|---|---|---|
| `disposition` | `attachment` | `inline` → `Content-Disposition: inline` for browser preview; `attachment` → download (unchanged) |

Text opener fetches via authenticated `fetch(contentUrl)` — same endpoint, inline disposition.

### Layout preferences (`localStorage`)

Key: `lcloud.explorer.prefs.v1`

```json
{
  "viewMode": "list",
  "listColumns": [
    { "id": "name", "visible": true, "width": 320, "order": 0 },
    { "id": "size", "visible": true, "width": 100, "order": 1 },
    { "id": "modified", "visible": true, "width": 180, "order": 2 },
    { "id": "type", "visible": false, "width": 120, "order": 3 }
  ],
  "treeWidth": 240
}
```

- **`name` column** — always visible, not hideable.
- **Launch columns:** name, size, modified (type optional via picker).
- Column resize: drag handle on header; reorder: drag header (list view).
- Invalid/missing prefs → defaults above.

### Inline rename

| Trigger | Behavior |
|---|---|
| F2 | Start rename on focused row |
| Context menu → Rename | Same |
| Slow double-click on name | Same (300ms threshold; single click still opens folder / selects) |
| Enter | Commit |
| Escape | Cancel |

Validation: reject empty name, `/`, `\`, `..`; trim whitespace; preserve extension on file rename optional (user can change full name). Collision → API `PATH_EXISTS` → inline error toast.

### Multi-file upload (queue + max 3 parallel)

- Toolbar **Upload** opens multi-select file picker.
- Drop accepts `FileList` — **all files enter a FIFO queue** immediately (nothing is dropped or ignored).
- **At most 3 uploads run in parallel.** When one finishes (success or error), the next **queued** file starts automatically — no user action required.
- Example: 10 files dropped → 3 upload immediately, 7 wait; as each active slot frees, the next waiting file starts.

**Per-file states in `UploadQueue` UI:**

| State | Meaning |
|---|---|
| `queued` | Waiting for a free slot (position visible if helpful) |
| `uploading` | Active `POST /files` in progress (progress bar) |
| `done` | Success — dismissible row |
| `error` | Failed — message + optional retry (re-queues at tail) |

- Queue bar below toolbar: filename, state badge, progress % when uploading, error text when failed.
- Reuse existing `POST /files` per file; no new batch endpoint in v1.
- **Cancel** (optional v1): cancel a `queued` item removes it from the queue; cancel an `uploading` item aborts the in-flight request if the browser allows it.

---

## Backend

### Shared file operations (`internal/volume/file_ops.go`)

Extract from `macro_ops.go`:

```go
// MoveFile moves a file within userdata; updates metadata, index, stats; emits file.moved.
func MoveFile(ctx context.Context, deps FileOpDeps, vol *Volume, fromRel, toRel string) error

// RenameEntry renames a file or directory (last path segment only).
// Files → file.renamed. Directories → cascade metadata/index paths + file.renamed on dir path.
func RenameEntry(ctx context.Context, deps FileOpDeps, vol *Volume, relPath, newName string) error
```

`MacroOps.MoveFileInternal` becomes a thin wrapper calling `MoveFile`.

**Directory rename cascade:** when renaming `photos/2024` → `photos/2024-summer`:

1. `os.Rename` on directory.
2. Walk metadata cache records with path prefix `photos/2024/` → rewrite `RelativePath` and `Name`.
3. For each affected file: delete old index path, index new path.
4. Thumbnail paths unchanged (keyed by file ID, not path).
5. Emit single `file.renamed` for the directory path change (payload: old_path, new_path, entry_type).

**Directory move (drag folder):** deferred v1 — would share cascade logic with a different API.

### API

Base path: `/api`. All endpoints require JWT.

| Method | Path | Description |
|---|---|---|
| `GET` | `/volumes/:id/files?path=` | Unchanged — list directory |
| `GET` | `/volumes/:id/files/content?path=&disposition=` | Extended — `inline` for preview, `attachment` (default) for download |
| `POST` | `/volumes/:id/files` | Unchanged — upload one file |
| `POST` | `/volumes/:id/files/directories` | Unchanged — create directory |
| `PATCH` | `/volumes/:id/files/move` | Move file to new relative path |
| `PATCH` | `/volumes/:id/files/rename` | Rename file or directory (same parent) |
| `DELETE` | `/volumes/:id/files?path=` | Unchanged — delete file |

#### `PATCH /volumes/:id/files/move`

```json
{
  "from_path": "report.pdf",
  "to_path": "documents/report.pdf"
}
```

Response `200`:

```json
{
  "path": "documents/report.pdf",
  "type": "file",
  "name": "report.pdf"
}
```

Rules:

- `from_path` must be an existing **file** (not directory in v1).
- `to_path` must not exist.
- `to_path` parent directory created if missing (same as upload).
- Destination must stay under `./userdata/` (PathResolver).
- Emits `file.moved`.

#### `PATCH /volumes/:id/files/rename`

```json
{
  "path": "documents/report.pdf",
  "new_name": "annual-report.pdf"
}
```

Response `200`:

```json
{
  "path": "documents/annual-report.pdf",
  "type": "file",
  "name": "annual-report.pdf"
}
```

Rules:

- `new_name` = final segment only (no slashes).
- Works for files and directories.
- Emits `file.renamed` (not `file.moved` — parent unchanged).

**Error codes (new or reused):**

| Code | Condition |
|---|---|
| `FILE_NOT_FOUND` | Source path missing |
| `PATH_EXISTS` | Destination or new name already exists |
| `INVALID_PATH` | Path traversal, empty name, invalid characters |
| `NOT_A_FILE` | Move requested on a directory |
| `FORBIDDEN` | User does not own volume (non-admin) |
| `QUOTA_EXCEEDED` | N/A for move/rename (no size change) |

### Event catalog (this feature)

| Event | When | Payload highlights |
|---|---|---|
| `file.moved` | File moved to different directory | `from_path`, `to_path`, `size_bytes` |
| `file.renamed` | File or folder renamed in place | `old_path`, `new_path`, `entry_type` (`file` \| `directory`) |

Add to `volume.EventPublisher`:

```go
type FileRenamedEvent struct {
    VolumeID  uuid.UUID
    OldPath   string
    NewPath   string
    EntryType string // "file" | "directory"
}
```

---

## Frontend

Follow [design/DESIGN.md](./design/DESIGN.md): dark canvas, electric yellow primary actions, Inter + JetBrains Mono, hairline borders.

### Route

| Path | Content |
|---|---|
| `/volumes/:id` | Redesigned explorer (replaces current layout) |
| `/volumes/:id?path=foo/bar` | Deep-link to folder (unchanged query param) |

### Toolbar actions

| Control | Behavior |
|---|---|
| **New ▾** | Dropdown: "New folder" → inline name input in pane or small dialog |
| **Upload** | Hidden `<input type="file" multiple>` |
| **View toggle** | List (rows icon) / Grid (grid icon); active state `text-primary` |
| **Columns** | List view only — checkbox menu for optional columns |

### List view

- Sort by name (default), size, modified — client-side sort on current page listing (API sort deferred).
- Row hover: muted background `bg-surface-elevated`.
- Folder row: double-click opens; chevron or click name navigates.
- File row: double-click → `resolveFileOpener(entry)` → preview panel if matched, else download (UX8/UX9).
- **Preview panel** — slide-over or right split inside contents area; ESC closes; does not unmount tree/toolbar.

### Grid view

- Cards with thumbnail (`ThumbnailPreview` when `has_thumbnail`) or MIME-based icon (lucide-react until iconography pass lands).
- Folder card: folder icon, name below.
- Same context menu and inline rename as list.

### Context menu (files and folders)

| Action | Available |
|---|---|
| Open | Files: preview if opener exists, else download; Folders: open |
| Download | Files only (always available) |
| Rename | Both |
| Delete | Files only (MVP parity — folder delete out of scope) |

### Data fetching

- Existing hooks extended:
  - `useMoveFile(volumeId)` → `PATCH .../move`, invalidate files + tree queries
  - `useRenameFile(volumeId)` → `PATCH .../rename`, invalidate files + tree queries
- `useExplorerPrefs()` — read/write localStorage, SSR-safe (default prefs when `window` undefined)

---

## Code Style

### Go — FileService wrapper

```go
func (s *FileService) Move(claims *auth.Claims, volumeID uuid.UUID, fromPath, toPath string) (*FileEntry, error) {
    vol, err := s.volumes.Get(claims, volumeID)
    if err != nil {
        return nil, err
    }
    if err := MoveFile(context.Background(), s.fileOpDeps(), vol, fromPath, toPath); err != nil {
        return nil, err
    }
    return s.entryForPath(vol, toPath)
}
```

- Auth check via `volumes.Get(claims, ...)` before any disk op.
- Path cleaning via existing `cleanRelativePath` + `PathResolver`.

### TypeScript — explorer prefs

```typescript
const STORAGE_KEY = "lcloud.explorer.prefs.v1";

export function loadExplorerPrefs(): ExplorerPrefs {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return DEFAULT_PREFS;
    return { ...DEFAULT_PREFS, ...JSON.parse(raw) };
  } catch {
    return DEFAULT_PREFS;
  }
}
```

- Prefer hooks over prop drilling for `currentPath` / `viewMode`.
- DnD kit: separate sensors for pointer + keyboard where feasible.

### TypeScript — file opener registration

```typescript
registerFileOpener({
  id: "core.image",
  label: "Image preview",
  priority: 100,
  canOpen: (entry) => entry.mime_type?.startsWith("image/") ?? false,
  open: ({ contentUrl }, panel) => {
    panel.render(<img src={contentUrl} alt="" className="max-h-full max-w-full object-contain" />);
  },
});
```

- Built-in openers live in `web/src/lib/fileOpeners/`; each calls `registerFileOpener` on import.
- `VolumeExplorer` imports `./fileOpeners` side-effect bundle once at startup.

---

## Testing Strategy

See [docs/project.md — Coverage targets](./docs/project.md#coverage-targets).

| Layer | Focus | Location |
|---|---|---|
| Backend unit | Move file happy path, collision, not found | `internal/volume/file_ops_test.go` |
| Backend unit | Rename file + rename directory cascade | `internal/volume/file_ops_test.go` |
| Backend integration | Move/rename API auth + error codes | `internal/api/file_handler_test.go` |
| Backend integration | Events emitted (mock publisher) | `internal/volume/file_ops_test.go` |
| Frontend unit | Explorer prefs load/save defaults | `useExplorerPrefs.test.ts` |
| Frontend unit | InlineRename commit/cancel/validation | `InlineRename.test.tsx` |
| Frontend unit | Toolbar view toggle persists | `ExplorerToolbar.test.tsx` |
| Frontend unit | Opener registry priority + fallback | `registry.test.ts` |
| Frontend unit | Text opener size cap / PDF/image match | `*Opener.test.ts` |
| Backend integration | `disposition=inline` Content-Disposition | `file_handler_test.go` |
| Manual | UX1–UX9 checklist in browser | See Success Criteria |

**Coverage target:** 80% minimum on new `file_ops.go` paths.

**Verify before ship:**

```bash
go test ./internal/volume/... ./internal/api/... -cover
cd web && npm run test && npm run build
# Manual: drag upload, drag move, F2 rename, tree navigation, prefs persist after reload
```

---

## Boundaries

### Always

- All disk mutations through `internal/volume/` — handlers stay thin.
- Index updates via `VolumeIndexer` interface only.
- Emit `file.moved` / `file.renamed` on successful web move/rename.
- Follow [design/DESIGN.md](./design/DESIGN.md) for all new UI.
- Layout prefs in `localStorage` for v1 — no PostgreSQL prefs table.
- Preserve `?path=` deep-link behavior.
- Write ADR for web move/rename API before merge.

### Ask first

- Adding npm dependencies beyond `@dnd-kit/*`, shadcn menu primitives, and a PDF renderer if iframe preview is insufficient on target browsers.
- Folder delete API (currently files only).
- Directory drag-move (cascade move).
- Server-side preference storage.
- Batch upload API (multi-part single request).

### Never

- Change `.volume.json` schema or volume directory structure without ADR.
- Call Bleve directly outside `internal/indexer/`.
- Bypass auth in web move/rename (unlike macro internal ops — web always checks JWT).
- Add FTP/WebDAV/sync UI in this feature.
- Remove permanent delete without Trash feature shipped.

---

## Success Criteria

Explorer redesign is **done** when all of the following pass:

- [ ] **SC-EX1** `/volumes/:id` shows sidebar tree + toolbar + contents pane (no permanent upload banner, no always-visible folder form)
- [ ] **SC-EX2** Lazy tree loads children on expand; selecting node updates contents pane
- [ ] **SC-EX3** Breadcrumb + tree selection stay in sync with `currentPath`
- [ ] **SC-EX4** Drop external files onto contents pane enqueues all files; max 3 parallel, rest wait in FIFO queue with visible states (UX2)
- [ ] **SC-EX5** **New → Folder** creates directory without permanent form (UX3)
- [ ] **SC-EX6** List/grid toggle works; preference survives page reload (UX6)
- [ ] **SC-EX7** List columns: name (fixed), size, modified launch visible; user can show/hide type; resize + reorder persist (UX7)
- [ ] **SC-EX8** Inline rename via F2, context menu, slow double-click — files and folders (UX4)
- [ ] **SC-EX9** Drag file onto folder (tree or list) moves file via API (UX5)
- [ ] **SC-EX10** `PATCH .../move` and `PATCH .../rename` API with auth, validation, tests
- [ ] **SC-EX11** `file.moved` emitted on move; `file.renamed` emitted on rename (plugin bridge wired)
- [ ] **SC-EX12** Directory rename cascades metadata + Bleve index paths
- [ ] **SC-EX13** ADR documents web file ops API
- [ ] **SC-EX14** `go test ./...` and `cd web && npm run test && npm run build` pass
- [ ] **SC-EX15** Double-click PDF, text, and image opens in `PreviewPanel` without page navigation (UX8)
- [ ] **SC-EX16** Double-click unsupported type downloads; context menu **Open** follows same resolver (UX9)
- [ ] **SC-EX17** `GET /files/content?disposition=inline` serves inline Content-Disposition for preview
- [ ] **SC-EX18** `FileOpener` registry accepts a third-party registration without explorer layout changes (manual stub test)

---

## Follow-up slice (v1.1)

Optional fast-follow — not blocking v1 ship:

- Multi-select (Shift/Cmd click) + bulk delete + bulk download
- Keyboard shortcuts cheat sheet (`?`)
- Coordinate with [IDEAS.md — Iconography pass](./IDEAS.md) for MIME/disk icons in grid view
- **Plugin-backed openers** — `file.open.requested` bus flow + dynamic registration (video/FFmpeg preview)

---

## Implementation order (preview for `/plan`)

Suggested vertical slices:

1. **ADR draft** — web move/rename API, event semantics
2. **Backend `file_ops.go`** — extract MoveFile, RenameEntry; refactor MacroOps; tests
3. **FileRenamed event** — publisher + plugin bridge
4. **FileService + API handlers** — PATCH routes, error codes, handler tests
5. **Frontend prefs hook** — localStorage schema + defaults
6. **Explorer shell** — layout, toolbar, tree (lazy), breadcrumb wiring
7. **List view** — columns, resize/reorder, context menu
8. **Grid view** — thumbnails + icons
9. **Inline rename** — UI + API integration
10. **DnD move** — internal file drag to folder targets
11. **Multi-file upload queue** — replace UploadZone
12. **File opener registry + PreviewPanel** — PDF, text, image; inline content disposition
13. **VolumeDetailPage cleanup** — remove old components
14. **Manual UX1–UX9 pass** — browser verification

---

## Decisions log

| # | Decision | Status |
|---|---|---|
| D1 | Drive/Explorer hybrid layout with lazy sidebar tree | ✅ From IDEAS.md |
| D2 | List + grid views; prefs in localStorage v1 | ✅ From IDEAS.md |
| D3 | Customizable list columns (name required) | ✅ From IDEAS.md |
| D4 | Inline rename files + folders | ✅ From IDEAS.md |
| D5 | Drag file → folder move (not folder drag) | ✅ Proposed v1 scope |
| D6 | Multi-file upload on contents pane drop | ✅ Proposed (IDEAS dependency) |
| D7 | Unified `file_ops.go` shared by FileService + MacroOps | ✅ Proposed |
| D8 | `file.renamed` when parent unchanged; `file.moved` when directory changes | ✅ Proposed |
| D9 | No new tree API — reuse list per path | ✅ Proposed |
| D10 | Bulk multi-select deferred to v1.1 | ✅ Proposed |
| D11 | `@dnd-kit/core` + `@dnd-kit/utilities` for DnD | ✅ Resolved (OQ1 — operator) |
| D12 | Double-click → preview via opener registry; PDF + text + image in v1 | ✅ Resolved (OQ2 — operator) |
| D13 | Video/audio preview via plugins later (FFmpeg example); not in core v1 | ✅ Resolved (OQ2 — operator) |
| D14 | `FileOpener` registry is the stable extension point for new types | ✅ Proposed |
| D15 | Upload queue: max 3 parallel, FIFO for remaining files | ✅ Resolved (OQ3 — operator) |

---

## Open Questions

| # | Question | Status |
|---|---|---|
| OQ1 | **Drag-and-drop library** | ✅ Resolved — `@dnd-kit/core` + `@dnd-kit/utilities` |
| OQ2 | **Double-click on file** — download vs preview | ✅ Resolved — opener registry; PDF/text/image in v1; fallback download |
| OQ3 | **Upload concurrency limit** | ✅ Resolved — max 3 parallel + FIFO queue for overflow |
| OQ4 | **Server-side layout prefs** | ✅ Deferred post-v1 (IDEAS.md) |

### OQ1 — Drag-and-drop library ✅ Resolved

**Decision:** `@dnd-kit/core` + `@dnd-kit/utilities` (operator approval).

---

### OQ2 — Double-click / file open ✅ Resolved

**Decision:** In-explorer preview via **`FileOpener` registry**.

| Scope | Types | Behavior |
|---|---|---|
| **v1 built-in** | PDF, text (`text/*` + common extensions), images (`image/*`) | Open in `PreviewPanel` inside explorer |
| **v1 fallback** | Everything else | Download on double-click |
| **Post-v1 plugins** | e.g. `video/*` via FFmpeg transcode | Plugin registers opener; core explorer unchanged |

Registry + `PreviewPanel` ship in v1 so plugins only add handlers — no second explorer redesign.

---

### OQ3 — Upload concurrency limit ✅ Resolved

**Decision:** Max **3 uploads in parallel**. Additional files are **queued (FIFO)** — not rejected, not uploaded all at once.

| Behavior | Detail |
|---|---|
| Parallel cap | 3 active `POST /files` at any time |
| Overflow | Files 4+ wait in queue until a slot opens |
| Order | First dropped/selected → first uploaded (FIFO) |
| On complete | Next queued file starts automatically |

---

## Next step

1. All OQ resolved — spec ready for `/plan`.  
2. On approval → `/plan` to produce `tasks/plan.md` and `tasks/todo.md`.  
3. On build → ADR in `docs/adr/` before merge (file ops API + file opener extension model).
