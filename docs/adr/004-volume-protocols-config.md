# ADR-004: Volume protocol access (WebDAV & FTP)

## Status

Accepted

## Date

2026-07-03

## Context

Homelab users need native OS access to volume files (Finder, Explorer, FileZilla). WebDAV and FTP must reuse lcloud auth, quota, filters, indexing, and event-bus semantics without bypassing `internal/volume/`.

Requirements:

- Per-volume opt-in for WebDAV and FTP (disabled on create)
- Instance master switches (admin-only, default off)
- Portable config in `.volume.json` (re-importable volumes)
- PostgreSQL JSONB cache for list/query performance
- FTP: shared port 2121, username = volume UUID, password = lcloud password
- WebDAV: `/dav/volumes/:volumeId/`, HTTP Basic auth (email + password)
- Owner and admin access (same rules as web UI)

## Decision

### `.volume.json` extension

Add optional `protocols` block — omitted means disabled:

```json
{
  "protocols": {
    "webdav": { "enabled": false },
    "ftp": { "enabled": false }
  }
}
```

Written via existing atomic `WriteVolumeConfig` path. On toggle: disk first, then PostgreSQL (same pattern as usage sync).

### PostgreSQL cache

Add JSONB column `protocols` on `volumes`, synced from disk on create/patch/import.

### Instance settings

Extend `InstanceSettings`:

| Field | Default | Effect |
|---|---|---|
| `protocols_webdav_enabled` | `false` | Master switch for WebDAV handler |
| `protocols_ftp_enabled` | `false` | Master switch for FTP listener |

Per-volume enable has no effect when the instance switch is off.

### Auth

- **WebDAV:** HTTP Basic — email + lcloud password (bcrypt). Volume ID from URL path.
- **FTP:** USER = volume UUID, PASS = lcloud password. `AuthenticateForVolume(volumeID, password)` tries owner, then any matching admin password.

No app-password table in v1.

### Module layout

`internal/protocols/` — thin adapters over `FileService`; no direct disk I/O outside `internal/volume/`.

Dependencies: `golang.org/x/net/webdav`, `github.com/fclairamb/ftpserverlib`.

### Runtime config

| Env | Default |
|---|---|
| `FTP_PORT` | `2121` |
| `FTP_PASV_MIN` | `30000` |
| `FTP_PASV_MAX` | `30010` |
| `FTP_PASV_ADDRESS` | `127.0.0.1` |

## Consequences

- Positive: Portable volume config; same access/quota/index/event semantics as web API
- Positive: Single FTP port with UUID username avoids per-volume TCP ports
- Negative: FTP credentials in plain text (documented in UI; FTPS deferred)
- Negative: Admin FTP auth matches first admin with same password (acceptable for homelab)
- Migration: GORM AutoMigrate adds columns; existing volumes get protocols disabled via zero value

## Alternatives considered

| Alternative | Rejected because |
|---|---|
| `{email}#{volumeId}` FTP username | Validated: UUID-only is simpler for FileZilla users |
| App passwords in v1 | Deferred; main password sufficient for homelab MVP |
| Plugin-based protocols | VISION Stage 2 lists WebDAV/FTP as platform capabilities |
