# Spec: FTP & WebDAV — native OS access to volumes

> **Status:** Implemented  
> **Scope:** Expose each volume's `userdata/` tree via WebDAV (HTTP, same port) and FTP (dedicated port), so users can mount volumes as network drives or use any FTP client — without leaving lcloud's auth, quota, filter, and event-bus semantics.  
> **Prerequisite:** Phase 1.1 (volumes + file ops), user-owned volumes (volume access rules), volume explorer file ops (move/rename/delete/index wiring).  
> **Sources:** [VISION.md — Stage 2](../../../VISION.md), [IDEAS.md — FTP / WebDAV](../../../IDEAS.md), [docs/project.md](../../project.md)

---

## Assumptions (correct me now or I proceed)

1. **Core feature, not a plugin** — VISION Stage 2 lists WebDAV and FTP as platform capabilities; they ship as `internal/protocols/` modules wired in `cmd/server/main.go`.
2. **Per-volume opt-in** — Each volume independently enables WebDAV and/or FTP; disabled by default on create.
3. **Root = `userdata/` only** — Clients never see `cache/`, `plugins/`, `.volume.json`, or volume root; same boundary as the web file API.
4. **Reuse lcloud credentials in v1** — Email + password (bcrypt, same as web login). No separate app-password table in v1 (post-MVP hardening).
5. **Same access rules as the web UI** — Volume owner and `admin` can connect; other users get auth failure (401), not 403 after login.
6. **WebDAV on the existing HTTP port** — Mounted at `/dav/volumes/:volumeId/` on port 8080; TLS is the deployer's reverse-proxy responsibility (same as REST API).
7. **Single shared FTP listener** — One FTP server process on port `2121`; username is the **volume UUID** (password = lcloud password). Avoids one TCP port per volume.
8. **`.volume.json` extension required** — Protocol flags live in volume config (portable, re-importable). Needs **ADR-004** before implementation.
9. **Instance kill switch** — Global `protocols_enabled` (or separate webdav/ftp flags) in `InstanceSettings`; when off, listeners are down regardless of per-volume flags.
10. **v1 WebDAV methods** — `OPTIONS`, `PROPFIND`, `GET`, `PUT`, `DELETE`, `MKCOL`, `MOVE` (no `COPY`, no `LOCK`/`UNLOCK`).
11. **v1 FTP commands** — `USER`/`PASS`, `PWD`, `CWD`, `LIST`/`NLST`, `RETR`, `STOR`, `DELE`, `MKD`, `RMD`, `RNFR`/`RNTO`, `SIZE`, `MDTM`, `TYPE`, `PASV`/`EPSV`, `QUIT`. No `APPE`, no FTPS in v1.
12. **All mutations go through `internal/volume/`** — Protocol adapters call `FileService` / `file_ops.go`; no direct disk writes from protocol handlers.
13. **Events unchanged** — Upload/delete/move/rename via FTP or WebDAV emit the same bus events as the web API (`file.uploaded`, etc.).
14. **UI = settings panel only** — Toggle + connection instructions (URL, username format); no in-app FTP/WebDAV client, no sync/conflict UI (explicitly out of explorer spec).
15. **Single PR delivery** — WebDAV, FTP, UI, ADR, and tests ship together in one PR.

---

## Objective

Homelab users want to edit files in a volume from Finder, Explorer, or VS Code — without downloading through the browser. FTP and WebDAV are the standard protocols OS clients already speak.

**Who:** Self-hosters who run lcloud on a local network or VPS and know how to mount a network drive (VISION Principle 3).

**Pain today:** Files are only reachable via the web UI or REST API. Native tools (Windows « Map network drive », macOS Finder « Connect to Server », Linux `davfs2`, FileZilla) cannot attach to a volume.

**Success looks like:**

| ID | Story |
|---|---|
| P1 | As a volume owner, I enable WebDAV on my volume and mount `https://host/dav/volumes/{id}/` in my OS file manager. |
| P2 | As a volume owner, I enable FTP, connect with `{volumeUuid}` / password, and see only that volume's files. |
| P3 | As a user, uploads and deletes via mounted drive respect quota, extension filters, and update the Bleve index. |
| P4 | As an admin, I can disable all protocol access instance-wide from Settings. |
| P5 | As a user, I see connection instructions (URL, username) in the volume settings panel when a protocol is enabled. |

**Out of scope v1:**

- FTPS / WebDAV over implicit TLS inside the Go process (TLS terminates at reverse proxy for WebDAV; plain FTP only in v1)
- App passwords, per-protocol tokens, or OAuth
- Volume sharing across users (only owner + admin, same as today)
- `COPY` (WebDAV), `APPE` (FTP), resumable/chunked upload tuning
- Sync engine, conflict resolution, offline cache, version history
- SIEM / audit log beyond existing plugin events
- Changing volume directory layout (only `.volume.json` schema extension via ADR)

---

## Architecture

### High-level

```
OS client ──WebDAV──► Gin /dav/* ──► protocols/webdav ──► FileService ──► userdata/
OS client ──FTP────► TCP :2121 ──► protocols/ftp ──────► FileService ──► userdata/
                              │
                              └── protocols/auth (email+password → Claims + volume gate)
```

### Module layout

| Module | Responsibility |
|---|---|
| `internal/protocols/auth.go` | Validate email/password; build `auth.Claims`; resolve volume from FTP username or WebDAV path |
| `internal/protocols/access.go` | `authorizeVolume(claims, volumeID)` — delegates to volume service rules |
| `internal/protocols/webdav/` | `webdav.Handler` + custom `webdav.FileSystem` backed by `FileService` |
| `internal/protocols/ftp/` | FTP server (driver implementing file ops via `FileService`) |
| `internal/protocols/config.go` | Read per-volume protocol flags from `VolumeConfig.Protocols` |
| `internal/api/protocol_handlers.go` | Admin/user REST: get/set protocol settings, connection info |
| `web/src/components/volumes/ProtocolSettingsPanel.tsx` | Toggle + copy-friendly connection strings |

### `.volume.json` extension (ADR-004)

Add optional block — omitted means disabled:

```json
{
  "protocols": {
    "webdav": { "enabled": false },
    "ftp": { "enabled": false }
  }
}
```

Rules:

- Written via existing `WriteVolumeConfig` atomic rename path
- Mirrored in PostgreSQL `volumes` table as JSONB column `protocols` for query/list (cache from disk on import/sync — same pattern as `filters`)
- On volume create: both disabled
- Toggle via `PATCH /api/volumes/:id/protocols` (owner or admin)

### Instance settings extension

Extend `InstanceSettings`:

| Field | Default | Effect |
|---|---|---|
| `protocols_webdav_enabled` | `false` | Master switch for WebDAV handler |
| `protocols_ftp_enabled` | `false` | Master switch for FTP listener |

Admin-only `PATCH /api/admin/settings`. Per-volume enable has no effect when instance switch is off.

### WebDAV contract

| Item | Value |
|---|---|
| Base path | `/dav/volumes/:volumeId/` |
| Auth | HTTP Basic (`Authorization: Basic …`) |
| Root | Maps to `userdata/` (trailing slash = volume root) |
| Depth | `PROPFIND` depth 0 and 1 (no infinite recursion requirement) |
| Overwrite | `PUT` overwrites existing file (same as web upload) |
| Move | `MOVE` with `Destination` header → `FileService.Move` (files only, same as web v1) |
| Directory rename | `MOVE` on collection where parent unchanged → `RenameEntry` |
| Errors | 401 unauthorized, 404 missing/disabled volume, 403 disabled instance, 507 quota, 415 filter rejection |

**Mount examples (shown in UI):**

- macOS Finder: `https://host/dav/volumes/{volumeId}/`
- Windows: « Map network drive » → `https://host/dav/volumes/{volumeId}/`
- Linux: `davfs2` → same URL

### FTP contract

| Item | Value |
|---|---|
| Port | `2121` default (non-privileged; env `FTP_PORT`) |
| Passive range | `30000–30010` default (`FTP_PASV_MIN`, `FTP_PASV_MAX`) |
| Advertised PASV address | `FTP_PASV_ADDRESS` (required behind NAT/docker) |
| Username | **Volume UUID** — e.g. `550e8400-e29b-41d4-a716-446655440000` |
| Password | lcloud account password |
| Root after login | `userdata/` of parsed volume |
| TLS | None in v1 (document security warning; FTPS deferred) |

### Auth flow

```
1. Client sends credentials
2. protocols/auth validates email + bcrypt password → User + Claims
3. Resolve volumeId (WebDAV: URL param; FTP: USER = UUID)
4. Check instance master switch + volume.protocols.{webdav|ftp}.enabled
5. volume.Service.authorizeVolume(claims, vol) — owner or admin
6. Bind session context {claims, volumeId} for subsequent ops
```

Rate-limit failed auth: 5 failures / minute / IP (in-memory counter v1; Redis deferred).

---

## Tech Stack

See [docs/project.md — Tech Stack](../../project.md#tech-stack).

**New dependencies (require ADR approval + Context7 validation before merge):**

| Library | Purpose | Notes |
|---|---|---|
| `golang.org/x/net/webdav` | WebDAV handler + interface | Extended stdlib; mature |
| `github.com/fclairamb/ftpserverlib` | FTP server framework | Driver pattern; verify age policy at `/plan` time |

No new frontend dependencies beyond existing shadcn patterns.

---

## Commands

### Development

```bash
# Backend tests (protocol packages)
go test ./internal/protocols/... ./internal/volume/...

# Full backend
go test ./internal/...

# Frontend
cd web && npm run build

# Local run with protocols (after implementation)
FTP_PORT=2121 FTP_PASV_ADDRESS=127.0.0.1 docker compose up --build
```

### Manual verification

```bash
# WebDAV — list root (replace credentials and volume id)
curl -u 'user@example.com:password' -X PROPFIND \
  http://localhost:8080/dav/volumes/{volumeId}/ -H 'Depth: 1'

# FTP
lftp -u '{volumeId},password' -p 2121 localhost
```

---

## Project Structure

See [docs/project.md — Project structure](../../project.md#project-structure).

Additions:

```
internal/protocols/
├── auth.go
├── auth_test.go
├── access.go
├── config.go
├── webdav/
│   ├── filesystem.go      # webdav.FileSystem → FileService
│   ├── handler.go         # http.Handler mount
│   └── filesystem_test.go
└── ftp/
    ├── driver.go          # ftpserverlib driver
    ├── server.go          # lifecycle start/stop
    └── driver_test.go

internal/api/
└── protocol_handlers.go   # PATCH protocols, GET connection info

docs/adr/
└── 004-volume-protocols-config.md   # .volume.json + InstanceSettings extension

docs/features/ftp-webdav/
└── SPEC.md                # this file

web/src/components/volumes/
└── ProtocolSettingsPanel.tsx
```

---

## REST API

| Method | Path | Auth | Body / response |
|---|---|---|---|
| `GET` | `/api/volumes/:id/protocols` | owner, admin | `{ webdav: { enabled }, ftp: { enabled }, connection: { webdav_url, ftp_host, ftp_port, ftp_username_format } }` |
| `PATCH` | `/api/volumes/:id/protocols` | owner, admin | `{ webdav?: { enabled }, ftp?: { enabled } }` → updated config |
| `GET` | `/api/admin/settings` | admin | includes `protocols_webdav_enabled`, `protocols_ftp_enabled` |
| `PATCH` | `/api/admin/settings` | admin | `{ protocols_webdav_enabled?, protocols_ftp_enabled? }` |

Connection info fields are derived server-side (host from `X-Forwarded-Host` or request Host; never trust client for URL construction).

---

## Code style

Protocol adapters are thin — delegate to existing services:

```go
// internal/protocols/webdav/filesystem.go (illustrative)
func (fs *volumeFS) OpenFile(name string, flag int, perm os.FileMode) (webdav.File, error) {
    rel, err := fs.paths.Clean(name)
    if err != nil {
        return nil, err
    }
    switch {
    case flag&os.O_CREATE != 0:
        // PUT / upload path → fs.files.Upload(fs.claims, fs.volumeID, dir, base, reader, size)
    default:
        // GET → fs.files.OpenContent(...)
    }
}
```

Conventions:

- Handlers never call `os.Open` on volume paths directly
- FTP driver and WebDAV filesystem share `protocols/auth` — no duplicated bcrypt logic
- Errors map to protocol-native codes (FTP 550 for quota/filter, WebDAV 507/415)

---

## Testing strategy

| Layer | What | Target |
|---|---|---|
| Unit | `protocols/auth` username parsing, access denied paths | 80%+ |
| Unit | WebDAV filesystem: PROPFIND listing, PUT quota breach, path traversal rejected | 80%+ |
| Unit | FTP driver: login format, CWD/LIST/STOR delegation mocks | 80%+ |
| Integration | Temp volume dir + real FileService: upload via WebDAV PUT → file on disk + index entry | Required |
| Integration | FTP STOR + RETR round-trip | Required |
| Manual | Mount on macOS Finder + Windows Explorer | Checklist in test plan |

Framework: Go `testing` + `testify`; no new frontend tests required beyond optional panel render smoke.

---

## Boundaries

**Always:**

- Route all file mutations through `internal/volume/` (`FileService`, `file_ops.go`)
- Emit plugin bus events on mutations
- Respect `PathResolver` for traversal safety
- Update `.volume.json` before PostgreSQL on protocol toggle
- Document FTP plain-text warning in UI

**Ask first:**

- Adding dependencies not listed above
- FTPS or embedded TLS
- App-password table / separate credential model
- Changing `.volume.json` fields beyond `protocols` block
- Exposing FTP on port 21 (privileged)

**Never:**

- Direct disk I/O from `internal/protocols/` outside volume package
- Bypass `VolumeIndexer` on file changes
- Store protocol secrets in `.volume.json`
- Enable protocols by default on new volumes

---

## Docker & environment

`.env.example` additions:

```env
# Protocols (optional — off by default at instance level via admin settings)
FTP_PORT=2121
FTP_PASV_MIN=30000
FTP_PASV_MAX=30010
FTP_PASV_ADDRESS=127.0.0.1
```

`docker-compose.yml` — expose FTP port and passive range on `app` service:

```yaml
ports:
  - "${APP_PORT:-8080}:8080"
  - "${FTP_PORT:-2121}:${FTP_PORT:-2121}"
  - "${FTP_PASV_MIN:-30000}-${FTP_PASV_MAX:-30010}:${FTP_PASV_MIN:-30000}-${FTP_PASV_MAX:-30010}"
```

Document NAT: deployer must set `FTP_PASV_ADDRESS` to the host IP clients can reach.

---

## UI (ProtocolSettingsPanel)

Location: volume detail page — new section below explorer or settings tab (follow `design/DESIGN.md`).

Contents when user opens volume settings:

- Toggle **WebDAV** / **FTP** (disabled if instance master switch off — show admin message)
- Read-only connection card with copy buttons:
  - WebDAV URL
  - FTP host:port
  - Username format with volume id pre-filled
- Warning callout: FTP credentials travel in plain text unless tunnelled; prefer WebDAV behind HTTPS

Admin Settings page: two master switches for WebDAV and FTP.

---

## Success criteria

- [ ] SC-P1 ADR-004 accepted; `.volume.json` supports `protocols` block; import/sync preserves flags
- [ ] SC-P2 Instance master switches default off; admin can enable globally
- [ ] SC-P3 Volume owner can enable WebDAV; PROPFIND + GET + PUT + DELETE + MKCOL work on `userdata/`
- [ ] SC-P4 WebDAV MOVE renames/moves files consistent with web API; emits events
- [ ] SC-P5 FTP login with volume UUID works; LIST/RETR/STOR/DELE/MKD/RMD/RNFR/RNTO work
- [ ] SC-P6 Quota and extension filters enforced on protocol uploads (test: reject over-quota STOR)
- [ ] SC-P7 Non-owner non-admin receives 401/530 on connect
- [ ] SC-P8 ProtocolSettingsPanel shows correct URLs and toggles persist across reload
- [ ] SC-P9 `go test ./internal/...` and `cd web && npm run build` pass

---

## Implementation plan (preview for `/plan`)

Recommended vertical slices:

1. **ADR-004 + config model** — `.volume.json`, DB cache, REST PATCH/GET, instance settings
2. **WebDAV slice** — auth middleware, filesystem, mount in Gin, tests, UI connection card (WebDAV only)
3. **FTP slice** — server lifecycle, driver, docker ports, UI FTP fields, integration tests

WebDAV before FTP: same auth and volume gate; WebDAV avoids passive-port operational complexity.

---

## Open questions (resolved)

| # | Question | Decision |
|---|---|---|
| OQ1 | Reuse main password vs app passwords in v1? | **Main password** (bcrypt) |
| OQ2 | Ship WebDAV and FTP in one PR or two? | **Single PR** |
| OQ3 | FTP username format? | **Volume UUID** (not `{email}#{volumeId}`) |
| OQ4 | Should `admin` mount any user's volume via protocols? | **Yes** — same as web |
| OQ5 | Global default: instance switches off, or on when admin enables first volume? | **Instance off; per-volume off** |

---

## Related

- [IDEAS.md — FTP server per volume](../../../IDEAS.md) — userdata-only root
- [IDEAS.md — WebDAV support](../../../IDEAS.md) — native OS mount
- Explorer spec ([SPEC.md](../../../SPEC.md)) — explicitly excludes protocol UI in explorer v1; this feature owns protocol settings
- Future: app passwords, FTPS, WebDAV `COPY`, volume sharing → IDEAS.md, not this spec
