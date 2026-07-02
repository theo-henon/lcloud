# ADR-001: Plugin system architecture

## Status

Accepted

## Date

2026-07-01

## Context

lcloud's extensibility model (VISION.md) requires external binaries to react to volume lifecycle events without coupling domain modules to go-plugin. Phase 1.3 delivers the MVP plugin infrastructure: event bus, subprocess runtime, registry, manager UI, and one reference plugin.

Requirements:

- Plugins run isolated in subprocesses; crash must not take down the core
- Domain modules (`volume`, future `task`) must not import go-plugin or plugin runtime
- Plugins access volume data only through a narrow host API — no direct DB or arbitrary disk writes
- Discovery via host `plugins/` directory at startup (restart required, no hot-reload in MVP)
- Post-upload audit plugin demonstrates the bus; core upload filter remains synchronous

## Decision

### Runtime: HashiCorp go-plugin with net/rpc protocol

Use `github.com/hashicorp/go-plugin` with the **net/rpc** protocol (not GRPC — avoids a protoc build step in MVP). Each plugin binary is a separate OS process. One subprocess per plugin ID.

### Event bus: in-process pub/sub

An in-process `EventBus` in `internal/plugin/` receives events from domain modules via a bridge implementing `volume.EventPublisher`. Dispatch to plugins is **async** (goroutine); user-facing operations never wait on plugin RPC.

### Discovery: manifest sidecar JSON

Each binary requires a `{name}.json` manifest next to the executable with `id`, `name`, `version`, `subscribe`, `emit`. Binaries without manifests are skipped with a warning.

### Persistence

- Plugin registry state: PostgreSQL `plugins` table (synced from filesystem scan at startup)
- Activity feed: PostgreSQL `plugin_log_entries` (500 entries max per plugin, purge on insert)
- Per-volume plugin data: `{volume.root}/plugins/{plugin-id}/` created lazily by the host on first dispatched event

### RPC interfaces

**LcloudPlugin** (plugin implements): `HandleEvent(ctx, event) → HandleResult`

**HostAPI** (host implements): `Emit`, `Log`, `GetVolumeConfig`, `PluginDataDir`

### MVP plugin ↔ host communication

In Phase 1.3, plugins communicate **outbound** via `HandleResult` only:

- `HandleResult.Logs` → persisted to the activity feed
- `HandleResult.Emit` → re-published on the event bus

The host does **not** expose HostAPI as a reverse RPC channel to the plugin subprocess in MVP. Instead, the host calls `PluginDataDir` internally when dispatching an event with a `volume_id`, ensuring `{volume}/plugins/{plugin-id}/` exists before `HandleEvent` runs.

Bidirectional HostAPI (plugin calls `Log` / `Emit` / `GetVolumeConfig` during `HandleEvent`) is deferred post-MVP; RPC types in `internal/plugin/rpc.go` are prepared for that upgrade.

### Domain coupling

`internal/volume` defines `EventPublisher` interface with typed domain events. `internal/plugin` implements the bridge — volume never imports `internal/plugin`.

### Core events emitted in Phase 1.3

`file.uploaded`, `file.deleted`, `volume.created`, `volume.updated`, `volume.deleted`

Defined but deferred: `file.moved`, `file.renamed`, `task.executed`, `task.failed`, `volume.alert.usage`

## Alternatives considered

### GRPC protocol (go-plugin)

- Pros: typed contracts, multi-language stubs
- Cons: requires protoc/codegen in build pipeline
- Rejected for MVP; net/rpc chosen for simplicity. GRPC migration possible without changing the manifest or event model.

### WASM runtime

- Pros: stronger sandbox, no subprocess overhead
- Cons: immature ecosystem, language constraints
- Rejected: deferred to IDEAS.md post-MVP

### Native `.so` plugins

- Pros: lower latency
- Cons: Linux-only, no isolation, ABI fragility
- Rejected

### Hot-reload without restart

- Pros: better DX for plugin authors
- Cons: go-plugin client lifecycle complexity, security review surface
- Rejected for MVP (UC3.1 explicitly requires restart)

### Plugin logs on disk (`{volume}/logs/`)

- Pros: aligns with volume portability
- Cons: new on-disk schema, harder UI aggregation
- Rejected for MVP; PostgreSQL feed with optional disk mirror post-MVP

## Consequences

- Plugin authors need Go (or any go-plugin net/rpc-capable language) + manifest; SDK stub in `pkg/pluginsdk/`
- ADR required before expanding HostAPI surface
- Phase 1.4 wires `task.executed` emission into the same bus
- Docker mounts `plugins/` read-only via `PLUGINS_PATH`
- REST API responses omit host filesystem paths (`binary_path`, `manifest_path`)

## Security rules

- HostAPI `GetVolumeConfig` returns filters/name/quota only — no raw disk paths
- Plugins write only to `{volume}/plugins/{plugin-id}/`
- Enable/disable is admin-only
- Plugin validation is audit-only; core filter rejects bad uploads before save
- Plugin registry API does not expose server filesystem paths to clients
