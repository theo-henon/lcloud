# Spec: Phase 1.3 — Plugin system

> **Status:** Draft — pending review
> **Scope:** Event bus, go-plugin runtime, plugin registry, plugin manager UI, example `file-type-validator` plugin, SDK stub.
> **Prerequisite:** Phase 1.1 complete (volumes, files, filters, layout with `plugins/` dir). Phase 1.2 complete (monitoring, search) recommended but not blocking for plugin wiring.
> **Sources:** [STARTUP.md](./STARTUP.md#phase-13--plugin-system), [VISION.md](./VISION.md), [docs/project.md](./docs/project.md)

---

## Assumptions (correct me now or I proceed)

1. **Phases 1.1–1.2 are shipped** — volume CRUD, file upload/delete, `.volume.json` filters, volume layout includes `./plugins/` and `./logs/`.
2. **`internal/plugin/` is greenfield** — module referenced in docs but not yet created; Phase 1.3 creates it entirely.
3. **Plugins are instance-wide binaries** — admin drops executables into a host `plugins/` directory; discovery happens at startup (UC3.1 requires restart, no hot-reload in MVP).
4. **Core upload filter stays synchronous** — Phase 1.1 `ValidateExtension` still rejects bad files before save; plugins receive `file.uploaded` only after a successful upload. The example plugin is a **post-upload audit** (demonstrates the bus), not a replacement for core validation.
5. **`file.moved` / `file.renamed` are defined but not emitted yet** — no move/rename API exists in Phase 1.1; event types are registered on the bus, emission deferred until a move operation ships (Phase 1.4 `move_files` macro or a dedicated file API).
6. **`task.executed` is defined but not emitted yet** — task scheduler is Phase 1.4; bus accepts the event type, emission wired in 1.4.
7. **Plugin logs are stored in PostgreSQL for UI** — recent activity feed in the plugin manager; optional mirror to `{volume}/logs/` is post-MVP. Keeps `./logs/` directory structure intact without new on-disk schema in 1.3.
8. **Per-volume plugin data dirs are created lazily** — `{volume.root}/plugins/{plugin-id}/` is created the first time a plugin handles an event for that volume (no separate "install per volume" UI in MVP).
9. **Enable/disable is admin-only** — toggling a plugin starts/stops its subprocess and controls event delivery.
10. **One subprocess per plugin binary** — go-plugin spawns an isolated child process; crash marks plugin `error` without taking down lcloud core.
11. **Plugin API is read-mostly for MVP** — plugins may read volume config (filters, name), emit custom events, write to their volume data dir, and append log entries. No file delete/move via plugin API in 1.3.
12. **ADR required before merge** — plugin interface contract and event bus design documented in `docs/adr/` via `documentation-and-adrs`.

---

## Objective

Build the foundational plugin infrastructure so users can extend lcloud with external binaries that react to volume events — the extensibility layer described in VISION.md.

**Who:** Self-hosters and developers who want to customize lcloud behavior without forking the core.

**User stories:**

| ID | Story |
|---|---|
| UC3.1 | As an admin, I drop the `file-type-validator` binary into `plugins/`, restart lcloud — the plugin appears in the plugin manager as **running**. |
| UC3.2 | As a user, I upload a file to a filtered volume — the plugin receives `file.uploaded`, validates it, and a notice appears in the plugin activity feed in the UI. |

**Out of scope for Phase 1.3:**

- Task scheduler and `task.executed` emission (Phase 1.4)
- File move/rename API and `file.moved` / `file.renamed` emission
- Hot-reload of plugins without restart
- Plugin marketplace / remote registry
- WASM runtime (IDEAS.md — post-MVP)
- Plugin file mutations (delete uploaded file on validation failure)
- Per-volume plugin install wizard (lazy data dir only)
- Community plugin signing / verification
- Changes to `.volume.json` schema or volume directory structure

---

## Tech Stack

See [docs/project.md — Tech Stack](./docs/project.md#tech-stack).

Phase 1.3 adds:

| Layer | Addition |
|---|---|
| Backend | `github.com/hashicorp/go-plugin` — subprocess RPC (GRPC handshake) |
| Backend | PostgreSQL models: `Plugin`, `PluginLogEntry` |
| Example | `examples/file-type-validator/` — reference Go plugin source + manifest |
| SDK | `pkg/pluginsdk/` — minimal Go helpers for plugin authors |
| Docs | `docs/plugins/README.md` — how to write a plugin |

---

## Commands

### Development

```bash
# Add go-plugin dependency (during /build)
go get github.com/hashicorp/go-plugin

# Build example plugin
go build -o plugins/file-type-validator ./examples/file-type-validator

# Backend tests (plugin module focus)
go test ./internal/plugin/... -cover
go test ./... -cover

# Frontend tests
cd web && npm run test

# Local stack (mount plugins dir — see Docker changes below)
docker compose up -d --build
docker compose logs -f app
```

### Verification (Phase 1.3 done)

```bash
# 1. Build and place example plugin
go build -o plugins/file-type-validator ./examples/file-type-validator
cp examples/file-type-validator/file-type-validator.json plugins/file-type-validator.json  # manifest sidecar

# 2. Restart stack
docker compose up -d --build

TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"<ADMIN_EMAIL>","password":"<ADMIN_PASSWORD>"}' \
  | jq -r '.access_token')

# 3. List plugins — file-type-validator running
curl -s http://localhost:8080/api/plugins \
  -H "Authorization: Bearer $TOKEN" | jq

# 4. Upload allowed file → plugin activity log entry
VOL_ID="<volume-uuid>"
curl -s -X POST "http://localhost:8080/api/volumes/$VOL_ID/files" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@photo.png" | jq

curl -s "http://localhost:8080/api/plugins/logs?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq

# 5. Disable plugin — status stopped, no new log entries on upload
curl -s -X PATCH "http://localhost:8080/api/plugins/file-type-validator" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}' | jq
```

---

## Project Structure

See [docs/project.md — Project structure](./docs/project.md#project-structure).

Phase 1.3 creates or extends:

```
lcloud/
├── internal/
│   └── plugin/
│       ├── eventbus.go           ← in-process pub/sub, typed events
│       ├── events.go             ← event types + payloads
│       ├── publisher.go          ← EventPublisher interface (injected into volume services)
│       ├── runtime.go            ← go-plugin client, subprocess lifecycle
│       ├── registry.go           ← scan PLUGINS_PATH, load manifests, validate binaries
│       ├── host_api.go           ← HostAPI served to plugins (Emit, Log, GetVolumeConfig)
│       ├── service.go            ← orchestration: registry + runtime + bus + logs
│       ├── model.go              ← GORM Plugin, PluginLogEntry
│       ├── rpc.go                ← go-plugin GRPC service definitions
│       └── service_test.go
├── pkg/
│   └── pluginsdk/
│       ├── plugin.go             ← plugin-side interface + Serve helper
│       └── events.go             ← shared event type constants
├── plugins/
│   ├── .gitkeep
│   └── file-type-validator/
│       ├── main.go               ← example plugin entrypoint
│       ├── validator.go          ← file.uploaded handler
│       └── file-type-validator.json  ← manifest sidecar (copied next to binary)
├── docs/
│   ├── adr/
│   │   └── NNN-plugin-system.md  ← interface + bus decision (created at /build)
│   └── plugins/
│       └── README.md             ← plugin author guide
├── internal/api/
│   ├── plugin_handler.go
│   └── router.go                 ← register plugin routes
├── internal/volume/
│   ├── file_service.go           ← inject EventPublisher; emit after upload/delete
│   └── service.go                ← emit volume.created / volume.deleted / volume.updated
├── cmd/server/main.go            ← wire plugin service, start registry
├── web/src/
│   ├── pages/PluginsPage.tsx     ← replace PlaceholderPage
│   ├── components/plugins/
│   │   ├── PluginList.tsx
│   │   ├── PluginStatusBadge.tsx
│   │   └── PluginActivityFeed.tsx
│   └── hooks/usePlugins.ts
├── docker-compose.yml              ← mount plugins volume
├── .env.example                    ← PLUGINS_PATH
└── Dockerfile                      ← optional: build example plugin in CI image
```

---

## Architecture

### Event flow

```
FileService.Upload (success)
    → EventPublisher.Publish(file.uploaded)
    → EventBus (in-process)
    → Plugin Runtime (for each enabled plugin subscribed)
    → go-plugin RPC → plugin subprocess HandleEvent()
    → plugin may call HostAPI.Log / HostAPI.Emit
    → PluginLogEntry persisted → UI activity feed
```

### Module boundaries (AGENTS.md)

- Domain modules (`volume`, future `task`) **only** call `EventPublisher.Publish` — never import go-plugin or talk to plugin subprocesses directly.
- All plugin lifecycle, RPC, and subscription logic lives in `internal/plugin/`.
- Plugins access volume data **only** through `HostAPI` — no DB, no direct disk writes outside their volume data dir.

### go-plugin handshake

| Side | Responsibility |
|---|---|
| **Host (lcloud)** | Scans `PLUGINS_PATH`, spawns plugin subprocess, implements `HostAPI` GRPC service, dispatches events |
| **Plugin binary** | Implements `LcloudPlugin` GRPC service, declares subscriptions in manifest, calls `HostAPI` for callbacks |

Communication: HashiCorp go-plugin with **GRPC** plugin protocol (default in go-plugin v1.x).

---

## Data models

### Plugin manifest (sidecar `{binary-name}.json` next to binary)

Required for discovery. Binary without manifest is skipped with a warning log.

```json
{
  "id": "file-type-validator",
  "name": "File Type Validator",
  "version": "1.0.0",
  "description": "Validates uploaded files against volume extension filters",
  "author": "lcloud",
  "subscribe": ["file.uploaded"],
  "emit": ["validation.failed", "validation.passed"]
}
```

**Rules:**

- `id` — stable slug, `[a-z0-9-]+`, unique per instance; matches volume data dir name `./plugins/{id}/`.
- `subscribe` — event names the plugin wants; runtime filters before RPC dispatch.
- `emit` — declared custom events (documentation; bus accepts any `plugin.*` or namespaced custom events).

### PostgreSQL — `plugins`

```go
type Plugin struct {
    ID            string    `gorm:"primaryKey"` // manifest id
    Name          string    `gorm:"not null"`
    Version       string    `gorm:"not null"`
    Description   string
    BinaryPath    string    `gorm:"not null"`
    ManifestPath  string    `gorm:"not null"`
    Enabled       bool      `gorm:"not null;default:true"`
    Status        string    `gorm:"not null;default:stopped"` // running | stopped | error
    Subscriptions StringArray `gorm:"type:jsonb;not null"` // from manifest
    LastError     string
    DiscoveredAt  time.Time
    UpdatedAt     time.Time
}
```

On startup: scan directory → upsert rows → start enabled plugins.

### PostgreSQL — `plugin_log_entries`

```go
type PluginLogEntry struct {
    ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
    PluginID  string     `gorm:"not null;index"`
    VolumeID  *uuid.UUID `gorm:"type:uuid;index"`
    EventType string     // e.g. file.uploaded, validation.passed
    Level     string     `gorm:"not null"` // info | warn | error
    Message   string     `gorm:"not null"`
    Payload   JSON       `gorm:"type:jsonb"` // optional structured detail
    CreatedAt time.Time  `gorm:"index"`
}
```

Retention: **500 entries max per plugin** (decision OQ1) — delete oldest on insert when count exceeds 500.

---

## Event catalog

### Core events (emitted by lcloud in Phase 1.3)

| Event | Emitted when | Payload highlights |
|---|---|---|
| `file.uploaded` | After successful `FileService.Upload` | `volume_id`, `name`, `relative_path`, `mime_type`, `size_bytes`, `filters` |
| `file.deleted` | After successful `FileService.Delete` | `volume_id`, `relative_path`, `size_bytes` |
| `volume.created` | After `VolumeService.Create` | `volume_id`, `name`, `owner_id`, `filters`, `quota_bytes` |
| `volume.deleted` | After `VolumeService.Delete` | `volume_id`, `name` |
| `volume.updated` | After `VolumeService.Patch` | `volume_id`, changed fields |

### Defined but not emitted until later phases

| Event | Phase |
|---|---|
| `file.moved` | When move API / `move_files` macro ships |
| `file.renamed` | When rename API ships |
| `task.executed` | Phase 1.4 |
| `task.failed` | Phase 1.4 |
| `volume.alert.usage` | Phase 1.4 (`alert_usage` macro) |

### Plugin lifecycle events (emitted by plugin runtime)

| Event | When |
|---|---|
| `plugin.registered` | Plugin subprocess started successfully |
| `plugin.unregistered` | Plugin stopped or crashed |

### Custom events (example plugin)

| Event | When |
|---|---|
| `validation.passed` | Extension matches volume filters |
| `validation.failed` | Extension mismatch (edge case: filter changed after upload, manual file copy) |

Custom events are re-published on the bus so other plugins can subscribe in future.

### Event envelope

```go
type Event struct {
    ID        string          `json:"id"`         // uuid
    Type      string          `json:"type"`
    Timestamp time.Time       `json:"timestamp"`
    VolumeID  string          `json:"volume_id,omitempty"`
    Payload   json.RawMessage `json:"payload"`
}
```

---

## Plugin RPC interfaces

### `LcloudPlugin` (implemented by plugin binary)

```go
type LcloudPlugin interface {
    // Called by host when a subscribed event occurs.
    HandleEvent(ctx context.Context, event *Event) (*HandleResult, error)
}

type HandleResult struct {
    // Optional custom events to emit onto the bus.
    Emit []Event `json:"emit,omitempty"`
}
```

### `HostAPI` (implemented by lcloud host, called by plugin)

```go
type HostAPI interface {
    // Re-emit an event onto the bus (for custom plugin events).
    Emit(ctx context.Context, event *Event) error

    // Append a log entry visible in the plugin manager UI.
    Log(ctx context.Context, entry LogRequest) error

    // Read volume config (filters, name, quota) — no file content access.
    GetVolumeConfig(ctx context.Context, volumeID string) (*VolumeConfigView, error)

    // Resolve path for plugin's isolated data dir; creates dir if missing.
    PluginDataDir(ctx context.Context, volumeID, pluginID string) (string, error)
}
```

**Security rules for HostAPI:**

- `GetVolumeConfig` — owner-scoped volumes only; returns filters/name/quota, not disk paths (masking-friendly).
- `PluginDataDir` — only `{volume.root}/plugins/{plugin-id}/`; created with `0755`.
- No raw `userdata/` file read/write in MVP.

---

## Example plugin — `file-type-validator`

**Purpose:** Demonstrate end-to-end plugin flow (UC3.1, UC3.2).

**Behavior on `file.uploaded`:**

1. Read `filters` from event payload (also available via `GetVolumeConfig`).
2. Run same logic as `volume.ValidateExtension(filters, filename)`.
3. If valid → `HostAPI.Log(info, "validation passed for {filename}")` + emit `validation.passed`.
4. If invalid → `HostAPI.Log(warn, "validation failed for {filename}")` + emit `validation.failed`.
5. Does **not** delete the file (audit-only in MVP; deletion would require a future HostAPI).

**Note for UC3.2 demo:** Upload a `.png` to a volume that allows `.png` — plugin logs success. To demo `validation.failed`, change volume filters after upload or copy a disallowed file directly into `userdata/` (manual test only).

---

## API (Phase 1.3)

Base path: `/api`. All endpoints require JWT unless noted.

### Plugins

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/plugins` | JWT | List discovered plugins with status |
| `GET` | `/plugins/:id` | JWT | Single plugin detail |
| `PATCH` | `/plugins/:id` | Admin | Enable/disable plugin |
| `GET` | `/plugins/logs` | JWT | Recent activity feed |

#### `GET /plugins`

**Response:**

```json
{
  "plugins": [
    {
      "id": "file-type-validator",
      "name": "File Type Validator",
      "version": "1.0.0",
      "description": "Validates uploaded files against volume extension filters",
      "enabled": true,
      "status": "running",
      "subscriptions": ["file.uploaded"],
      "last_error": "",
      "discovered_at": "2026-07-01T10:00:00Z"
    }
  ]
}
```

#### `PATCH /plugins/:id`

**Request:**

```json
{ "enabled": false }
```

**Behavior:**

- `enabled: true` → start subprocess if not running; set status `running` or `error`.
- `enabled: false` → graceful stop subprocess; set status `stopped`.

#### `GET /plugins/logs`

**Query parameters:**

| Param | Type | Description |
|---|---|---|
| `plugin_id` | string | Filter by plugin (optional) |
| `volume_id` | uuid | Filter by volume (optional) |
| `limit` | int | Max entries (default 50, max 200) |

**Response:**

```json
{
  "entries": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "plugin_id": "file-type-validator",
      "volume_id": "660e8400-e29b-41d4-a716-446655440001",
      "event_type": "validation.passed",
      "level": "info",
      "message": "validation passed for photo.png",
      "created_at": "2026-07-01T14:30:00Z"
    }
  ]
}
```

**Authorization:**

- Regular users see log entries for volumes they own.
- Admin sees all entries.

**Error codes:**

| Code | Condition |
|---|---|
| `PLUGIN_NOT_FOUND` | Unknown plugin id |
| `PLUGIN_START_FAILED` | Enable true but subprocess failed to start |
| `FORBIDDEN` | Non-admin PATCH |

---

## Infrastructure changes

### Environment

Add to `.env.example`:

```bash
# Plugin binaries directory (inside container)
PLUGINS_PATH=/plugins
```

Default in `config.Load()`: `./plugins` for local dev, `/plugins` when `PLUGINS_PATH` set.

### Docker Compose

```yaml
app:
  environment:
    PLUGINS_PATH: /plugins
  volumes:
    - ${PLUGINS_PATH:-./plugins}:/plugins:ro
```

Mount read-only: lcloud executes but does not modify plugin binaries.

### Startup sequence (`cmd/server/main.go`)

1. Migrate `Plugin`, `PluginLogEntry` models.
2. Construct `plugin.Service` with config, volume service, event bus.
3. `registry.ScanAndLoad()` — discover binaries + manifests.
4. `runtime.StartEnabled()` — spawn go-plugin clients.
5. Inject `EventPublisher` into `VolumeService` and `FileService`.
6. On shutdown (SIGTERM): `runtime.StopAll()`.

---

## Frontend (Phase 1.3)

Follow [design/DESIGN.md](./design/DESIGN.md): dark cards, status badges (green `running`, muted `stopped`, red `error`), yellow accents sparingly.

### Routes

| Path | Access | Content |
|---|---|---|
| `/plugins` | Protected | Plugin manager (replaces placeholder) |

Remove "Soon" badge from Plugins sidebar item when Phase 1.3 ships.

### `/plugins` — Layout

**Section 1 — Plugin list (`PluginList`):**

- Table/cards: name, version, status badge, subscribed events, enabled toggle (admin only).
- Empty state: "No plugins found — drop a binary + manifest into the plugins directory and restart."

**Section 2 — Activity feed (`PluginActivityFeed`):**

- Recent log entries from `GET /plugins/logs`.
- Columns: time, plugin name, volume name (linked), level, message.
- Auto-refresh every 10s via TanStack Query `refetchInterval`.

### Data fetching

- `usePlugins()` — key `['plugins']`, staleTime 15s.
- `usePluginLogs(params)` — key `['plugins', 'logs', params]`, refetchInterval 10s.
- `useTogglePlugin(id)` — mutation PATCH, invalidate plugins query.

---

## Code Style

### Go — EventPublisher injection (volume stays ignorant of plugins)

```go
// internal/plugin/publisher.go
type EventPublisher interface {
    Publish(ctx context.Context, event Event)
}

// internal/volume/file_service.go — after successful upload
if s.events != nil {
    s.events.Publish(ctx, plugin.NewFileUploadedEvent(vol, record))
}
```

- `Publish` is fire-and-forget (async goroutine per event batch) — upload API latency must not wait on plugin RPC.
- Plugin RPC failures log to `Plugin.LastError` + `plugin_log_entries`; never fail the originating user operation.

### Go — Plugin runtime

```go
func (r *Runtime) Dispatch(ctx context.Context, event Event) {
    for _, p := range r.runningPlugins() {
        if !p.SubscribesTo(event.Type) {
            continue
        }
        go func(plugin *RunningPlugin) {
            if _, err := plugin.Client.HandleEvent(ctx, &event); err != nil {
                r.markError(plugin.ID, err)
            }
        }(p)
    }
}
```

### TypeScript

- Status badge maps: `running` → emerald, `stopped` → muted, `error` → rose (design tokens).
- Admin-only toggle: hide disable control for non-admin users (read-only list).

---

## Testing Strategy

See [docs/project.md — Coverage targets](./docs/project.md#coverage-targets).

| Layer | Focus | Location |
|---|---|---|
| Backend unit | Event bus subscribe/dispatch, manifest parsing | `internal/plugin/eventbus_test.go`, `registry_test.go` |
| Backend unit | EventPublisher async behavior | `internal/plugin/service_test.go` |
| Backend integration | Mock plugin binary (test helper) handles event | `internal/plugin/runtime_test.go` |
| Backend integration | Plugin API handlers enable/disable | `internal/api/plugin_handler_test.go` |
| Backend integration | File upload emits event (mock publisher) | `internal/volume/file_service_test.go` |
| Example plugin | Validator logic | `examples/file-type-validator/validator_test.go` |
| Frontend unit | PluginList, PluginStatusBadge, activity feed | `web/src/components/plugins/*.test.tsx` |

**Coverage target:** 80% minimum on `internal/plugin/`.

**Verify before ship:**

```bash
go build -o plugins/file-type-validator ./examples/file-type-validator
go test ./internal/plugin/... -cover
go test ./...
cd web && npm run test
# Manual UC3.1 – UC3.2 via UI + curl script above
```

---

## Boundaries

### Always

- Domain modules publish events via `EventPublisher` only — never import `internal/plugin/runtime` or go-plugin from `volume/`.
- Plugin subprocess crash must not crash lcloud core.
- User-facing operations (upload, delete, volume CRUD) succeed even if all plugins are down.
- Plugin API surface is minimal and explicit — expand via ADR, not ad-hoc.
- Write ADR in `docs/adr/` before merge documenting interface + bus design.
- Follow [design/DESIGN.md](./design/DESIGN.md) for plugin manager UI.
- Mount `plugins/` read-only in Docker.

### Ask first

- Adding plugin HostAPI methods beyond Log / Emit / GetVolumeConfig / PluginDataDir.
- Hot-reload without restart.
- Plugin ability to mutate userdata files.
- New core event types beyond the catalog above.
- Signing / verification of plugin binaries.

### Never

- Direct DB access from plugin subprocess.
- Direct disk writes outside `{volume}/plugins/{plugin-id}/` from plugins.
- Call plugin RPC from Gin handlers — handlers call `plugin.Service` for management only.
- Block file upload on plugin validation failure (core filter handles rejection; plugin is audit).
- Change `.volume.json` schema for plugin state — plugin instance state lives in PostgreSQL + volume plugin dir.

---

## Success Criteria

Phase 1.3 is **done** when all of the following pass:

- [ ] **SC3.1** `PLUGINS_PATH` configurable; defaults documented in `.env.example`
- [ ] **SC3.2** Docker mounts plugins directory; example plugin binary discoverable after restart
- [ ] **SC3.3** Registry scans binaries + manifest sidecars; invalid entries logged, skipped
- [ ] **SC3.4** `GET /api/plugins` lists discovered plugins with `running` / `stopped` / `error` status
- [ ] **SC3.5** Admin `PATCH /api/plugins/:id` enable/disable starts/stops subprocess
- [ ] **SC3.6** Successful file upload emits `file.uploaded` on the bus (verified via mock + integration test)
- [ ] **SC3.7** File delete emits `file.deleted`; volume create/delete/patch emit corresponding events
- [ ] **SC3.8** Enabled plugin subprocess receives subscribed events via go-plugin RPC
- [ ] **SC3.9** `file-type-validator` logs validation result to `plugin_log_entries`
- [ ] **SC3.10** `GET /api/plugins/logs` returns entries; scoped to volume owner; admin sees all
- [ ] **SC3.11** Custom event `validation.passed` / `validation.failed` re-published on bus
- [ ] **SC3.12** `{volume}/plugins/file-type-validator/` created on first handled event
- [ ] **SC3.13** Plugin crash marks status `error`, sets `last_error`; core remains healthy
- [ ] **SC3.14** UI: plugin manager shows plugin list with status (UC3.1)
- [ ] **SC3.15** UI: activity feed shows notice after upload (UC3.2)
- [ ] **SC3.16** UI: Plugins sidebar active without "Soon" badge
- [ ] **SC3.17** `pkg/pluginsdk/` + `docs/plugins/README.md` exist with minimal working example
- [ ] **SC3.18** ADR documents plugin interface and event bus
- [ ] **SC3.19** `go test ./...` and `cd web && npm run test` pass

---

## Implementation order (preview for `/plan`)

Suggested vertical slices — `/plan` will expand into `tasks/plan.md`:

1. **ADR draft** — plugin interface, event bus, go-plugin GRPC choice
2. **Event types + EventBus** — in-process pub/sub, tests
3. **EventPublisher interface** — inject into volume/file services; emit core events
4. **PostgreSQL models** — Plugin, PluginLogEntry; migrate in main
5. **Plugin manifest parser + registry** — scan PLUGINS_PATH
6. **HostAPI implementation** — Log, Emit, GetVolumeConfig, PluginDataDir
7. **go-plugin runtime** — spawn, dispatch, stop, error handling
8. **Plugin service orchestration** — startup/shutdown wiring in main.go
9. **Plugin API handlers** — list, patch, logs
10. **pluginsdk package** — Serve helper + shared types
11. **file-type-validator example** — build target + manifest
12. **Frontend PluginsPage** — list, toggle, activity feed
13. **Docker + config** — PLUGINS_PATH, volume mount
14. **Integration** — Docker end-to-end UC3.1–UC3.2

---

## Decisions log

| # | Decision | Status |
|---|---|---|
| D1 | go-plugin with GRPC handshake | ✅ Proposed |
| D2 | Manifest sidecar JSON next to binary | ✅ Proposed |
| D3 | Plugin logs in PostgreSQL for UI feed | ✅ Proposed |
| D4 | Async event dispatch — user ops never wait on plugins | ✅ Proposed |
| D5 | Core upload filter unchanged; plugin is post-upload audit | ✅ Proposed |
| D6 | Lazy creation of `{volume}/plugins/{id}/` on first event | ✅ Proposed |
| D7 | Restart required for new binaries (no hot-reload) | ✅ Proposed — matches UC3.1 |
| D8 | `file.moved` / `task.executed` defined, emission deferred | ✅ Proposed |
| D9 | Log retention: 500 entries max per plugin, purge on insert | ✅ Resolved (OQ1) |
| D10 | Example plugin not bundled in Docker image | ✅ Resolved (OQ2) |
| D11 | UC3.2 demo: success-only in UI; failed case documented manually | ✅ Resolved (OQ3) |

---

## Open Questions

| # | Question | Status |
|---|---|---|
| OQ1 | **Plugin log retention** | ✅ Resolved — Option A: 500 entries per plugin, purge on insert |
| OQ2 | **Build example plugin in Docker image** | ✅ Resolved — Option A: not bundled; user builds/copies manually |
| OQ3 | **validation.failed demo path** | ✅ Resolved — Option A: success-only in UI; edge case in docs |

### OQ1 — Plugin log retention (plain language)

**Decision:** Option A — last **500 entries per plugin**. Old entries deleted on insert when limit exceeded. Query always uses `limit`.

### OQ2 — Example plugin in Docker image (plain language)

**Decision:** Option A — **not bundled**. `plugins/` stays git-ignored and user-managed. Build command documented in `docs/plugins/README.md`.

### OQ3 — validation.failed demo (plain language)

**Decision:** Option A — **success-only in UI**. UC3.2 shows "validation passed" after allowed upload. Manual edge-case test (filter change / direct file copy) documented in `docs/plugins/README.md`.

---

## Next step

After spec approval → `/plan` to produce `tasks/plan.md` with ordered, verifiable tasks.
