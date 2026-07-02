# Ideas — lcloud

> Parking lot for future ideas, improvements, and planned integrations.
> No sorting, no prioritization — a living list.
> Ideas that become concrete leave this file for either VISION.md (if strategic)
> or direct implementation via `/spec` in the IDE (if tactical).

---

## Post-MVP — adding implementation ideas

The MVP is shipped. New work is **feature-scoped** (`/spec` → `/build` → `/ship`), not phase-based.
When you discover a gap while using the product — placeholder page, API without UI, deferred behavior —
**capture it here first** before writing a spec.

### When to add an entry

- A sidebar route or page still shows **Soon** or a placeholder after MVP
- Backend capability exists but has **no web UI** (or the reverse)
- Something was **explicitly deferred** in a phase spec and is now worth scheduling
- Exploration surfaced a feature that does **not** belong in VISION.md (tactical, not strategic)

Do **not** add entries for bugs, typos, or one-line fixes — use `/build` or `fix/*` directly.

### Entry template

Each idea uses the same shape (see **In-app notification system** or **User management UI** below):

```markdown
### Short title
**Context:** What exists today, what is missing, where it was deferred.
**Value:** Why it matters for the operator (plain language).
**Estimated effort:** Quick | Medium | Heavy
**Dependencies:** Shipped modules or prior ideas
**Date:** YYYY-MM-DD

**Possible scope (post-MVP `/spec`):**  ← optional, for ideas ready to spec soon
- Backend / API bullets
- UI bullets
- Out of scope for v1
```

### Lifecycle

1. **Park** — append a new entry to `## Ideas` below (English, same template).
2. **Prioritize** — pick when ready; no ordering required in this file.
3. **Spec** — run `/spec` on the idea title + context; agent reads VISION.md, docs/project.md, and this entry.
4. **Ship** — implementation removes or archives the entry once the feature is merged (or move to VISION.md if it changed product direction).

Strategic shifts (new audience, new deployment model) → **VISION.md**, not IDEAS.md.

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

### User management UI (admin)
**Context:** Phase 0 shipped admin-only user creation via API (`POST /api/admin/users`) and a placeholder page at `/settings/users` (sidebar link with "Soon" badge). Post-MVP, admins still create users via curl or external tooling — no list, role change, or disable flow in the web UI.
**Value:** Manage team or family accounts from the control panel: see who has access, create users, assign `admin` / `user` roles, and revoke access without leaving the browser.
**Estimated effort:** Medium
**Dependencies:** Phase 0 auth (JWT, roles, admin seed); MVP complete
**Date:** 2026-07-02

**Possible scope (post-MVP `/spec`):**
- Backend: `GET /api/admin/users`, optional `PATCH` (role, disable) and `DELETE` or soft-disable
- UI: replace placeholder at `/settings/users` — user table, create-user form, role badges, admin-only route (already wired)
- Out of scope for v1 of this idea: self-service signup, password reset email, OAuth/SSO, per-volume ACL (volume ownership is separate)

---

### Volume file explorer UX redesign
**Context:** MVP ships a functional but minimal file browser on `/volumes/:id` — flat table (`FileBrowser`), breadcrumb only, a permanent drag-and-drop upload banner above the list, and an always-visible "New folder" text field + button. Operator feedback (post-MVP): the layout does not resemble any familiar product; upload and folder creation **pollute** the main area instead of living in standard controls; files cannot be dropped **onto** the file list itself. Navigation is click-folder-name-in-table only — no sidebar tree, view switcher, or inline rename. FTP/WebDAV are separate ideas; the web UI should feel like a known cloud drive on day one.
**Value:** Reuse interaction patterns from **OneDrive, Google Drive, Dropbox**, plus desktop habits (Windows Explorer inline rename): users already know where upload, "New folder", views, and rename live — no learning curve for a custom lcloud layout.
**Estimated effort:** Medium to Heavy
**Dependencies:** Phase 1.1 volumes + file API (shipped); **rename/move API** not exposed to web yet (`MoveFileInternal` exists for macros only); `design/DESIGN.md` for UI pass
**Date:** 2026-07-02

**Operator pain points to fix:**
- Remove the permanent upload zone — **drop files directly on the file list / folder area** (visual feedback on drag-over)
- Remove the always-visible folder name text box — **"New" / "+" control** opens create-folder flow (inline or small dialog)
- Overall layout = **standard cloud file manager**, not admin table + forms stack

**Operator decisions (direction for `/spec`):**
- **Sidebar folder tree** — yes; lazy-loaded tree on the left, contents on the right (Drive / Explorer hybrid)
- **Multiple views** — switchable **list** and **grid** (thumbnails); view preference remembered in the browser
- **Flexible list view** — user can show/hide columns, **resize** column widths, **reorder** columns; **launch columns:** name (required), **size**, **modified date** — type and others optional via column picker
- **Layout prefs** — **localStorage first** (view mode, column set, order, widths); note for later: evaluate whether **server-side persistence** per user is worth it once the UI is in daily use
- **Inline rename** — Windows Explorer style: F2, right-click → Rename, or slow double-click on name; label becomes editable in place; files and folders
- **Drag-and-drop move** — drag file(s) onto a folder in the tree or list to move them (same mental model as desktop/cloud drives); distinct from upload drop (external files → volume)

**Possible scope (post-MVP `/spec`):**
- **Layout:** left folder tree + top toolbar (New, Upload, view toggle, column picker) + main contents pane
- **Upload:** drop external files onto contents pane; optional Upload button in toolbar
- **Move:** drag internal items onto folder targets (tree node or folder row); visual highlight on valid drop target
- **Navigation:** tree selection + breadcrumb/path bar, double-click folder to open
- **Views:** list (customizable columns) + grid (icons/thumbnails)
- **Actions:** context menu (download, delete, rename); inline rename as above
- **Backend:** expose rename + move for web UI (wrap `MoveFileInternal` / volume layer); emit `file.moved` / `file.renamed` on event bus
- **Out of scope for v1:** FTP/WebDAV UI, sync/conflict UI, in-browser editor, server-side layout prefs (deferred — test need after v1)

**Follow-up to validate post-ship:**
- Is server-side storage of view/column preferences useful (multi-device, shared admin workstation)?

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
