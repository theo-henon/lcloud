# Ideas — lcloud

> Parking lot for future ideas, improvements, and planned integrations.
> No sorting, no prioritization — a living list.
> Ideas that become concrete leave this file for either VISION.md (if strategic)
> or direct implementation via `/spec` in the IDE (if tactical).

---

## Ideas

### FTP server per volume
**Context:** Surfaced during exploration — expose `./userdata` of a volume as an FTP server.
**Value:** Native OS integration on Windows, Linux, macOS without installing a dedicated client. Access restricted to `./userdata` only.
**Estimated effort:** Medium
**Dependencies:** Phase 1.1 (volumes)
**Date:** 2026-07-01

---

### WebDAV support
**Context:** Surfaced during exploration — mount volumes as network drives natively on Windows, macOS, Linux.
**Value:** Transparent OS-level integration; edit files in a volume without downloading them.
**Estimated effort:** Medium
**Dependencies:** Phase 1.1 (volumes)
**Date:** 2026-07-01

---

### Kubernetes deployment option
**Context:** Discussed during stack planning — Docker Compose chosen for MVP; K8s deferred as disproportionate for a solo self-hosted tool.
**Value:** Multi-node deployment, production-grade orchestration for larger setups.
**Estimated effort:** Heavy
**Dependencies:** Stable core architecture
**Date:** 2026-07-01

---

### Macro: record & replay interactions (ImageJ-style)
**Context:** Original macro concept — record UI interactions and replay them as a script. Deferred in favor of the predefined task system for MVP.
**Value:** No-code automation for users who want macros without scripting.
**Estimated effort:** Heavy
**Dependencies:** Stable web UI + task system (Phase 1.4)
**Date:** 2026-07-01

---

### Custom scripting language for macros
**Context:** Alternative to the predefined macro vocabulary — a lightweight scripting DSL for power users.
**Value:** Express complex automation logic beyond the fixed macro list.
**Estimated effort:** Heavy
**Dependencies:** Phase 1.4 (task system stable)
**Date:** 2026-07-01

---

### Backup and archiving system
**Context:** Mentioned during exploration — periodic volume backups with restoration.
**Value:** Safety net against data loss; point-in-time restore of a volume.
**Estimated effort:** Medium
**Dependencies:** Phase 1.1 (volumes)
**Date:** 2026-07-01

---

### Per-volume and global logs
**Context:** Mentioned during exploration — `./logs/` folder structure planned in volume layout (stub exists), global log dashboard post-MVP.
**Value:** Auditability, debugging, and operational history of file and task events.
**Estimated effort:** Quick
**Dependencies:** Phase 1.4 (tasks emit actions to log)
**Date:** 2026-07-01

---

### Full-text content search via Bleve
**Context:** Discussed during indexer design — Bleve is in use for metadata search (MVP). Content extraction (text files, PDFs, OCR for images, audio/video metadata) deferred.
**Value:** Search inside file contents, not just filenames and metadata.
**Estimated effort:** Medium (text/PDF) to Heavy (OCR pipeline)
**Dependencies:** Phase 1.2 (Bleve running), content extraction libraries
**Date:** 2026-07-01

---

### Volume encryption — filesystem level (gocryptfs / EncFS)
**Context:** Encryption stub present in `.volume.json` from Phase 1.1. Option A: encrypt entire `./userdata/` via a FUSE-mounted encrypted filesystem.
**Value:** Data at rest protection — volume contents unreadable without the passphrase even with direct disk access.
**Estimated effort:** Medium
**Dependencies:** Phase 1.1 (volume structure stable)
**Date:** 2026-07-01

---

### Volume encryption — file-level AES
**Context:** Option B for encryption — encrypt each file individually before writing to disk.
**Value:** More granular control; no FUSE dependency; cross-platform.
**Estimated effort:** Medium
**Dependencies:** Phase 1.1 (volume structure stable)
**Date:** 2026-07-01

---

### VS Code integration plugin
**Context:** Mentioned during exploration as an example of what a community plugin could enable.
**Value:** Edit files in a volume directly from VS Code without mounting a network drive.
**Estimated effort:** Heavy
**Dependencies:** Phase 1.3 (plugin system stable)
**Date:** 2026-07-01

---

### AI tools plugins
**Context:** Mentioned during exploration as a Stage 2/3 community plugin example.
**Value:** Summarize documents, classify images, transcribe audio — all operating directly on volume files.
**Estimated effort:** Medium (depends on plugin scope)
**Dependencies:** Phase 1.3 (plugin system stable)
**Date:** 2026-07-01

---

### Cross-volume copy / move macro
**Context:** Post-MVP macro — move or copy files between different volumes.
**Value:** Volume reorganization without manual download/upload.
**Estimated effort:** Quick
**Dependencies:** Phase 1.4 (task system)
**Date:** 2026-07-01

---

### In-app notification system (alert center)
**Context:** Phase 1.4 ships `alert_usage` and other task events on the internal event bus (`volume.alert.usage`, `task.executed`, etc.), but the MVP has no user-facing notification layer — only task run history messages. Users expect a visible signal when a quota alert fires, not just a line in execution logs.
**Value:** A first-class notification channel in the web UI: in-app toasts, a notification center (bell icon + unread list), optional persistence and read/unread state. Task alerts (quota usage, failed runs) and plugin-emitted events would publish into this system so operators see what happened without digging into task history or building a plugin first.
**Estimated effort:** Medium
**Dependencies:** Phase 1.4 (task system + event bus); optionally PostgreSQL model for notification inbox
**Date:** 2026-07-02

**Possible scope (post-MVP `/spec`):**
- Notification model + API (`GET /notifications`, mark read, dismiss)
- UI: toast on new alert, sidebar/header bell with history
- Subscribers: `volume.alert.usage`, `task.failed`, plugin custom events
- Out of scope for v1 of this idea: email, push, SMS (those stay plugin/webhook territory)

---

### Webhook notifications macro
**Context:** Post-MVP macro — send an HTTP POST to an external URL on a condition (e.g. volume > 90% full).
**Value:** Integration with external systems (Slack, Discord, custom automation webhooks).
**Estimated effort:** Quick
**Dependencies:** Phase 1.4 (task system)
**Date:** 2026-07-01

---

### File deduplication
**Context:** SHA256 hash of each file is indexed in Bleve from Phase 1.1 — deduplication detection is a natural next step.
**Value:** Reclaim disk space from duplicate files within `./userdata`.
**Estimated effort:** Medium
**Dependencies:** Phase 1.1 (hash in Bleve index)
**Date:** 2026-07-01

---

### Multiple plugin runtimes (WASM support)
**Context:** Discussed during plugin design — go-plugin (subprocess) chosen for MVP; WASM support deferred as a second runtime option.
**Value:** Language-agnostic plugins without a compiled binary; improved sandboxing.
**Estimated effort:** Heavy
**Dependencies:** Phase 1.3 (plugin system stable)
**Date:** 2026-07-01

---

### Desktop / CLI interface
**Context:** Web UI chosen for MVP; other interface types explicitly deferred.
**Value:** Native desktop experience; CLI for power users and scripting/automation outside the web UI.
**Estimated effort:** Heavy
**Dependencies:** Stable REST API (Phase 1.1+)
**Date:** 2026-07-01
