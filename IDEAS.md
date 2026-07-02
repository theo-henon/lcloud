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
**Context:** Discussed during indexer design — Bleve is in use for metadata search (MVP). Content extraction (text files, PDFs, OCR for images, audio/video metadata) deferred. Powers the **Content** category inside **Unified Spotlight search** — same palette, searches inside file bodies, not just names/paths.
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

**Product constraint (not up for debate in this idea):**
- **Admin-provisioned accounts only** — the admin creates every user account; there is no public signup, no invitation link, no self-registration flow
- **No OAuth / social login** (Google, GitHub, etc.) — email + password only, aligned with self-hosted homelab use
- Login page stays **sign-in only**; account creation lives exclusively on `/settings/users` (admin)

**Possible scope (post-MVP `/spec`):**
- Backend: `GET /api/admin/users`, `PATCH` (role, disable), optional soft-disable; admin-initiated password reset (set new password — no email pipeline)
- UI: replace placeholder at `/settings/users` — user table, **admin-only create-user form**, role badges
- Safeguards: cannot demote/delete the last admin; optional disable instead of hard delete
- **Explicitly out of scope** (this project, for now): self-service signup, registration page, OAuth/SSO, password-reset email, per-volume ACL (volume ownership is a separate idea)

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

### Explorer grid — single-click vs double-click interaction
**Context:** Volume explorer redesign (shipped) opens folders and files on **double-click** in grid and list views. Single click on a grid card only shows hover styling — no selection state, no open. Operator feedback (2026-07-02): interaction should feel more responsive; either **single-click** (select or open) or **double-click** should clearly drive the primary action, aligned with desktop/cloud habits (Explorer: select on click, open on double-click; some products: single-click open in grid).
**Value:** Less friction in grid view — users know what one click vs two clicks do; optional path to single-click-open for touch/simple workflows without breaking power-user double-click.
**Estimated effort:** Quick
**Dependencies:** Volume explorer redesign (shipped); optional tie-in to **Multi-file selection and bulk actions** if single-click = select
**Date:** 2026-07-02

**Possible scope (post-MVP `/spec`):**
- **Default (desktop):** single-click selects (visual highlight); double-click opens folder or preview
- **Alternative or setting:** single-click opens in grid view (Drive-style); double-click still works
- Consistent behavior between list and grid; document choice in explorer prefs (localStorage)
- Out of scope: touch-only long-press menus (separate mobile pass)

---

### Admin-configurable max upload size
**Context:** The server rejects uploads above a fixed limit (**100 MB** by default, `104857600` bytes). This is set via the environment variable `MAX_UPLOAD_BYTES` at startup (`internal/config/config.go`, `.env.example`, `README.md`) — **not** in the volume explorer SPEC (which covers the upload queue and parallel slots only). Enforcement happens in `FileService.Upload`; the UI shows the generic error `upload exceeds max size`. Operator feedback: everyday large files (e.g. `.dmg` installers) fail until the env var is raised and the process restarted. The Settings page already stores instance prefs in PostgreSQL (`mask_disk_names`) but upload limit is not exposed there.
**Value:** Admin can see and adjust the cap from the control panel; upload queue errors can name the limit in plain language (e.g. "Max 100 MB per file"); no SSH to edit `.env` for homelab tweaks. Safer than an unbounded default.
**Estimated effort:** Medium
**Dependencies:** Settings API + `instance_settings` (shipped pattern); FileService must apply a runtime or hot-reloaded limit; optional `GET /settings` field for read-only display to all users
**Date:** 2026-07-02

**Workaround today:** set `MAX_UPLOAD_BYTES` in `.env` (bytes), e.g. `524288000` for 500 MB or `1073741824` for 1 GB, then restart Docker / the Go server.

**Possible scope (post-MVP `/spec`):**
- **Backend:** `max_upload_bytes` on `instance_settings`; admin-only `PATCH`; validate min/max bounds (e.g. 1 MB – 2 GB); FileService reads from settings service (with env as fallback / ceiling)
- **UI:** Settings → Security or Storage card — numeric input in MB, helper text, save; upload queue uses limit in error copy when API returns `UPLOAD_TOO_LARGE`
- **Out of scope v1:** per-volume upload limits (volume quota already exists separately)

---

### Explorer **New** menu — plugin-extensible create actions
**Context:** Volume explorer redesign (shipped) exposes a **New ▾** toolbar dropdown with a single built-in action: **New folder** (`ExplorerToolbar.tsx`). Operator direction: the menu should support **multiple create targets**, with **plugins** able to register entries — e.g. a Word plugin adds "Word document", an Excel plugin adds "Spreadsheet", similar to Google Drive / Microsoft 365 "New" menus. Today, creation is hard-coded in the frontend; plugins run on the event bus (`internal/plugin/`) but do not extend explorer UI. ADR 003 already defines a frontend `FileOpener` registry for *opening* files; no symmetric **CreateAction** / **NewMenuProvider** contract exists yet.
**Value:** Office and domain plugins can offer one-click "create blank file in current folder" without forking the explorer; core stays a thin shell (folder + upload + plugin contributions); aligns with lcloud's plugin-first extensibility model.
**Estimated effort:** Medium (frontend registry + API hook) to Heavy (plugin SDK + server-side file creation templates)
**Dependencies:** Volume explorer redesign (shipped); plugin system (Phase 1.3); optional parallel to **FileOpener** registry pattern in `web/src/lib/fileOpeners/`
**Date:** 2026-07-02

**Possible scope (post-MVP `/spec`):**
- **Frontend:** `registerCreateAction({ id, label, priority, canCreate?, create(ctx) })` merged into **New ▾** below core "New folder"; context = `{ volumeId, currentPath }`
- **Plugin bridge:** event or RPC so plugins publish create actions at runtime (e.g. subscribe to `explorer.create.menu` or extend plugin manifest)
- **Backend (if needed):** `POST .../files/create` with template bytes or empty-file MIME for plugin-created documents; reuse upload/quota/filter rules
- **Out of scope v1:** full in-browser Office editing; cloud template galleries; per-user menu customization UI

**Relationship:** mirrors **FileOpener** (open) ↔ **CreateAction** (new); complements plugin-backed openers in SPEC follow-up slice.

---

### Monitoring dashboard — visual upgrade
**Context:** MVP monitoring (`/monitoring`) shows disk cards, a volume table, and per-volume stats (used/quota, category bars, top MIME types). All usage bars use the same primary yellow — no color signal when a quota is nearly full. Data is **snapshot-only** (stats cache per volume, no history over time). The `alert_usage` task can fire on the event bus when a threshold is exceeded, but the monitoring UI does not surface alert state or tie into a notification center yet.
**Value:** Understand disk and volume health at a glance: colored quota bars (green → yellow → red), charts for file-type breakdown, and optional trends so operators see problems before uploads fail — without reading raw numbers.
**Estimated effort:** Medium (UI + thresholds) to Heavy (if usage history / time-series is included)
**Dependencies:** Phase 1.2 monitoring (shipped); optional link to in-app notification idea for quota alerts
**Date:** 2026-07-02

**Operator direction:**
- **Quota color gradient** — progress bar and/or stat text shifts by usage % (e.g. &lt;70% emerald, 70–85% warning, 85–95% orange, ≥95% rose); same logic on disk cards, volume list, and volume detail panel
- **Health status pills** — small green / yellow / red indicator on each disk card (and elsewhere as needed): obvious at-a-glance status (OK, warning, critical) derived from free-space thresholds and mount/readability; instance-level pill optional (API + DB reachable via `/api/health`)
- **Charts** — donut or bar chart for category breakdown (images, videos, documents…); optional stacked view of volumes per physical disk
- **At-a-glance clarity** — sort or badge volumes "at risk" when quota is high; show remaining space prominently, not only used bytes

**Possible scope (post-MVP `/spec`):**
- **UI v1 (no new backend):** threshold-colored bars everywhere quotas appear; health pills on disk cards; Recharts (or similar) for category/disk charts; global summary row (total used / total quota / volumes in warning)
- **UI v2:** align colors with `alert_usage` threshold_percent from tasks (or a default 80/90/95% palette)
- **History (heavier):** periodic usage snapshots (PostgreSQL or volume cache) → line chart "usage over 7/30 days"; growth hint ("full in ~N days" if trend stable)
- **Cross-links:** click volume in chart → volume detail; surface last `volume.alert.usage` event if notification system exists
- **Out of scope for v1:** Prometheus/Grafana export, per-file analytics, real-time streaming metrics

**Follow-up to validate post-ship:**
- Is usage **history** worth the storage/complexity, or are colored snapshots enough for homelab scale?

---

### Dashboard home — configurable widget layout
**Context:** `/dashboard` is still a Phase 0 placeholder (welcome message only). VISION Stage 1 calls for *"a dashboard that shows what's happening"*, but no post-MVP spec exists. Monitoring, Tasks, Plugins, and Volumes each have full pages — the home dashboard should aggregate them, not duplicate them.
**Value:** Land on a personal **cockpit**: mini-versions of the modules you care about, arranged how you want. Admin and regular users each get a layout that follows them on any device (phone, laptop) via server-stored preferences.
**Estimated effort:** Medium (fixed widget dashboard) to Heavy (drag-and-drop layout + server persistence + widget catalog)
**Dependencies:** MVP modules shipped; monitoring visual upgrade (health pills, colored quotas) feeds disk/volume widgets; optional notification idea for activity widget
**Date:** 2026-07-02

**Operator direction:**
- **Configurable layout** — admin or user adds, removes, resizes, and **repositions widgets** on a grid (drag-and-drop), like a lightweight personal homepage
- **Widget catalog** — each widget is a **mini view** of an existing area, e.g.:
  - Storage summary (disks + health pills, volumes at risk)
  - Recent / failed task runs + next scheduled runs
  - Plugin status strip
  - Quick actions (new volume, open monitoring)
  - System health (instance OK / degraded)
- **Server-side persistence** — layout + widget config saved **per user in PostgreSQL** (not localStorage), so the same dashboard on any browser/device; `GET` / `PATCH /api/users/me/dashboard` or dedicated preferences endpoint
- **Role-aware** — widget picker hides admin-only widgets for `user` role (e.g. all-instance summary vs own volumes only)

**Possible scope (post-MVP `/spec`):**
- **Phase A — curated home (lighter):** fixed set of widgets, default layout, no drag-and-drop yet; replaces placeholder
- **Phase B — full customization:** react-grid-layout (or similar), widget registry, per-user layout JSON in DB, edit mode toggle ("Customize dashboard")
- **Backend:** `user_dashboard_layout` model (user_id, layout JSON, updated_at); validate widget IDs server-side
- **Out of scope for v1:** third-party widget plugins, shared/dashboard templates between users, mobile-native widget SDK

**Relationship to other ideas:**
- Monitoring upgrade → disk health pills + quota colors appear in **Storage widgets** on both `/monitoring` and `/dashboard`
- In-app notifications → optional **Activity / alerts** widget later

---

### Control panel iconography pass
**Context:** MVP UI is text-heavy: sidebar nav items are labels only (no icons), toolbar/action buttons are mostly text ("Download", "Delete", "Create folder", "Logout"), and file rows lack type icons. `lucide-react` is already in `web/package.json` but unused in components. Operator feedback: the control panel feels sparse and harder to scan — icons beside nav entries and actions would improve recognition and visual hierarchy without changing behavior.
**Value:** Faster navigation and clearer affordances: users spot Volumes vs Tasks vs Settings at a glance; actions read like a familiar desktop/cloud app; overall polish aligned with `design/DESIGN.md` (icon buttons documented, yellow + dark surfaces).
**Estimated effort:** Quick to Medium
**Dependencies:** `design/DESIGN.md`; optional coordination with volume explorer redesign and dashboard widgets (icons should stay consistent)
**Date:** 2026-07-02

**Operator direction:**
- **Sidebar** — one icon per nav item (Dashboard, Volumes, Monitoring, Plugins, Tasks, Settings, Users); consistent size and muted color, primary tint when active
- **Actions** — icon + label (or icon-only with tooltip) on common buttons: upload, new folder, refresh, download, delete, run task, logout
- **Files & volumes** — folder / file-type icons in lists (image, video, document, generic file); optional disk/volume icons in cards
- **Empty & status states** — light illustration or icon where it helps (empty folder, loading, error)

**Possible scope (post-MVP `/spec`):**
- Define a small **icon map** (nav + file categories + actions) in one module; Lucide defaults, no custom SVG sprawl
- Roll out in passes: sidebar + header first, then volume/file browser, then tasks/plugins
- Accessibility: decorative icons `aria-hidden`; icon-only buttons keep `aria-label`
- **Out of scope:** custom brand icon set, animated icons, emoji-as-icons

---

### Unified Spotlight search (command palette)
**Context:** MVP search is **files-only**: `GlobalSearch` lives in the sidebar (Bleve metadata across volumes). Operator direction (macOS **Spotlight** reference): a **central search** invoked from anywhere — keyboard shortcut (e.g. `⌘K` / `Ctrl+K`) or header button — overlay modal, one query box, results grouped by **category**, not a deep file-only hunt in the sidebar.
**Value:** "Search everything in lcloud" from one place: jump to a file, task, plugin, volume, or page. Admin and user both use it; results respect existing permissions (user sees own volumes/tasks only). Familiar mental model like Spotlight / VS Code command palette / Raycast.
**Estimated effort:** Medium (v1 categories + UI) to Heavy (content search + unified backend index)
**Dependencies:** Phase 1.1 search API (files); tasks/plugins/volumes list APIs (shipped); **Full-text content search via Bleve** for content category; iconography pass for result icons
**Date:** 2026-07-02

**Search categories (decomposed):**

| Category | What it finds | MVP backend today |
|---|---|---|
| **Files** | Filenames, paths, MIME (all accessible volumes) | `GET /api/search` ✅ |
| **Content** | Text inside files (PDF, txt…) | ❌ → Full-text content search idea |
| **Volumes** | Volume name, disk | `GET /api/volumes` — client filter or new endpoint |
| **Tasks** | Task name, macro | `GET /api/tasks` — client filter or new endpoint |
| **Plugins** | Plugin name, status | `GET /api/plugins` — client filter or new endpoint |
| **Navigation** | Pages / actions ("Monitoring", "Settings", "New volume") | Static registry in frontend |
| **Users** *(admin)* | Account email | ❌ until User management UI API |

**Operator direction:**
- **Spotlight-style overlay** — dim background, centered search field, instant results below grouped by category ("Files", "Tasks", "Plugins"…)
- **Keyboard-first** — global shortcut opens palette; ↑↓ to move, Enter to open, Esc to close
- **Simple default** — one query searches **all categories**; optional pills/tabs to narrow ("Files only", "Tasks only")
- **Not replacing** full Monitoring/Tasks pages — quick jump, not a second UI for everything

**Possible scope (post-MVP `/spec`):**
- **Phase A:** Command palette UI + Files + Navigation + client-side filter on volumes/tasks/plugins lists
- **Phase B:** unified `GET /api/search/unified?q=` aggregating categories server-side (faster, consistent ranking)
- **Phase C:** Content category when full-text indexing ships
- Remove or slim sidebar `GlobalSearch` once palette is primary entry
- **Out of scope:** natural language / AI answers, searching outside lcloud, plugin-provided search sources (until plugin API extends)

**Related ideas:**
- **Full-text content search via Bleve** — powers the **Content** category inside the same palette
- **Control panel iconography pass** — category icons in result rows

---

## Cloud table-stakes gaps (vs OneDrive / Drive / Dropbox / Nextcloud)

> Competitive scan 2026-07-02 — standard file-cloud features lcloud MVP lacks. Complements operator-driven ideas above (explorer redesign, Spotlight, etc.).

### Trash / recycle bin
**Context:** OneDrive, Google Drive, Dropbox, and Nextcloud all move deleted files to a **Trash** first (recoverable for a period). lcloud **permanently deletes** on confirm (`DELETE /api/volumes/:id/files`) — no undo, no retention folder.
**Value:** Recover from accidental deletes; matches user expectations from every major cloud.
**Estimated effort:** Medium
**Dependencies:** Phase 1.1 file ops; volume explorer redesign (Trash view in UI)
**Date:** 2026-07-02

**Possible scope (post-MVP `/spec`):**
- Soft-delete to `{volume}/.trash/` or PostgreSQL tombstone + blob retention; `POST restore`, auto-purge after N days (task or setting)
- UI: Trash pseudo-folder or sidebar entry; empty trash action
- Event bus: `file.trashed`, `file.restored`

---

### In-browser file preview (lightbox)
**Context:** Drive and OneDrive open images, PDFs, and video in a **built-in viewer**; lcloud only shows a small list thumbnail or forces **download**.
**Value:** Read and inspect files without leaving the browser or saving locally.
**Estimated effort:** Medium
**Dependencies:** Phase 1.1 file content + thumbnail API; iconography pass
**Date:** 2026-07-02

**Possible scope (post-MVP `/spec`):**
- Lightbox/modal: images (full res), PDF (browser embed or pdf.js), video/audio (`<video>` / `<audio>`)
- Open from grid view, list double-click, Spotlight result
- Out of scope v1: Office editing, annotation, collaborative viewing

---

### Multi-file selection and bulk actions
**Context:** All major clouds support **checkbox selection** and batch delete, download, move. lcloud acts on **one file at a time** in the UI.
**Value:** Manage large folders efficiently; prerequisite for a credible explorer redesign.
**Estimated effort:** Medium
**Dependencies:** Volume explorer redesign; move/delete/download APIs (bulk endpoints or parallel calls)
**Date:** 2026-07-02

**Possible scope (post-MVP `/spec`):**
- Shift/Cmd-click, select all; toolbar: Delete, Download (zip if multiple), Move to folder
- Backend: optional `POST /files/bulk-delete` for atomic quota updates

---

### Multi-file upload with progress queue
**Context:** Drive/OneDrive accept **many files and folders** in one drop with per-file progress bars. lcloud UI uploads **one file** per action (`files[0]` in `VolumeDetailPage`).
**Value:** Everyday photo/document uploads without repeating the flow.
**Estimated effort:** Medium
**Dependencies:** Volume explorer redesign (drop target); existing `POST /files` per file
**Date:** 2026-07-02

**Possible scope (post-MVP `/spec`):**
- Queue UI: progress per file, cancel, retry failed; concurrent uploads (limit 3–5)
- Folder upload via `webkitdirectory` where supported
- Out of scope v1: tus/resumable multipart for huge files (separate idea if needed)

---

### Copy / duplicate file (in-volume)
**Context:** Drive « Make a copy », Explorer Ctrl+C/V — duplicate within the same folder or elsewhere. lcloud has **move** internally (`MoveFileInternal`) but no user-facing **copy** in UI or web API.
**Value:** Duplicate templates, backups ad hoc, workflows without download/re-upload.
**Estimated effort:** Quick to Medium
**Dependencies:** Volume layer copy primitive; explorer context menu
**Date:** 2026-07-02

---

### File details side panel
**Context:** Drive/OneDrive show a **Details** pane: size, dates, type, owner, checksum. lcloud exposes size/modified in table columns only; SHA256 is indexed but not shown in UI.
**Value:** Inspect metadata without downloading; trust/verify files (hash display).
**Estimated effort:** Quick
**Dependencies:** Volume explorer redesign; existing metadata cache / Bleve index
**Date:** 2026-07-02

---

### Recent files and favorites (starred)
**Context:** Google Drive homepage highlights **Recent** and **Starred**; lcloud dashboard is empty and there is no per-user file pinning or activity feed.
**Value:** Jump back to last-worked files; personal shortcuts across volumes.
**Estimated effort:** Medium
**Dependencies:** Dashboard widgets idea; PostgreSQL `user_favorites`; recent = index by `modified_at` or event log
**Date:** 2026-07-02

**Possible scope (post-MVP `/spec`):**
- `GET /api/files/recent`, `POST/DELETE /api/files/favorite`
- Widgets on `/dashboard` + optional sidebar section
- Cross-link **Unified Spotlight search** (recent as default palette state)

---

### Download folder as ZIP
**Context:** OneDrive/Dropbox/Nextcloud let users **download a folder as ZIP**. lcloud supports single-file download only.
**Value:** Export a project folder, backup a subtree, share offline without sync client.
**Estimated effort:** Medium
**Dependencies:** Phase 1.1 file read; streaming zip from Go (`archive/zip`)
**Date:** 2026-07-02

---

### In-instance volume sharing (per-user access)
**Context:** Google Drive **Share with specific people** on the same tenant. lcloud: **one owner per volume**; admin sees all, user sees own — no grant access to a colleague's volume. Distinct from public links (out of scope per VISION).
**Value:** Family/small team on one lcloud: share a « Photos » volume without duplicating data or making everyone admin.
**Estimated effort:** Heavy
**Dependencies:** User management UI; ACL model (volume_members join table); authorize all file/volume APIs
**Date:** 2026-07-02

**Product constraint:** sharing **within the instance only** (existing admin-provisioned accounts) — not anonymous public URLs.

---

### Per-file version history
**Context:** OneDrive and Drive keep **previous versions** of a file (restore older copy). lcloud **Backup and archiving** idea covers volume-level backup, not per-file revision timeline.
**Value:** Undo overwrite (« I saved the wrong version ») without restoring an entire volume.
**Estimated effort:** Heavy
**Dependencies:** Volume file ops; storage strategy (copy-on-write snapshots or `{file}.v{n}` sidecar)
**Date:** 2026-07-02

**Relationship:** complements **Trash** (accidental delete) vs **Versions** (intentional overwrite).

---

### User profile — change own password
**Context:** Every cloud has **Account → Security → change password**. lcloud: admin can reset via future Users UI; no **self-service** password change on Settings page.
**Value:** Users rotate credentials without admin intervention.
**Estimated effort:** Quick
**Dependencies:** Auth service (`ChangePassword(current, new)`); Settings page UI
**Date:** 2026-07-02

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
