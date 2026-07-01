# Startup Plan — lcloud

> ⚠️ **Temporary document.** This file holds the startup plan from Phase 0 to MVP.
> Once MVP is reached, **DELETE this file** — git log preserves the history.
> The project then continues via `/spec` on individual features with agent-skills.
>
> Do NOT extend this file with post-MVP phases. Future ideas go into [IDEAS.md](./IDEAS.md).
> Major architectural decisions go into `docs/adr/` via the `documentation-and-adrs` skill.

---

## Phase 0 — Technical foundations

**Objective:** Set up the complete technical environment before any visible functional development.

**Contents:**
- Go module init, directory structure (`cmd/`, `internal/`, `pkg/`, `web/`, `plugins/`)
- Docker Compose: Go service + PostgreSQL 17
- GORM setup + migrations: `User`, `Volume` base models
- JWT auth: register/login endpoints, middleware (golang-jwt/jwt v5), admin-only user creation
- React + Vite + TypeScript + Tailwind CSS + shadcn/ui configured, `design/DESIGN.md` integrated
- Navigation skeleton (React Router, empty but functional routes, login page, empty dashboard shell)
- `.env.example` with all required variables
- Initial deployment on local Docker

**End criterion:** The app runs in Docker, login works, empty dashboard loads, all base models exist.

---

## Phase 1.1 — Volume management

**Objective:** Implement the core volume primitive — creation, storage, and file operations.

**What a user can do at the end:** Create a volume on a chosen disk, upload a file to it, browse and download files from `./userdata`, see volume metadata in the UI.

**Technical contents:**
- Volume CRUD (create, list, rename, delete) with physical disk selection
- Physical volume structure created on disk at volume creation:
  ```
  ./ (volume root — identified by UUID)
  ├── .volume.json          ← source of truth (id, name, owner_id, quota_bytes, filters,
  │                           encryption stub, created_at, disk_path)
  ├── userdata/             ← user-accessible files
  ├── cache/
  │   ├── index/            ← Bleve index files (metadata-only for MVP)
  │   ├── thumbnails/       ← image/video thumbnails for UI
  │   └── metadata/         ← file metadata cache (MIME, hash, size, dates)
  ├── plugins/
  │   └── <plugin-id>/      ← isolated per-plugin data (created on plugin install)
  └── logs/                 ← planned structure, populated post-MVP
  ```
- `.volume.json` as source of truth — portable: lcloud can re-import a volume by pointing to its root directory
- PostgreSQL caches volume config for performance (synced from `.volume.json`)
- File upload/download to `./userdata`
- File type filters per volume (allowed/blocked extensions)
- Quota enforcement (max size in bytes per volume)
- `VolumeIndexer` interface + `BleveIndexer` implementation (indexes: name, path, MIME, size, dates, SHA256 hash)
- Thumbnail generation for images
- Encryption stub in `.volume.json` (`"encryption": {"enabled": false, "method": null}` — implementations post-MVP)
- Multi-user isolation: users access only their own volumes; admin sees all

**End criterion:** User creates a volume on a selected disk, uploads a file, sees it in the UI, downloads it back. Upload rejected when it doesn't match the volume's file filter.

---

## Phase 1.2 — Monitoring

**Objective:** Give users visibility into their volumes — space usage, content breakdown, multi-disk overview.

**What a user can do at the end:** Open the monitoring dashboard and see real-time space stats, file type distribution, and a unified list of all volumes across all disks.

**Technical contents:**
- Volume space stats (total / used / free per volume and per disk)
- File type breakdown (MIME groups, proportions, file count per type) — computed from `./cache/metadata/`
- Unified volume list across multiple physical disks
- Physical disk abstraction toggle: option to mask disk names — show only available space instead of disk identifiers
- Bleve-powered search within a volume (filename, metadata filters: type, size range, date range)
- Stats cached in `./cache/metadata/` and refreshed by `compute_stats` macro (Phase 1.4)

**End criterion:** Monitoring dashboard shows real data from a live volume. Multi-disk unified list visible. Disk name masking toggle works.

---

## Phase 1.3 — Plugin system

**Objective:** Build the foundational plugin infrastructure — event bus, plugin runtime, one working external plugin.

**What a user can do at the end:** Drop a plugin binary into the plugins directory, see it appear in the plugin manager UI, and observe it reacting to a volume event.

**Technical contents:**
- Plugin Go interface definition
- go-plugin (HashiCorp) integration — subprocess communication via stdin/stdout, any language that compiles to a binary
- Plugin registry: scan a `plugins/` directory, load and validate binaries at startup
- Event bus: volume emits typed events — `file.uploaded`, `file.deleted`, `file.moved`, `volume.created`, `volume.deleted`, `task.executed`
- Plugin subscribes to events via the bus; plugin may emit custom events other plugins can subscribe to
- Plugin data isolation: `./plugins/<plugin-id>/` created per volume on plugin install
- Plugin manager UI: list installed plugins, status (running/stopped), enable/disable
- Example plugin: `file-type-validator` (subscribes to `file.uploaded`, validates against volume filters, emits `validation.failed` on mismatch)
- Plugin SDK stub: minimal Go package + `docs/plugins/README.md`

**End criterion:** External plugin binary runs as subprocess, subscribes to `file.uploaded`, acts on it, result visible in UI or logs.

---

## Phase 1.4 — Task system

**Objective:** Build the task scheduler and connect it to predefined macros operating on volumes.

**What a user can do at the end:** Create a periodic task on a volume (e.g. "delete files older than 30 days, every Sunday at 2am") and see it execute on schedule with a result in task history.

**Technical contents:**
- Task model: type (macro name), parameters (typed), scope (volume-scoped or global), schedule (cron expression or fixed interval), enabled flag
- PostgreSQL persistence + gocron: tasks loaded at startup, schedules registered
- Task execution engine: resolves macro type, runs with typed parameters against the target volume
- Predefined macro vocabulary (closed for MVP — changes require a `/spec` cycle):
  - **Cleanup:**
    - `delete_old_files` — delete files not modified in the last N days (optional extension filter)
    - `delete_large_files` — delete files larger than N MB (optional extension filter)
    - `clear_cache` — wipe thumbnails and metadata cache for the volume
  - **Organization:**
    - `move_files` — move files matching a glob pattern to a target subfolder in `./userdata`
    - `sort_by_type` — auto-sort files into type-based subfolders (`images/`, `documents/`, `videos/`, `audio/`, `other/`)
    - `sort_by_date` — auto-sort files into `YYYY/MM/` subfolders based on modification date
  - **Maintenance:**
    - `rebuild_index` — force full Bleve reindex of the volume
    - `compute_stats` — force recalculation of volume statistics
  - **Alerts:**
    - `alert_usage` — write a log entry and emit a `volume.alert.usage` event when volume usage exceeds X%
- Task CRUD in the web UI (create, list, enable/disable, delete, view next scheduled run)
- Task execution history (last run time, result, count of affected files)

**End criterion:** Periodic task created in UI, executes on schedule, result visible in task history and reflected on the volume (e.g. files deleted or moved as expected).

---

## 🎯 MVP reached at end of Phase 1.4

lcloud MVP: a single user (or a family on a local network) can deploy lcloud via Docker Compose, create volumes on their own disks, upload and manage files, monitor volume usage across multiple disks, install a plugin that reacts to volume events, and automate recurring file operations via scheduled tasks — all through a web interface, with zero external cloud dependency.

---

## Use cases per phase

### Phase 0
- UC0.1: User clones the repo, copies `.env.example`, runs `docker compose up`, opens `http://localhost:8080`, sees the login page.
- UC0.2: User logs in with admin credentials, lands on the empty dashboard skeleton.

### Phase 1.1
- UC1.1: User creates a volume "Photos" on `/mnt/disk1`, sets a 50 GB quota and a `.jpg .png .webp` extension filter.
- UC1.2: User uploads a photo — it appears in the file browser. User downloads it back.
- UC1.3: User tries to upload a `.mp4` — upload is rejected by the filter with a clear error.

### Phase 1.2
- UC2.1: User opens the monitoring dashboard, sees "Photos" at 12 GB / 50 GB with images at 100%.
- UC2.2: User adds a "Documents" volume on a second disk — both volumes appear in the unified list.
- UC2.3: User enables disk name masking — the UI shows available space only, not disk identifiers.
- UC2.4: User searches "vacation" — finds matching filenames from the Bleve index.

### Phase 1.3
- UC3.1: User drops the `file-type-validator` binary into `plugins/`, restarts lcloud — the plugin appears in the plugin manager UI as "running".
- UC3.2: User uploads a file to a filtered volume — the plugin catches `file.uploaded`, validates it, logs a notice visible in the UI.

### Phase 1.4
- UC4.1: User creates a task "Delete images older than 90 days" on the Photos volume, scheduled every Sunday at 2am.
- UC4.2: The task runs (or is manually triggered), deletes old files, execution appears in task history with the count of deleted files.

---

## Open assumptions

- ASSUMPTION: Users manage their own Docker installation and are responsible for exposing the service on their local network or VPS. lcloud does not handle reverse proxy setup or TLS termination.
- ASSUMPTION: Disk paths used for volume storage are pre-mounted and accessible at the paths specified during volume creation.

---

## How to use this plan

For each phase, in the IDE with agent-skills installed:

1. `/spec [phase description]` — the agent reads STARTUP.md + VISION.md + docs/project.md for context
2. `/plan` — the agent breaks the phase into verifiable tasks (`tasks/plan.md`)
3. `/build` — incremental implementation, slice by slice
4. `/test` — verification with test-driven-development
5. `/review` — code-review-and-quality + security-and-hardening
6. `/ship` — when the phase is deliverable

Then move to the next phase.

---

## ⚠️ Final reminder

This file will be **deleted** once the MVP is reached. Its role is to carry the project bootstrap, not to become a history document. Git log keeps the phase history. Lasting architectural decisions go into `docs/adr/`. Long-term vision lives in VISION.md.
