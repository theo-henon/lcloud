# Spec: Phase 1.1 — Volume management

> **Status:** Draft — pending review
> **Scope:** Core volume primitive — creation, storage, file operations, indexing, thumbnails.
> **Prerequisite:** Phase 0 complete (auth, Docker, UI skeleton, Volume GORM model schema-only).
> **Sources:** [STARTUP.md](./STARTUP.md#phase-11--volume-management), [VISION.md](./VISION.md), [docs/project.md](./docs/project.md)

---

## Assumptions (correct me now or I proceed)

1. **Phase 0 is shipped** — JWT auth, Docker Compose, React shell, and the base `Volume` GORM model exist; Phase 1.1 extends them.
2. **Multiple physical disks supported** — each disk is a separate host mount mapped into the container (e.g. SSD + two HDDs). lcloud does not mount disks itself; the deployer configures Docker volume mounts ([STARTUP.md](./STARTUP.md#open-assumptions)).
3. **Disk registry via `STORAGE_DISK_PATHS`** — comma-separated list of absolute paths inside the container (e.g. `/data/disks/ssd,/data/disks/hdd1,/data/disks/hdd2`). Each entry is one physical disk. `GET /api/disks` exposes this list with free/total space. Fallback for local dev: if unset, scan immediate subdirectories of `STORAGE_BASE_PATH`.
4. **Volume created on one chosen disk** — at creation the user picks one disk from the registry; volumes on different disks coexist independently (multi-disk is a first-class requirement, not a Phase 1.2 feature).
5. **Volume root directory name = volume UUID** — `{disk_path}/{volume_id}/`; display name lives in `.volume.json` only.
6. **Filter mode is exclusive** — each volume uses either an **allowlist** or a **blocklist**, not both simultaneously.
7. **Extensions are normalized** — stored lowercase with leading dot (`.jpg`); comparison is case-insensitive.
8. **Quota `0` = unlimited** — consistent with common self-hosted conventions; documented in UI.
9. **Image thumbnails only** — JPEG, PNG, WebP, GIF; no video thumbnails in Phase 1.1 ([STARTUP.md](./STARTUP.md#phase-11--volume-management)).
10. **No event bus yet** — plugin events arrive in Phase 1.3; volume/file services expose clean hooks but do not emit bus events in 1.1.
11. **No volume re-import UI** — `.volume.json` portability is enforced by implementation; explicit re-import flow is post-MVP (structure must remain valid).
12. **No Bleve search UI** — indexer runs on upload/delete; search UI is Phase 1.2; `VolumeIndexer.Search` is implemented and tested but not exposed via REST in 1.1.
13. **Single-file upload per request** — multipart upload of one file; batch upload deferred.
14. **Max upload body size** — configurable via env `MAX_UPLOAD_BYTES` (default `100MB`); separate from volume quota.
15. **Delete protection** — regular users cannot delete a volume that still contains files; admin may force-delete the entire volume tree with explicit confirmation.
16. **Subfolders in UI** — users can create directories inside `userdata/` from the file browser.

---

## Objective

Implement the core storage primitive: users create volumes on chosen disks, upload files to `./userdata`, browse and download them, and see volume metadata in the UI. Uploads are rejected when they violate the volume's extension filter or quota.

**Who:** Self-hosters who already run lcloud via Docker and understand that volumes are real directories on their disks.

**User stories:**

| ID | Story |
|---|---|
| UC1.1 | As a user, I create a volume "Photos" on disk `disk1`, set a 50 GB quota and an allowlist filter `.jpg .png .webp`. |
| UC1.2 | As a user, I upload a photo — it appears in the file browser. I download it back and get identical bytes. |
| UC1.3 | As a user, I try to upload a `.mp4` to a JPG-only volume — upload is rejected with a clear error. |
| UC1.4 | As a user, I list my volumes, rename one, and delete an empty volume. |
| UC1.5 | As a user, I only see my own volumes; as admin, I see all volumes. |
| UC1.6 | As a user, when I upload an image, a thumbnail is generated and visible in the file browser. |

**Out of scope for Phase 1.1:**

- Monitoring dashboard, space stats UI, multi-disk unified list (Phase 1.2)
- Bleve search API and search UI (Phase 1.2)
- Plugin system and event bus (Phase 1.3)
- Task scheduler and macros (Phase 1.4)
- Volume re-import wizard
- File move/rename within volume (defer unless trivial; not in end criterion)
- Video thumbnail generation
- Encryption implementation (stub only in `.volume.json`)
- Content search (metadata-only indexing)

---

## Tech Stack

See [docs/project.md — Tech Stack](./docs/project.md#tech-stack).

Phase 1.1 adds:

| Layer | Addition |
|---|---|
| Backend | Bleve v2 (`VolumeIndexer`), Go stdlib `image` for thumbnails *(see OQ3 — pending confirmation)* |
| Frontend | TanStack Query for volume/file data; file upload via `FormData` |

---

## Commands

### Prerequisites

Same as Phase 0 — Go 1.25+, Node 20+, Docker Compose.

### Development

```bash
# Backend tests (volume + indexer focus)
go test ./internal/volume/... ./internal/indexer/... -cover
go test ./... -cover

# Frontend tests
cd web && npm run test

# Local stack with multiple disk mounts (example)
cp .env.example .env
mkdir -p ./data/disks/ssd ./data/disks/hdd1 ./data/disks/hdd2
docker compose up -d --build
docker compose logs -f app
```

### Verification (Phase 1.1 done)

```bash
# 1. Login and capture token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"<ADMIN_EMAIL>","password":"<ADMIN_PASSWORD>"}' \
  | jq -r '.access_token')

# 2. List available disks
curl -s http://localhost:8080/api/disks \
  -H "Authorization: Bearer $TOKEN"

# 3. Create volume
VOL=$(curl -s -X POST http://localhost:8080/api/volumes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Photos","disk_path":"/data/storage/disk1","quota_bytes":53687091200,"filters":{"mode":"allow","extensions":[".jpg",".png",".webp"]}}')
VOL_ID=$(echo "$VOL" | jq -r '.id')

# 4. Upload allowed file
curl -s -X POST "http://localhost:8080/api/volumes/$VOL_ID/files" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@./test.jpg" \
  -F "path=."

# 5. Reject disallowed file (expect 422)
curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:8080/api/volumes/$VOL_ID/files" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@./test.mp4" \
  -F "path=."
# expect 422

# 6. List files
curl -s "http://localhost:8080/api/volumes/$VOL_ID/files?path=." \
  -H "Authorization: Bearer $TOKEN"

# 7. Download
curl -s "http://localhost:8080/api/volumes/$VOL_ID/files/content?path=test.jpg" \
  -H "Authorization: Bearer $TOKEN" -o /tmp/downloaded.jpg
cmp test.jpg /tmp/downloaded.jpg  # expect identical
```

---

## Project Structure

See [docs/project.md — Project structure](./docs/project.md#project-structure).

Phase 1.1 creates or extends:

```
lcloud/
├── internal/
│   ├── volume/
│   │   ├── model.go                 ← extend GORM model (filters JSON, root_path)
│   │   ├── volume_json.go           ← .volume.json read/write (source of truth)
│   │   ├── service.go               ← CRUD, sync PG ↔ .volume.json
│   │   ├── file_service.go          ← upload, download, list, delete
│   │   ├── disk.go                  ← disk discovery under STORAGE_BASE_PATH
│   │   ├── filter.go                ← extension allow/block validation
│   │   ├── quota.go                 ← usage calculation + enforcement
│   │   ├── thumbnail.go             ← image thumbnail generation
│   │   ├── metadata_cache.go        ← ./cache/metadata/ per-file JSON
│   │   ├── paths.go                 ← safe path resolution (anti-traversal)
│   │   └── *_test.go
│   ├── indexer/
│   │   ├── indexer.go               ← VolumeIndexer interface
│   │   ├── bleve.go                 ← BleveIndexer implementation
│   │   ├── types.go                 ← FileMetadata, SearchQuery
│   │   └── bleve_test.go
│   └── api/
│       ├── volume_handler.go
│       ├── file_handler.go
│       ├── disk_handler.go
│       └── router.go                ← register new routes
├── web/src/
│   ├── pages/
│   │   ├── VolumesPage.tsx          ← list + create volume
│   │   └── VolumeDetailPage.tsx     ← file browser + upload
│   ├── components/volumes/          ← VolumeCard, CreateVolumeForm, FileBrowser, UploadZone
│   ├── hooks/
│   │   ├── useVolumes.ts
│   │   └── useVolumeFiles.ts
│   └── lib/api.ts                   ← volume/file API methods
└── .env.example                     ← add MAX_UPLOAD_BYTES
```

On volume creation, the service creates this tree on disk:

```
{disk_path}/{volume_uuid}/
├── .volume.json
├── userdata/
├── cache/
│   ├── index/
│   ├── thumbnails/
│   └── metadata/
├── plugins/
└── logs/
```

---

## Data models

### `.volume.json` (source of truth)

Written atomically on every config change. PostgreSQL is synced after a successful disk write.

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Photos",
  "owner_id": "660e8400-e29b-41d4-a716-446655440001",
  "quota_bytes": 53687091200,
  "filters": {
    "mode": "allow",
    "extensions": [".jpg", ".png", ".webp"]
  },
  "encryption": {
    "enabled": false,
    "method": null
  },
  "created_at": "2026-07-01T12:00:00Z",
  "disk_path": "/data/storage/disk1"
}
```

**Rules:**

- Any config mutation → update `.volume.json` first → then upsert PostgreSQL row.
- Volume delete → remove directory tree from disk → delete PG row → close Bleve index.
- `id` and `disk_path` are immutable after creation.

### GORM `Volume` (PostgreSQL cache)

Extends Phase 0 model:

| Field | Type | Notes |
|---|---|---|
| `id` | UUID (PK) | Same as `.volume.json` id |
| `name` | string | Display name (rename updates `.volume.json`) |
| `owner_id` | UUID (FK) | Owner |
| `disk_path` | string | Parent directory (immutable) |
| `root_path` | string | `{disk_path}/{id}` — denormalized for queries |
| `quota_bytes` | int64 | `0` = unlimited |
| `filters` | JSONB | `{ mode, extensions }` |
| `used_bytes` | int64 | Cached aggregate; refreshed on upload/delete |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |

### File metadata cache (`./cache/metadata/{file_id}.json`)

```json
{
  "id": "file-uuid",
  "name": "vacation.jpg",
  "relative_path": "vacation.jpg",
  "mime_type": "image/jpeg",
  "size_bytes": 2048576,
  "sha256": "abc123…",
  "modified_at": "2026-07-01T12:05:00Z",
  "thumbnail_path": "cache/thumbnails/file-uuid.jpg"
}
```

### Bleve document (via `FileMetadata`)

Indexed fields: `name`, `relative_path`, `mime_type`, `size_bytes`, `modified_at`, `sha256`.

---

## API (Phase 1.1)

Base path: `/api`. All endpoints require JWT unless noted.

**Error shape** (unchanged from Phase 0):

```json
{ "error": "human-readable message", "code": "FILE_FILTER_REJECTED" }
```

### Disks

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/disks` | JWT | List registered physical disks (`STORAGE_DISK_PATHS`) with space stats |

**Response:**

```json
{
  "disks": [
    {
      "path": "/data/disks/ssd",
      "name": "ssd",
      "label": "SSD",
      "total_bytes": 1000000000000,
      "free_bytes": 800000000000
    },
    {
      "path": "/data/disks/hdd1",
      "name": "hdd1",
      "label": "HDD 1",
      "total_bytes": 4000000000000,
      "free_bytes": 3200000000000
    }
  ]
}
```

**Disk registry:**

- Primary: `STORAGE_DISK_PATHS` — comma-separated absolute paths, each mapped to a distinct physical disk in `docker-compose.yml`.
- Each path must exist, be a directory, and be writable; invalid entries are skipped with a startup warning.
- Display `name` = last path segment; optional `label` from env `STORAGE_DISK_LABELS` (same order, comma-separated) for human-friendly UI labels.
- Fallback (local dev only): if `STORAGE_DISK_PATHS` is empty, scan immediate subdirectories of `STORAGE_BASE_PATH`.

**Example `docker-compose.yml` mounts:**

```yaml
volumes:
  - ${DISK_SSD:-./data/disks/ssd}:/data/disks/ssd
  - ${DISK_HDD1:-./data/disks/hdd1}:/data/disks/hdd1
  - ${DISK_HDD2:-./data/disks/hdd2}:/data/disks/hdd2
environment:
  STORAGE_DISK_PATHS: /data/disks/ssd,/data/disks/hdd1,/data/disks/hdd2
  STORAGE_DISK_LABELS: SSD,HDD 1,HDD 2
```

### Volumes

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/volumes` | JWT | List volumes (owner-scoped; admin sees all) |
| `POST` | `/volumes` | JWT | Create volume |
| `GET` | `/volumes/:id` | JWT | Get volume detail + usage |
| `PATCH` | `/volumes/:id` | JWT | Rename, update quota/filters (owner or admin) |
| `DELETE` | `/volumes/:id` | JWT | Delete volume (see delete rules below) |

**Create request:**

```json
{
  "name": "Photos",
  "disk_path": "/data/storage/disk1",
  "quota_bytes": 53687091200,
  "filters": {
    "mode": "allow",
    "extensions": [".jpg", ".png", ".webp"]
  }
}
```

**Volume response (representative fields):**

```json
{
  "id": "…",
  "name": "Photos",
  "owner_id": "…",
  "disk_path": "/data/storage/disk1",
  "quota_bytes": 53687091200,
  "used_bytes": 2048576,
  "filters": { "mode": "allow", "extensions": [".jpg", ".png", ".webp"] },
  "created_at": "…",
  "updated_at": "…"
}
```

**Authorization:**

- Regular user: CRUD only on volumes where `owner_id = sub`.
- Admin: read/write all volumes.

**Delete rules (validated OQ1):**

| Actor | Volume state | Behavior |
|---|---|---|
| Regular user | Contains files | **Rejected** — `422 VOLUME_NOT_EMPTY` |
| Regular user | Empty (`userdata/` has no files) | Delete allowed — removes disk tree + PG row |
| Admin | Contains files | Delete with `?force=true` — recursive wipe of entire volume directory after UI confirmation |
| Admin | Empty | Delete allowed (no `force` required) |

**Validation errors (422):**

| Code | Condition |
|---|---|
| `DISK_NOT_FOUND` | `disk_path` not under allowed storage roots |
| `DISK_NOT_WRITABLE` | Target disk not writable |
| `VOLUME_NAME_TAKEN` | Same owner already has a volume with that name |
| `INVALID_FILTER` | Unknown mode or empty extensions when mode is set |
| `QUOTA_EXCEEDED` | Upload would exceed `quota_bytes` |
| `FILE_FILTER_REJECTED` | Extension not allowed / blocked |
| `PATH_TRAVERSAL` | Resolved path escapes `userdata/` |
| `VOLUME_NOT_EMPTY` | Non-admin delete rejected when files remain |

### Files

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/volumes/:id/files` | JWT | List directory contents (`?path=` relative to `userdata/`, default `.`) |
| `POST` | `/volumes/:id/files/directories` | JWT | Create subdirectory (`{ "path": "2024/vacation" }`) |
| `POST` | `/volumes/:id/files` | JWT | Upload file (`multipart/form-data`: `file`, optional `path` target dir) |
| `GET` | `/volumes/:id/files/content` | JWT | Download file (`?path=`) |
| `GET` | `/volumes/:id/files/thumbnail` | JWT | Thumbnail for image (`?path=`); 404 if not an image |
| `DELETE` | `/volumes/:id/files` | JWT | Delete file (`?path=`) |

**List response:**

```json
{
  "path": ".",
  "entries": [
    {
      "name": "vacation.jpg",
      "path": "vacation.jpg",
      "type": "file",
      "size_bytes": 2048576,
      "mime_type": "image/jpeg",
      "modified_at": "…",
      "has_thumbnail": true
    },
    {
      "name": "2024",
      "path": "2024",
      "type": "directory"
    }
  ]
}
```

**Upload flow (service layer):**

1. Resolve and validate path (no traversal).
2. Validate extension against volume filter.
3. Check quota (`used_bytes + file_size <= quota_bytes`; skip if quota is 0).
4. Write file to `userdata/`.
5. Compute SHA256, detect MIME, write metadata cache JSON.
6. Generate thumbnail if image.
7. Index via `VolumeIndexer.Index`.
8. Update `used_bytes` in `.volume.json` + PostgreSQL.

---

## Frontend (Phase 1.1)

Follow [design/DESIGN.md](./design/DESIGN.md): dark canvas, yellow primary CTAs, card surfaces, stat numbers in yellow where relevant.

### Routes

| Path | Access | Content |
|---|---|---|
| `/volumes` | Protected | Volume list + "Create volume" action |
| `/volumes/:id` | Protected | Volume detail: metadata header + file browser |

Remove the "Soon" badge from the Volumes sidebar item when Phase 1.1 ships.

### `/volumes` — Volume list

- Card grid or table: name, disk, used/quota bar, filter summary, created date.
- Primary CTA: **Create volume** (yellow button).
- Create form (modal or slide-over):
  - Name (required)
  - Disk selector (dropdown from `GET /api/disks`, show free space)
  - Quota (optional, GB input → bytes; empty = unlimited)
  - Filter mode toggle: Allow / Block
  - Extensions input (comma-separated or tag input → normalized)
- Row actions: Open, Rename, Delete (with confirmation; non-admin blocked if volume not empty).
- Admin delete on non-empty volume: second confirmation step warning that all files will be permanently deleted.

### `/volumes/:id` — File browser

- Breadcrumb navigation within `userdata/`.
- **New folder** action — creates a subdirectory in the current path (validated OQ2).
- File table: name, size, modified date, thumbnail preview for images.
- Upload zone (drag-and-drop + file picker) respecting current directory path.
- Download action per file.
- Clear inline error when filter rejects upload (red accent per design system).
- Empty state when volume has no files.

### Data fetching

- TanStack Query hooks: `useVolumes`, `useVolume`, `useVolumeFiles`, `useDisks`.
- Optimistic invalidation after upload/delete/create.
- Upload progress indicator (basic percentage if feasible; otherwise spinner).

---

## Code Style

### Go — volume service pattern

Handlers stay thin; all disk I/O and business rules in `internal/volume/`.

```go
// internal/volume/file_service.go — upload validates before write
func (s *FileService) Upload(ctx context.Context, vol *Volume, relPath string, r io.Reader, size int64) (*FileEntry, error) {
    absPath, err := s.paths.ResolveUserdata(vol.RootPath, relPath)
    if err != nil {
        return nil, err
    }
    if err := s.filters.Validate(vol.Filters, filepath.Base(absPath)); err != nil {
        return nil, err
    }
    if err := s.quota.Check(vol, size); err != nil {
        return nil, err
    }
    // write → metadata → thumbnail → indexer.Index → sync usage
}
```

- All path operations go through `paths.go` — never concatenate user input into filesystem paths directly.
- Bleve accessed only through `internal/indexer/` — never import Bleve outside that package.
- Atomic `.volume.json` writes: write temp file + rename.

### TypeScript

- Page components orchestrate; presentational components in `components/volumes/`.
- Upload uses `FormData` via `api.uploadFile()` helper (not JSON body).
- Display bytes with human-readable formatter (shared util).

---

## Testing Strategy

See [docs/project.md — Coverage targets](./docs/project.md#coverage-targets).

| Layer | Focus | Location |
|---|---|---|
| Backend unit | Filter validation, path traversal rejection, quota math, `.volume.json` sync | `internal/volume/*_test.go` |
| Backend unit | BleveIndexer Index/Delete/Search/Rebuild | `internal/indexer/bleve_test.go` |
| Backend integration | Volume CRUD + upload/download round-trip (temp dir as disk) | `internal/volume/service_test.go`, `internal/api/*_test.go` |
| Frontend unit | CreateVolumeForm validation, file browser render, upload error display | `web/src/**/*.test.tsx` |

**Coverage target:** 80% minimum on `internal/volume/` and `internal/indexer/`.

**Verify before ship:**

```bash
go test ./internal/volume/... ./internal/indexer/... -cover
go test ./...
cd web && npm run test
# Manual UC1.1 – UC1.3 via UI + curl script above
```

---

## Boundaries

### Always

- Write volume config to `.volume.json` before PostgreSQL.
- Route all indexing through `VolumeIndexer` interface.
- Validate paths — reject `..`, absolute paths, and symlinks escaping `userdata/`.
- Scope volume list/detail to owner; admin bypass explicit in service layer (not only in handlers).
- Follow [design/DESIGN.md](./design/DESIGN.md) for UI.
- Run tests before marking phase complete.

### Ask first

- Adding dependencies not in [approved list](./docs/project.md#approved-third-party-libraries) (only if OQ3 chooses an external imaging library).
- Changing `.volume.json` schema fields (requires ADR).
- Changing volume directory structure (requires ADR).
- Exposing Bleve search via REST before Phase 1.2 scope review.

### Never

- Write volume config exclusively to PostgreSQL.
- Import Bleve outside `internal/indexer/`.
- Emit plugin bus events directly from handlers (Phase 1.3).
- Allow path traversal in file operations.
- Skip filter/quota checks on upload.

---

## Success Criteria

Phase 1.1 is **done** when all of the following pass:

- [ ] **SC1.1** `GET /api/disks` returns all disks from `STORAGE_DISK_PATHS` with correct free/total space
- [ ] **SC1.1b** Volumes can be created on different disks; each volume lives under its chosen disk path
- [ ] **SC1.2** `POST /api/volumes` creates on-disk structure + `.volume.json` + PostgreSQL row
- [ ] **SC1.3** `.volume.json` fields match PostgreSQL after create and after PATCH
- [ ] **SC1.4** User lists only own volumes; admin lists all
- [ ] **SC1.5** Upload to `userdata/` succeeds; file appears in `GET /api/volumes/:id/files`
- [ ] **SC1.6** Download returns identical bytes to uploaded file
- [ ] **SC1.7** Upload with disallowed extension returns 422 + `FILE_FILTER_REJECTED`
- [ ] **SC1.8** Upload exceeding quota returns 422 + `QUOTA_EXCEEDED`
- [ ] **SC1.9** Path traversal attempts (`../etc/passwd`) rejected with 422
- [ ] **SC1.10** Image upload generates thumbnail served by thumbnail endpoint
- [ ] **SC1.11** Bleve index contains uploaded file metadata; delete removes index entry
- [ ] **SC1.12** Rename volume updates `.volume.json` name without changing directory UUID
- [ ] **SC1.13** Empty volume delete removes disk tree and PostgreSQL row
- [ ] **SC1.13b** Non-admin delete of non-empty volume returns `VOLUME_NOT_EMPTY`
- [ ] **SC1.13c** Admin force-delete removes non-empty volume tree entirely
- [ ] **SC1.13d** User can create a subfolder from the file browser UI
- [ ] **SC1.14** UI: create volume flow (UC1.1), file browser + upload + download (UC1.2), filter error (UC1.3)
- [ ] **SC1.15** Volumes sidebar item active without "Soon" badge
- [ ] **SC1.16** `go test ./...` and `cd web && npm run test` pass
- [ ] **SC1.17** `.env.example` documents `MAX_UPLOAD_BYTES`, `STORAGE_DISK_PATHS`, `STORAGE_DISK_LABELS`

---

## Implementation order (preview for `/plan`)

Suggested vertical slices — `/plan` will expand into `tasks/plan.md`:

1. **Volume JSON + paths** — schema, atomic read/write, safe path resolver
2. **Disk registry** — `STORAGE_DISK_PATHS` config, multi-mount docker-compose, API endpoint
3. **Volume CRUD service** — create/list/get/patch/delete with PG sync
4. **Volume API handlers** — REST endpoints + auth scoping
5. **Filter + quota** — validation primitives + tests
6. **File service** — upload, list, mkdir, download, delete
7. **Metadata cache + thumbnails** — sidecar JSON + image resize
8. **VolumeIndexer + BleveIndexer** — interface, impl, wired on upload/delete
9. **File API handlers** — multipart upload, content/thumbnail download
10. **Frontend volumes list** — disks fetch, create form, volume cards
11. **Frontend file browser** — browse, new folder, upload, download, errors
12. **Integration** — Docker end-to-end UC1.1–UC1.3

---

## Decisions log

| # | Decision | Status |
|---|---|---|
| D1 | Filter mode: exclusive allow **or** block per volume | ✅ Validated |
| D2 | Quota `0` = unlimited | ✅ Validated |
| D3 | Volume directory named by UUID (immutable) | ✅ Aligned with docs/project.md |
| D4 | Bleve search API deferred to Phase 1.2; indexer built in 1.1 | ✅ Validated |
| D5 | Multi-disk via `STORAGE_DISK_PATHS` — each physical disk independently mounted | ✅ Validated |
| D6 | Delete: user blocked if non-empty; admin may `?force=true` with confirmation | ✅ Validated (OQ1) |
| D7 | Subfolder creation in file browser UI | ✅ Validated (OQ2) |
| D8 | Thumbnail max size 256×256 px, JPEG output | ✅ Validated (OQ4) |

---

## Open Questions

| # | Question | Status |
|---|---|---|
| OQ3 | **Thumbnail engine** — see explanation below | ⏳ Pending |
| OQ5 | **Volume ownership on create** — see explanation below | ⏳ Pending |

### OQ3 — Thumbnail engine (plain language)

When you upload a photo, lcloud generates a **small preview** so the file browser doesn't load full-size images. Two ways to shrink images in Go:

| Option | What it means | Trade-off |
|---|---|---|
| **A — Built-in Go tools** (`image` package) | No extra dependency; resize JPG/PNG/GIF/WebP | Simpler setup; photos taken with phone rotation may appear sideways (no EXIF auto-rotate) |
| **B — Small external library** (`disintegration/imaging`) | Better resize quality + auto-rotation from EXIF | One extra dependency to approve and maintain |

**Recommendation:** Option A for Phase 1.1 (keep it simple). Upgrade to B later if rotated previews become a problem.

### OQ5 — Who owns a new volume? (plain language)

When someone clicks **Create volume**, who is recorded as the owner in `.volume.json`?

| Option | Meaning |
|---|---|
| **A — Always the logged-in user** | Even the admin creates volumes for themselves only. Other users create their own volumes after logging in. |
| **B — Admin can assign an owner** | Admin picks a user in the create form (useful if admin sets up storage for family members who don't manage disks). |

**Recommendation:** Option A for Phase 1.1 (simpler). Option B can be added later without changing the volume structure.

---

## Next step

After spec approval → `/plan` to produce `tasks/plan.md` with ordered, verifiable tasks.
