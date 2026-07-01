# Spec: Phase 1.2 — Monitoring

> **Status:** Draft — pending review
> **Scope:** Volume visibility — space stats, file type breakdown, multi-disk overview, disk name masking, Bleve search.
> **Prerequisite:** Phase 1.1 complete (volume CRUD, file ops, metadata cache, Bleve indexer wired on upload/delete).
> **Sources:** [STARTUP.md](./STARTUP.md#phase-12--monitoring), [VISION.md](./VISION.md), [docs/project.md](./docs/project.md)

---

## Assumptions (correct me now or I proceed)

1. **Phase 1.1 is shipped** — volumes, files, metadata cache, Bleve indexation, `GET /api/disks`, `used_bytes` per volume all work.
2. **`internal/monitoring/` is greenfield** — module referenced in docs but not yet created; Phase 1.2 creates it.
3. **Stats source of truth for breakdown** — file type distribution is computed by scanning `./cache/metadata/*.json` (per-file records written in Phase 1.1); `used_bytes` in `.volume.json` remains the source of truth for volume usage totals.
4. **Aggregated stats are cached on disk** — a single `stats.json` file per volume under `./cache/metadata/` stores precomputed breakdown; invalidated on file upload/delete; recomputed on read or explicit refresh.
5. **`compute_stats` macro is Phase 1.4** — Phase 1.2 implements the stats computation function and a manual refresh API; Phase 1.4 wires the same function into the task scheduler. No gocron dependency in 1.2.
6. **Disk name masking is an admin security setting** — stored server-side in PostgreSQL (`instance_settings.mask_disk_names`); the API masks disk paths and labels for all users when enabled, except on `GET /api/disks` for administrators (volume creation). Configured in **Settings** (`/settings`).
7. **Search is volume-scoped** — Bleve index is per-volume; search UI lives on the monitoring page with a volume selector (UC2.4). No cross-volume search in MVP.
8. **MIME groups match macro vocabulary** — breakdown categories align with Phase 1.4 `sort_by_type` folders: `images`, `documents`, `videos`, `audio`, `archives`, `other`.
9. **Admin sees all volumes/disks** — same owner-scoping rules as Phase 1.1; monitoring endpoints respect volume ownership.
10. **No new Bleve fields** — Phase 1.1 already indexes `name`, `relative_path`, `mime_type`, `size_bytes`, `modified_at`, `sha256`; Phase 1.2 adds filter queries on existing fields only.
11. **Dashboard replaces placeholder** — `/monitoring` becomes the real monitoring page; sidebar "Soon" badge removed when shipped.
12. **Real-time enough for MVP** — stats refresh on page load + after explicit refresh; no WebSocket push. Upload/delete invalidates stats cache; next read recomputes if needed.

---

## Objective

Give users visibility into their storage: space usage per volume and per disk, file type distribution, a unified multi-disk volume overview, optional disk name masking, and filename/metadata search within a volume.

**Who:** Self-hosters running lcloud who manage volumes across one or more physical disks and want a single dashboard to understand what's stored where.

**User stories:**

| ID | Story |
|---|---|
| UC2.1 | As a user, I open the monitoring dashboard and see my "Photos" volume at 12 GB / 50 GB with images at 100% of volume content. |
| UC2.2 | As a user, I add a "Documents" volume on a second disk — both volumes appear in the unified list grouped by disk. |
| UC2.3 | As an admin, I enable disk name masking in Settings — non-admin users see generic labels (e.g. "Storage 1") instead of real disk paths across the app. |
| UC2.4 | As a user, I search "vacation" in my Photos volume — matching filenames appear from the Bleve index. |

**Out of scope for Phase 1.2:**

- Plugin system (Phase 1.3)
- Task scheduler and `compute_stats` cron (Phase 1.4 — function built here, scheduler wired later)
- Cross-volume / global search
- Server-side user preferences persistence
- Infrastructure monitoring (CPU, RAM, container health)
- Historical usage trends / time-series graphs
- Content (full-text) search inside files
- Alerting (`alert_usage` macro — Phase 1.4)
- Changes to `.volume.json` schema or volume directory structure

---

## Tech Stack

See [docs/project.md — Tech Stack](./docs/project.md#tech-stack).

Phase 1.2 adds no new dependencies. Reuses:

| Layer | Reuse |
|---|---|
| Backend | Bleve v2 (`VolumeIndexer.Search` extended), existing `DiskRegistry`, `MetadataCache` |
| Frontend | TanStack Query, Settings page (admin security options), shadcn/ui chart primitives or CSS progress bars |

---

## Commands

### Development

```bash
# Backend tests (monitoring + indexer search focus)
go test ./internal/monitoring/... ./internal/indexer/... -cover
go test ./... -cover

# Frontend tests
cd web && npm run test

# Local stack
docker compose up -d --build
docker compose logs -f app
```

### Verification (Phase 1.2 done)

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"<ADMIN_EMAIL>","password":"<ADMIN_PASSWORD>"}' \
  | jq -r '.access_token')

VOL_ID="<volume-uuid>"

# 1. Monitoring overview (disks + volumes + stats)
curl -s http://localhost:8080/api/monitoring/overview \
  -H "Authorization: Bearer $TOKEN" | jq

# 2. Volume stats + MIME breakdown
curl -s "http://localhost:8080/api/monitoring/volumes/$VOL_ID/stats" \
  -H "Authorization: Bearer $TOKEN" | jq

# 3. Force stats refresh (same logic compute_stats will call in 1.4)
curl -s -X POST "http://localhost:8080/api/monitoring/volumes/$VOL_ID/stats/refresh" \
  -H "Authorization: Bearer $TOKEN" | jq

# 4. Search within volume (filename + filters)
curl -s "http://localhost:8080/api/volumes/$VOL_ID/search?q=vacation&mime_prefix=image/" \
  -H "Authorization: Bearer $TOKEN" | jq

# 5. Search with size/date filters
curl -s "http://localhost:8080/api/volumes/$VOL_ID/search?q=&min_size=1048576&modified_after=2026-01-01T00:00:00Z" \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Project Structure

See [docs/project.md — Project structure](./docs/project.md#project-structure).

Phase 1.2 creates or extends:

```
lcloud/
├── internal/
│   ├── monitoring/
│   │   ├── service.go              ← stats computation, overview aggregation
│   │   ├── stats_cache.go          ← read/write stats.json per volume
│   │   ├── mime_groups.go          ← MIME → category mapping
│   │   ├── types.go                ← VolumeStats, DiskOverview, OverviewResponse
│   │   └── service_test.go
│   ├── indexer/
│   │   ├── indexer.go              ← extend SearchQuery with filters
│   │   ├── bleve.go                ← conjunction query for filters
│   │   └── bleve_test.go           ← add Search filter tests
│   ├── volume/
│   │   └── metadata_cache.go       ← add ListAll(rootPath) for stats scan
│   └── api/
│       ├── monitoring_handler.go   ← overview + volume stats endpoints
│       ├── search_handler.go       ← volume search endpoint
│       └── router.go               ← register new routes
├── web/src/
│   ├── pages/
│   │   └── MonitoringPage.tsx      ← replace PlaceholderPage
│   ├── components/monitoring/
│   │   ├── DiskOverviewCard.tsx
│   │   ├── VolumeStatsCard.tsx
│   │   ├── MimeBreakdownChart.tsx
│   │   ├── UnifiedVolumeList.tsx
│   │   ├── VolumeSearchPanel.tsx
│   │   └── DiskMaskToggle.tsx
│   ├── hooks/
│   │   ├── useMonitoringOverview.ts
│   │   ├── useVolumeStats.ts
│   │   └── useVolumeSearch.ts
│   ├── store/
│   │   └── useSettings.ts          ← instance settings (mask_disk_names)
│   └── lib/api.ts                  ← monitoring + search methods
```

---

## Data models

### Volume stats cache (`./cache/metadata/stats.json`)

Written by `monitoring.Service.ComputeStats`. Invalidated (deleted) on file upload/delete. Portable — travels with the volume.

```json
{
  "computed_at": "2026-07-01T14:30:00Z",
  "file_count": 142,
  "total_bytes": 12884901888,
  "by_category": [
    {
      "category": "images",
      "file_count": 140,
      "bytes": 12800000000,
      "proportion": 0.993
    },
    {
      "category": "other",
      "file_count": 2,
      "bytes": 84901888,
      "proportion": 0.007
    }
  ],
  "by_mime": [
    {
      "mime_type": "image/jpeg",
      "file_count": 120,
      "bytes": 11000000000,
      "proportion": 0.854
    },
    {
      "mime_type": "image/png",
      "file_count": 20,
      "bytes": 1800000000,
      "proportion": 0.140
    }
  ]
}
```

**Rules:**

- `total_bytes` must equal sum of `by_category[].bytes` (within rounding).
- `proportion` = `bytes / total_bytes`; `0` when `total_bytes` is 0.
- Empty volume: `file_count: 0`, empty arrays, `total_bytes: 0`.
- Stats computation scans all `*.json` in `cache/metadata/` except `stats.json` itself.

### MIME category mapping

| Category | MIME patterns |
|---|---|
| `images` | `image/*` |
| `videos` | `video/*` |
| `audio` | `audio/*` |
| `documents` | `application/pdf`, `application/msword`, `application/vnd.*`, `text/*`, `application/rtf` |
| `archives` | `application/zip`, `application/x-tar`, `application/gzip`, `application/x-7z-compressed`, `application/x-rar-compressed` |
| `other` | everything else |

Mapping lives in `internal/monitoring/mime_groups.go` — single function `CategoryForMIME(mime string) string`.

### Extended `SearchQuery` (indexer)

```go
type SearchQuery struct {
    Term           string     // filename/path text search; empty = filter-only
    MimePrefix     string     // e.g. "image/" — prefix match on mime_type
    MinSizeBytes   *int64
    MaxSizeBytes   *int64
    ModifiedAfter  *time.Time
    ModifiedBefore *time.Time
    Limit          int        // default 50, max 200
}
```

Bleve implementation builds a conjunction query: optional `MatchQuery` on `name` + optional numeric/date range filters.

---

## API (Phase 1.2)

Base path: `/api`. All endpoints require JWT.

**Error shape** (unchanged):

```json
{ "error": "human-readable message", "code": "STATS_NOT_FOUND" }
```

### Monitoring

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/monitoring/overview` | JWT | Unified multi-disk overview: disks, volumes, usage, cached stats summary |
| `GET` | `/monitoring/volumes/:id/stats` | JWT | Full stats for one volume (reads cache or computes) |
| `POST` | `/monitoring/volumes/:id/stats/refresh` | JWT | Force recompute + write `stats.json` |

**Authorization:** same as volumes — owner or admin.

#### `GET /monitoring/overview`

**Response:**

```json
{
  "disks": [
    {
      "path": "/data/disks/ssd",
      "name": "ssd",
      "label": "SSD",
      "total_bytes": 1000000000000,
      "free_bytes": 800000000000,
      "used_by_volumes_bytes": 12884901888,
      "volume_count": 1
    },
    {
      "path": "/data/disks/hdd1",
      "name": "hdd1",
      "label": "HDD 1",
      "total_bytes": 4000000000000,
      "free_bytes": 3200000000000,
      "used_by_volumes_bytes": 5368709120,
      "volume_count": 1
    }
  ],
  "volumes": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Photos",
      "disk_path": "/data/disks/ssd",
      "quota_bytes": 53687091200,
      "used_bytes": 12884901888,
      "file_count": 142,
      "top_category": "images",
      "stats_computed_at": "2026-07-01T14:30:00Z"
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "Documents",
      "disk_path": "/data/disks/hdd1",
      "quota_bytes": 107374182400,
      "used_bytes": 5368709120,
      "file_count": 48,
      "top_category": "documents",
      "stats_computed_at": "2026-07-01T12:00:00Z"
    }
  ]
}
```

**Computation:**

- `used_by_volumes_bytes` = sum of `used_bytes` for volumes on that disk (from PostgreSQL).
- `volume_count` = count of volumes on that disk.
- Volume summary fields (`file_count`, `top_category`, `stats_computed_at`) come from cached `stats.json` if present; otherwise computed lazily on this request.

#### `GET /monitoring/volumes/:id/stats`

**Response:**

```json
{
  "volume_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Photos",
  "quota_bytes": 53687091200,
  "used_bytes": 12884901888,
  "free_bytes": 40802189312,
  "usage_percent": 0.24,
  "file_count": 142,
  "computed_at": "2026-07-01T14:30:00Z",
  "by_category": [ "..." ],
  "by_mime": [ "..." ]
}
```

- `free_bytes` = `quota_bytes - used_bytes` when quota > 0; `null` when quota is 0 (unlimited).
- `usage_percent` = `used_bytes / quota_bytes` when quota > 0; `null` when unlimited.

#### `POST /monitoring/volumes/:id/stats/refresh`

Recomputes stats from metadata cache, writes `stats.json`, returns same body as GET stats.

**Future (Phase 1.4):** the `compute_stats` macro calls the same `monitoring.Service.ComputeStats(ctx, volumeID)` function — no duplicate logic.

### Search

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/volumes/:id/search` | JWT | Bleve search within volume |

**Query parameters:**

| Param | Type | Description |
|---|---|---|
| `q` | string | Filename/path search term (optional) |
| `mime_prefix` | string | MIME prefix filter, e.g. `image/` |
| `min_size` | int64 | Minimum file size in bytes |
| `max_size` | int64 | Maximum file size in bytes |
| `modified_after` | RFC3339 | Modified on or after |
| `modified_before` | RFC3339 | Modified on or before |
| `limit` | int | Max results (default 50, max 200) |

**Response:**

```json
{
  "query": {
    "q": "vacation",
    "mime_prefix": "image/",
    "limit": 50
  },
  "total": 3,
  "results": [
    {
      "name": "vacation.jpg",
      "relative_path": "2024/vacation.jpg",
      "mime_type": "image/jpeg",
      "size_bytes": 2048576,
      "modified_at": "2026-06-15T10:00:00Z",
      "has_thumbnail": true
    }
  ]
}
```

**Validation errors (422):**

| Code | Condition |
|---|---|
| `VOLUME_NOT_FOUND` | Unknown volume ID |
| `FORBIDDEN` | Volume not owned by caller (non-admin) |
| `INVALID_SEARCH_QUERY` | Malformed date or negative size |
| `INDEX_UNAVAILABLE` | Bleve index cannot be opened |

---

## Stats lifecycle

```
Upload/Delete file
    → invalidate stats.json (delete file)
    → used_bytes updated (existing Phase 1.1 flow)

GET /monitoring/volumes/:id/stats
    → if stats.json exists → return cached
    → else → ComputeStats() → write stats.json → return

POST .../stats/refresh
    → ComputeStats() → write stats.json → return

Phase 1.4 compute_stats macro
    → calls ComputeStats() (same function)
```

**ComputeStats algorithm:**

1. List all `FileMetadataRecord` from `MetadataCache.ListAll(rootPath)`.
2. Aggregate by MIME type and by category.
3. Write `stats.json` atomically (temp + rename).
4. Return structured stats.

---

## Frontend (Phase 1.2)

Follow [design/DESIGN.md](./design/DESIGN.md): dark canvas, yellow stat numbers, card surfaces, progress bars for usage.

### Routes

| Path | Access | Content |
|---|---|---|
| `/monitoring` | Protected | Monitoring dashboard (replaces placeholder) |

Remove "Soon" badge from Monitoring sidebar item when Phase 1.2 ships.

### `/monitoring` — Dashboard layout

**Header row:**

- Page title: **Monitoring**
- **Settings page** — admin toggle "Hide disk names"; persisted via `PATCH /api/admin/settings`

**Section 1 — Disk overview (top cards):**

- One card per physical disk from overview API.
- **Mask off:** show label (e.g. "SSD"), total/free bytes, bar of `used_by_volumes_bytes / total_bytes`, volume count.
- **Mask on:** show generic title "Storage" + index ("Storage 1", "Storage 2"), hide path/name/label; show only free/total bytes and volume count.

**Section 2 — Unified volume list:**

- All volumes across disks in a table or card grid.
- Columns: name, usage bar (used/quota), file count, top category, disk (hidden when mask on).
- Click volume row → expands or navigates to stats detail panel.

**Section 3 — Volume detail panel (inline or drawer):**

- Selected volume: large yellow stat numbers for used/quota/percent.
- MIME category breakdown — horizontal bars or donut chart (`MimeBreakdownChart`).
- Top MIME types list (top 5 by bytes).
- **Refresh stats** button → POST refresh endpoint.

**Section 4 — Search (`VolumeSearchPanel`):**

- Volume selector dropdown (user's volumes).
- Text input for filename search.
- Optional filter chips: type (images/documents/…), size range, date range.
- Results table: name, path, size, modified, thumbnail if image.
- Empty state when no matches.

### Disk masking behavior (server-enforced)

| Element | Mask off | Mask on (non-admin) |
|---|---|---|
| Disk card title | Label ("SSD") | "Storage 1" (ordered by API response index) |
| Disk path / name | Shown where relevant | Hidden (API returns empty path, generic label) |
| Volume list disk column | Label | Hidden |
| `VolumeCard` on `/volumes` | Disk path shown | Disk path hidden |
| Admin user | Real paths in volume creation only | Masked everywhere else |

Configured in **Settings** (`PATCH /api/admin/settings`). Toggle removed from Monitoring page header.

### Settings page (minimal Phase 1.2 scope)

- `/settings` replaces Phase 0 placeholder for this security option only.
- Admin: toggle "Hide disk names" with security explanation.
- Non-admin: read-only status ("Disk names are hidden/visible").

### Data fetching

- `useMonitoringOverview()` — TanStack Query, key `['monitoring', 'overview']`, staleTime 30s.
- `useVolumeStats(volumeId)` — enabled when volume selected.
- `useVolumeSearch(volumeId, params)` — debounced 300ms on `q` input.
- Invalidate overview query after stats refresh.

---

## Code Style

### Go — monitoring service pattern

Handlers stay thin; stats logic in `internal/monitoring/`.

```go
// internal/monitoring/service.go
func (s *Service) GetVolumeStats(ctx context.Context, vol *volume.Volume) (*VolumeStats, error) {
    if cached, err := s.statsCache.Read(vol.RootPath); err == nil {
        return cached, nil
    }
    return s.ComputeStats(ctx, vol)
}

func (s *Service) ComputeStats(ctx context.Context, vol *volume.Volume) (*VolumeStats, error) {
    records, err := s.metadata.ListAll(vol.RootPath)
    if err != nil {
        return nil, err
    }
    stats := aggregateRecords(records)
    if err := s.statsCache.Write(vol.RootPath, stats); err != nil {
        return nil, err
    }
    return stats, nil
}
```

- Bleve accessed only through `VolumeIndexer` interface — extend `SearchQuery`, implement filters in `bleve.go`.
- Stats cache writes: temp file + rename (same pattern as `.volume.json`).
- `ComputeStats` is exported/public on the service — Phase 1.4 task module imports it.

### TypeScript

- Presentational components in `components/monitoring/`.
- Human-readable bytes via shared formatter (reuse from volumes UI).
- Percentages: one decimal place; yellow color for stat-display numbers per design system.

---

## Testing Strategy

See [docs/project.md — Coverage targets](./docs/project.md#coverage-targets).

| Layer | Focus | Location |
|---|---|---|
| Backend unit | MIME category mapping, stats aggregation math, stats.json read/write | `internal/monitoring/*_test.go` |
| Backend unit | Bleve Search with term + mime_prefix + size + date filters | `internal/indexer/bleve_test.go` |
| Backend unit | MetadataCache.ListAll | `internal/volume/metadata_cache_test.go` |
| Backend integration | Overview endpoint with temp disk + volumes | `internal/api/monitoring_handler_test.go` |
| Backend integration | Search endpoint | `internal/api/search_handler_test.go` |
| Frontend unit | DiskMaskToggle, MimeBreakdownChart render, search debounce | `web/src/components/monitoring/*.test.tsx` |

**Coverage target:** 80% minimum on `internal/monitoring/` and search extensions in `internal/indexer/`.

**Verify before ship:**

```bash
go test ./internal/monitoring/... ./internal/indexer/... -cover
go test ./...
cd web && npm run test
# Manual UC2.1 – UC2.4 via UI + curl script above
```

---

## Boundaries

### Always

- Route all search through `VolumeIndexer` interface — never import Bleve outside `internal/indexer/`.
- Stats computation reads from `./cache/metadata/` — never walk `userdata/` for breakdown (metadata cache is the scan target).
- Invalidate `stats.json` on file upload and delete (hook in existing `FileService`).
- Scope monitoring data to volume owner; admin bypass explicit in service layer.
- Follow [design/DESIGN.md](./design/DESIGN.md) for monitoring UI.
- Export `ComputeStats` as the single stats entry point for Phase 1.4 reuse.

### Ask first

- Adding chart libraries not already in the frontend stack.
- Per-user display preferences beyond instance-wide admin settings.
- Cross-volume search (architectural scope change).
- Changing `stats.json` location or schema after Phase 1.2 ships (requires ADR).

### Never

- Walk `userdata/` recursively for stats when metadata cache exists.
- Import Bleve outside `internal/indexer/`.
- Duplicate stats computation logic for Phase 1.4 macro — one function only.
- Expose disk paths in masked UI mode (client must strip even if API returns them).
- Block Phase 1.2 on Phase 1.4 task scheduler.

---

## Success Criteria

Phase 1.2 is **done** when all of the following pass:

- [ ] **SC2.1** `GET /api/monitoring/overview` returns all disks with total/free/used_by_volumes and all user volumes with summary stats
- [ ] **SC2.2** Admin overview includes all users' volumes; regular user sees only own volumes
- [ ] **SC2.3** `GET /api/monitoring/volumes/:id/stats` returns usage, file count, category breakdown, MIME breakdown
- [ ] **SC2.4** Stats match actual files after upload (categories sum to total; proportions correct)
- [ ] **SC2.5** `stats.json` written to `./cache/metadata/` after compute; deleted on file upload/delete
- [ ] **SC2.6** `POST .../stats/refresh` forces recompute and updates cache
- [ ] **SC2.7** Second disk with second volume — both appear in overview unified list
- [ ] **SC2.8** `GET /api/volumes/:id/search?q=vacation` returns matching filenames
- [ ] **SC2.9** Search filters work: `mime_prefix`, `min_size`, `max_size`, `modified_after`, `modified_before`
- [ ] **SC2.10** Empty search term with filters only returns filter-matched files
- [ ] **SC2.11** UI: monitoring dashboard shows real data from live volume (UC2.1)
- [ ] **SC2.12** UI: multi-disk unified volume list (UC2.2)
- [ ] **SC2.13** Admin can enable disk name masking in Settings; non-admins see masked disk identifiers app-wide (UC2.3)
- [ ] **SC2.14** UI: volume search finds "vacation" matches (UC2.4)
- [ ] **SC2.15** Monitoring sidebar item active without "Soon" badge
- [ ] **SC2.16** `ComputeStats` callable from monitoring service (documented for Phase 1.4 integration)
- [ ] **SC2.17** `go test ./...` and `cd web && npm run test` pass

---

## Implementation order (preview for `/plan`)

Suggested vertical slices — `/plan` will expand into `tasks/plan.md`:

1. **MetadataCache.ListAll** — scan all per-file JSON records
2. **MIME groups + stats aggregation** — `mime_groups.go`, aggregate logic, tests
3. **Stats cache** — `stats.json` read/write/invalidate
4. **Monitoring service** — `ComputeStats`, `GetOverview`, `GetVolumeStats`
5. **Stats invalidation hook** — wire into `FileService` upload/delete
6. **Extend SearchQuery + Bleve filters** — indexer changes + tests
7. **Monitoring API handlers** — overview, stats, refresh
8. **Search API handler** — volume search endpoint
9. **Settings page + API** — `internal/settings/`, admin toggle for disk masking
10. **Frontend monitoring page** — disk cards, unified list, stats panel
11. **Frontend search panel** — volume selector, filters, results
12. **Integration** — Docker end-to-end UC2.1–UC2.4

---

## Decisions log

| # | Decision | Status |
|---|---|---|
| D1 | Stats cached in `cache/metadata/stats.json` per volume | ✅ Proposed |
| D2 | MIME categories align with `sort_by_type` macro folders | ✅ Proposed |
| D3 | Disk masking = admin setting in Settings, server-side enforcement for all display APIs | ✅ Shipped |
| D4 | Search UI on monitoring page with volume selector | ✅ Proposed |
| D5 | `ComputeStats` shared with Phase 1.4 `compute_stats` macro | ✅ Proposed |
| D6 | Lazy stats compute on read if cache missing | ✅ Proposed |
| D7 | Search endpoint at `/volumes/:id/search` (not under `/monitoring`) | ✅ Proposed — keeps search co-located with volume resource |

---

## Open Questions

| # | Question | Status |
|---|---|---|
| OQ1 | **Chart library for MIME breakdown** — see below | ⏳ Pending |
| OQ2 | **Mask toggle scope** — app-wide via API (resolved) | ✅ Resolved |
| OQ3 | **Search result actions** — see below | ⏳ Pending |

### OQ1 — Chart library (plain language)

The monitoring dashboard needs a visual breakdown of file types (e.g. "80% images, 20% documents"). Two options:

| Option | What it means | Trade-off |
|---|---|---|
| **A — CSS progress bars** | Horizontal bars built with Tailwind, no new dependency | Simpler; less polished for many categories |
| **B — Recharts (or similar)** | Donut/bar chart library | Richer visuals; one new frontend dependency to approve |

**Recommendation:** Option A for Phase 1.2 — horizontal category bars match the design system and avoid a new dependency. Upgrade to charts later if needed.

### OQ2 — Disk mask scope (plain language)

When the user hides disk names, should that apply only on the monitoring page, or everywhere (including the volume list at `/volumes`)?

| Option | Meaning |
|---|---|
| **A — Monitoring page only** | Simpler; `/volumes` still shows disk labels for context when creating/managing |
| **B — Global preference** | Consistent everywhere; slightly more UI work |

**Recommendation:** Option A — masking is a monitoring-dashboard privacy feature, not a global UI mode.

### OQ3 — Search result actions (plain language)

When search returns a file, what can the user do from the result?

| Option | Meaning |
|---|---|
| **A — View only** | Show name, size, date, thumbnail; click navigates to `/volumes/:id` at file path |
| **B — Inline download** | Download button directly in search results |

**Recommendation:** Option A — click row opens the volume file browser at the file's directory. Keeps search panel simple.

---

## Next step

After spec approval → `/plan` to produce `tasks/plan.md` with ordered, verifiable tasks.
