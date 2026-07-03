# Spec: Control panel iconography pass

> **Status:** Draft — pending review  
> **Scope:** Add consistent Lucide icons across the lcloud web control panel — sidebar navigation, common actions, file/volume lists, and empty/status states — without changing behavior or APIs.  
> **Prerequisite:** MVP web UI shipped (React SPA, shadcn/ui, Tailwind). Volume explorer redesign may land before or in parallel; this spec consolidates and extends partial icon usage already present in the explorer.  
> **Sources:** [VISION.md](../../../VISION.md), [docs/project.md](../../project.md), [IDEAS.md — Control panel iconography pass](../../../IDEAS.md), [design/DESIGN.md](../../../design/DESIGN.md)

---

## Assumptions (correct me now or I proceed)

1. **Frontend-only change** — no backend, API, or `.volume.json` changes.
2. **Lucide as the single icon library** — `lucide-react` is already in `web/package.json`; no new icon dependency without ADR.
3. **Partial explorer adoption exists** — `ExplorerToolbar`, `FolderTree`, `FileListView`, and `FileGridView` already import Lucide icons ad hoc; this pass **centralizes** choices in one module and fills gaps elsewhere.
4. **Behavior unchanged** — icons are decorative/affordance only; labels, routes, and actions stay the same unless an icon-only control already has an `aria-label` (explorer view toggle pattern).
5. **Design system alignment** — colors and sizes follow Tailwind tokens mapped from [design/DESIGN.md](../../../design/DESIGN.md) (`text-muted`, `text-primary` when active, `h-4 w-4` inline, `button-icon-circular` where applicable).
6. **Rollout in vertical slices** — ship sidebar + shell first, then cross-page actions, then file/volume MIME icons, then tasks/plugins/monitoring polish.
7. **No custom SVG brand set** — Lucide defaults only; no animated icons; no emoji-as-icons (per IDEAS.md operator direction).
8. **Accessibility is non-negotiable** — decorative icons `aria-hidden="true"`; icon-only buttons keep visible text or `aria-label`; icon + label buttons do not duplicate accessible name on the icon.
9. **Unified Spotlight search is out of scope** — that feature depends on this pass for result icons but is not built here.
10. **Mobile sidebar collapse is out of scope** — icons must work at current desktop-first layout; hamburger nav is a separate responsive pass.

---

## Objective

The control panel MVP is **text-heavy**: sidebar nav items are labels only, many action buttons are plain text ("Download", "Delete", "Create volume", "Logout"), and file rows outside the explorer use generic `File` / `Folder` icons without MIME differentiation. Operator feedback: the UI feels sparse and harder to scan.

**Who:** Self-hosters who use the web UI daily (VISION.md Principle 3 — familiar, technical tooling).

**Why:** Faster navigation and clearer affordances — users spot Volumes vs Tasks vs Settings at a glance; actions read like a familiar desktop/cloud app; visual polish aligned with the ClickHouse-inspired design system.

**Pain points addressed:**

| Today | Target |
|---|---|
| Sidebar: text-only nav links | Icon + label per entry; primary tint when active |
| Actions: text-only buttons | Icon + label on primary actions; icon-only only with tooltip/`aria-label` |
| Files: generic `File` icon everywhere | MIME/category icons (image, video, document, archive, code, generic) |
| Empty states: plain text in cards | Light icon above message (folder, plugin, task, alert) |
| Explorer: ad hoc Lucide imports | Shared icon map + `getFileIcon()` helper |

**User stories:**

| ID | Story |
|---|---|
| IC1 | As a user, I see an icon beside each sidebar item and instantly recognize the section (Dashboard, Volumes, Tasks, …). |
| IC2 | As a user, common buttons (Upload, New folder, Download, Delete, Run task, Logout) show a recognizable icon next to the label. |
| IC3 | As a user, file rows and grid cards show a type-appropriate icon when no thumbnail is available. |
| IC4 | As a user, empty lists (no volumes, no tasks, no plugins) show a simple icon so the page feels intentional, not broken. |
| IC5 | As a screen-reader user, icon-only controls remain fully labeled; decorative icons are ignored by assistive tech. |

**Out of scope:**

- Custom brand icon set or illustrated empty-state artwork
- Animated or emoji icons
- New pages or navigation structure
- Icon picker / user theming
- Backend MIME detection changes (use existing `FileEntry.mime_type` + extension fallback)
- Mobile hamburger nav redesign
- Unified Spotlight search UI (consumes this icon map later)

---

## Tech Stack

See [docs/project.md — Tech Stack](../../project.md#tech-stack).

This feature adds or extends:

| Layer | Change |
|---|---|
| Frontend | `lucide-react` (already approved, `^0.511.0`) |
| Frontend | New module `web/src/lib/icons/` — icon map + helpers |
| Frontend | Optional thin wrapper `web/src/components/ui/icon.tsx` for size/color defaults |

No new npm dependencies expected.

---

## Commands

### Development

```bash
# Frontend — primary verification surface
cd web && npm run test
cd web && npm run build

# Optional: run only icon-related tests once added
cd web && npm run test -- src/lib/icons

# Full stack smoke (visual check in browser)
docker compose up -d --build
# → http://localhost:8080 (or configured port)
```

### Lint / typecheck

```bash
cd web && npm run build   # tsc + vite — catches bad Lucide imports
```

No backend test commands required (no Go changes).

---

## Project Structure

See [docs/project.md — Project Structure](../../project.md#project-structure).

Feature-specific layout:

```
web/src/
├── lib/icons/
│   ├── index.ts              → re-exports
│   ├── navIcons.ts           → sidebar route → Lucide component
│   ├── actionIcons.ts        → action key → Lucide component
│   ├── fileIcons.ts          → getFileIcon(entry) by mime + extension
│   └── fileIcons.test.ts     → MIME/extension mapping tests
├── components/
│   ├── ui/icon.tsx           → optional Icon wrapper (size, className, aria-hidden)
│   └── layout/Sidebar.tsx    → pass 1: nav icons
└── …                         → passes 2–4 touch page/component files listed below
```

Docs:

```
docs/features/control-panel-iconography/
└── SPEC.md                   → this file
```

---

## Code Style

Follow existing React + Tailwind conventions. One central map beats scattered `import { X } from "lucide-react"` per file.

**Icon map (nav):**

```tsx
// web/src/lib/icons/navIcons.ts
import {
  LayoutDashboard,
  HardDrive,
  Activity,
  Puzzle,
  ListTodo,
  Settings,
  Users,
  type LucideIcon,
} from "lucide-react";

export const navIcons: Record<string, LucideIcon> = {
  "/dashboard": LayoutDashboard,
  "/volumes": HardDrive,
  "/monitoring": Activity,
  "/plugins": Puzzle,
  "/tasks": ListTodo,
  "/settings": Settings,
  "/settings/users": Users,
};
```

**Sidebar nav item usage:**

```tsx
const Icon = navIcons[item.to];
// …
<NavLink …>
  <span className="flex items-center gap-2">
    {Icon ? (
      <Icon
        className={cn("h-4 w-4 shrink-0", isActive ? "text-primary" : "text-muted")}
        aria-hidden
      />
    ) : null}
    <span>{item.label}</span>
  </span>
  …
</NavLink>
```

**File icon helper:**

```tsx
// web/src/lib/icons/fileIcons.ts
export function getFileIcon(entry: Pick<FileEntry, "type" | "mime_type" | "name">): LucideIcon {
  if (entry.type === "directory") return Folder;
  // mime prefix checks → Image, Film, FileText, FileArchive, FileCode, File
}
```

**Conventions:**

| Concern | Rule |
|---|---|
| Default size | `h-4 w-4` inline; `h-5 w-5` in empty-state blocks |
| Active nav | Icon inherits `text-primary` with label |
| Inactive nav | Icon `text-muted`; label `text-body` |
| Icon + label button | Icon `mr-2`, `aria-hidden` on icon; button text is accessible name |
| Icon-only button | `aria-label` on `<Button>` (existing explorer list/grid pattern) |
| Destructive actions | `Trash2` icon + label in `text-accent-rose`; keep existing confirm dialogs |

**Migrate explorer ad hoc imports** to the shared map in pass 3 so `Upload`, `FolderPlus`, `List`, `Grid3X3`, `Folder`, `File` are not duplicated.

---

## Testing Strategy

See [docs/project.md — Coverage targets](../../project.md#coverage-targets).

| Layer | Focus | Location |
|---|---|---|
| Frontend unit | `getFileIcon()` MIME prefix + extension fallback | `web/src/lib/icons/fileIcons.test.ts` |
| Frontend unit | Nav/action maps expose expected keys | `web/src/lib/icons/icons.test.ts` (optional, lightweight) |
| Frontend component | Sidebar renders icons for all nav items (snapshot or role query) | `Sidebar.test.tsx` (optional pass 1) |
| Manual | Visual scan all main routes | Browser smoke checklist below |

No backend tests. No E2E requirement for v1 — manual visual pass is acceptable given low logic risk.

**Manual smoke checklist (each pass):**

1. Log in → sidebar shows icons for every visible nav item; active route highlights icon + text yellow.
2. Logout button shows icon + "Logout".
3. Volumes page → "Create volume" shows icon; volume cards show disk icon (optional).
4. Volume explorer → toolbar icons unchanged visually; list/grid use MIME icons for sample files (pdf, jpg, txt, zip).
5. Tasks → Run / Edit / Delete show icons; empty state shows task icon.
6. Plugins → empty state shows plugin icon.
7. Keyboard / screen reader → tab to icon-only view toggles; hear "List view" / "Grid view".

---

## Boundaries

- **Always:** Run `cd web && npm run build` before PR; keep icons in `web/src/lib/icons/`; use `aria-hidden` on decorative icons; match [design/DESIGN.md](../../../design/DESIGN.md) tokens.
- **Ask first:** Adding any icon library other than Lucide; changing button labels or removing text labels; icon-only primary CTAs without tooltip.
- **Never:** Change API routes or auth; introduce custom SVG assets in `public/` for this pass; use emoji as icons; bypass the central icon map with one-off Lucide imports in page files (explorer migration excepted during transition).

---

## Icon inventory (proposed v1 map)

### Navigation (`navIcons`)

| Route | Label | Lucide icon |
|---|---|---|
| `/dashboard` | Dashboard | `LayoutDashboard` |
| `/volumes` | Volumes | `HardDrive` |
| `/monitoring` | Monitoring | `Activity` |
| `/plugins` | Plugins | `Puzzle` |
| `/tasks` | Tasks | `ListTodo` |
| `/settings` | Settings | `Settings` |
| `/settings/users` | Users | `Users` |

### Actions (`actionIcons`)

| Key | Used on | Lucide icon | Pattern |
|---|---|---|---|
| `upload` | Explorer toolbar | `Upload` | icon + label |
| `newFolder` | Explorer toolbar | `FolderPlus` | icon + label |
| `download` | Context menu, file actions | `Download` | icon + label |
| `delete` | Volume card, task list, context menu | `Trash2` | icon + label |
| `rename` | Volume card | `Pencil` | icon + label |
| `refresh` | Future / monitoring | `RefreshCw` | icon + label |
| `run` | Task list | `Play` | icon + label |
| `edit` | Task list | `Pencil` | icon + label |
| `create` | Create volume, create task | `Plus` | icon + label |
| `logout` | Sidebar footer | `LogOut` | icon + label |
| `open` | Volume card | `FolderOpen` | icon + label |
| `listView` | Explorer | `List` | icon-only + `aria-label` |
| `gridView` | Explorer | `Grid3X3` | icon-only + `aria-label` |
| `columns` | List column picker | `Columns3` | icon-only + `aria-label` |

### File categories (`getFileIcon`)

| Condition | Lucide icon |
|---|---|
| `type === "directory"` | `Folder` |
| `mime_type` starts with `image/` | `FileImage` |
| `mime_type` starts with `video/` | `FileVideo` |
| `mime_type` starts with `audio/` | `FileAudio` |
| `mime_type` starts with `text/` or common doc MIME | `FileText` |
| pdf | `FileType` |
| zip, gzip, x-tar, … | `FileArchive` |
| json, javascript, xml, … | `FileCode` |
| default | `File` |

Extension fallback when `mime_type` is missing (match existing opener logic in `web/src/lib/fileOpeners/`).

### Empty / status states

| Context | Lucide icon |
|---|---|
| No volumes | `HardDrive` |
| No tasks | `ListTodo` |
| No plugins | `Puzzle` |
| Empty folder (explorer) | `FolderOpen` |
| Loading (optional) | `Loader2` with `animate-spin` — **only if already used elsewhere; otherwise skip** |
| Error (generic) | `AlertCircle` |

---

## Implementation plan (vertical slices)

Dependency order — do not skip pass 1 before pass 2.

### Pass 1 — Shell & navigation (IC1, IC5)

**Files:** `web/src/lib/icons/*`, `Sidebar.tsx`, optional `icon.tsx`

- Create icon map modules.
- Add nav icons to `Sidebar` with active/inactive colors.
- Add logout icon + label.
- Verify `GlobalSearch` unchanged (search icon deferred — not in IDEAS operator list for v1).

**Checkpoint:** Sidebar visual review; build passes.

### Pass 2 — Page actions & empty states (IC2, IC4)

**Files:** `VolumesPage.tsx`, `VolumeCard.tsx`, `TasksPage.tsx`, `TaskList.tsx`, `PluginsPage.tsx`, `PluginList.tsx`, `DashboardPage.tsx` (minimal), `SettingsPage.tsx` if actions exist

- Icon + label on header CTAs ("Create volume", "Create task").
- Icon + label on card/list actions (Open, Rename, Delete, Run, Edit).
- Empty-state icons in cards (no volumes / tasks / plugins).

**Checkpoint:** Manual smoke items 2–3, 5–6.

### Pass 3 — File & volume browser consolidation (IC3)

**Files:** `ExplorerToolbar.tsx`, `FileListView.tsx`, `FileGridView.tsx`, `FolderTree.tsx`, `ListColumnHeader.tsx`, `context-menu.tsx` usages in explorer, `fileIcons.ts`

- Replace generic `File` with `getFileIcon(entry)`.
- Refactor explorer components to import from `@/lib/icons` instead of direct Lucide.
- Add icons to context menu items (Open, Download, Rename, Delete) where menu exists.

**Checkpoint:** Manual smoke item 4; `fileIcons.test.ts` green.

### Pass 4 — Monitoring & polish

**Files:** `DiskOverviewCard.tsx`, `UnifiedVolumeList.tsx`, `MimeBreakdownBars.tsx`, `UploadQueue.tsx`, `BreadcrumbNav.tsx` (optional chevron already present)

- Optional `HardDrive` on disk cards.
- MIME breakdown category icons if low effort.
- Upload queue status icons (success / error / in progress) — optional; skip if scope creeps.

**Checkpoint:** Full manual smoke checklist; no regressions in explorer DnD or rename.

---

## Success Criteria

Iconography pass is **done** when all of the following pass:

- [ ] **SC-IC1** Every sidebar nav item shows a Lucide icon with label; active item uses `text-primary` on icon + text.
- [ ] **SC-IC2** Logout button shows icon + "Logout".
- [ ] **SC-IC3** Central module `web/src/lib/icons/` exists; explorer toolbar/list/grid import shared maps (no duplicate icon choice strings across files).
- [ ] **SC-IC4** `getFileIcon()` returns distinct icons for directory, image, video, document/pdf, archive, and generic file — covered by unit tests.
- [ ] **SC-IC5** Primary page actions on Volumes, Tasks, and Plugins pages use icon + label pattern from action map.
- [ ] **SC-IC6** Empty states for volumes, tasks, and plugins show a centered icon above the message.
- [ ] **SC-IC7** All icon-only controls have `aria-label`; decorative icons have `aria-hidden`.
- [ ] **SC-IC8** `cd web && npm run test && npm run build` pass with no new lint/type errors.
- [ ] **SC-IC9** No backend files changed; no new npm dependencies.

---

## Open Questions

| ID | Question | Default if no answer |
|---|---|---|
| OQ1 | Should **GlobalSearch** input get a search icon (`Search`) inside the field? | **No** in v1 — out of operator list; add in Spotlight spec |
| OQ2 | Volume cards: add a leading **HardDrive** icon next to volume name? | **Yes** — low effort, helps scan |
| OQ3 | Context menu items: icon left of every item, or only destructive (Delete)? | **All items** — consistent with cloud drive UX |
| OQ4 | Destructive actions: red-tinted `Trash2` icon or neutral muted icon? | **Resolved:** red icon + label (`text-accent-rose`) |
| OQ5 | Pass 4 monitoring icons in scope for v1 or defer? | **Resolved:** in v1 scope |
| OQ6 | Add `Sidebar.test.tsx` or rely on manual + `fileIcons.test.ts` only? | **`fileIcons.test.ts` required**; Sidebar test optional |

---

## Relationship to other work

| Related | Relationship |
|---|---|
| [IDEAS.md — Unified Spotlight search](../../../IDEAS.md) | Depends on this pass for category/result row icons |
| [SPEC.md — Volume file explorer](../../../SPEC.md) | Explorer SC references MIME icons "until iconography pass lands" — pass 3 fulfills that |
| [IDEAS.md — Dashboard widgets](../../../IDEAS.md) | Future widgets should import the same `@/lib/icons` maps |
| [design/DESIGN.md](../../../design/DESIGN.md) | `button-icon-circular` (36×36) applies to icon-only controls; yellow primary unchanged |

---

## Revision history

| Date | Change |
|---|---|
| 2026-07-03 | Initial draft from IDEAS.md operator direction |
