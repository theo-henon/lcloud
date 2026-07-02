# AGENTS.md — lcloud

> Context file for AI coding agents (Cursor, Claude Code, Codex, OpenCode,
> Copilot, etc.). This project uses **agent-skills**
> (https://github.com/addyosmani/agent-skills) installed globally. This file
> provides project-specific context only — it does not duplicate skill content.

---

## Project

lcloud is a self-hosted, open-source, extensible cloud storage platform. Users deploy it on their local network or a VPS via Docker Compose, store files on their own hardware in structured Volumes, and extend every behavior through a first-class plugin system (VS Code event-bus model). Target audience: experienced users (homelab enthusiasts, developers, sysadmins) who want total control over their data.

**Current phase:** Phase 0 — Technical foundations
**Primary language:** English
**Users:** Experienced self-hosting enthusiasts — solo or small groups (family, small team) on a local network or VPS

---

## Operators

Single operator. At session start, greet them by their `git config user.name` — a light recognition signal, nothing more.

- **Théo** — builder + architect. Stack-agnostic (choices made at project start for merit). Wants to understand architectural trade-offs on non-trivial decisions. Surface options and reasoning; don't hide what matters.

---

## Foundation files

Read these for project context (in this order):

1. **VISION.md** — Strategic vision AND **decision filter** — read before any `/spec` or architecture decision. Every new feature should align with the vision; if alignment isn't clear, stop and clarify before specifying.
2. **docs/project.md** — Tech Stack, architecture, deployment workflow, constraints. **Source of truth for technical matters.**
3. **README.md** — Current state, local setup, deployment workflow summary.
4. **IDEAS.md** — Parking lot of future ideas (consult only when relevant).
5. **design/DESIGN.md** *(present)* — ClickHouse design system (getdesign.md convention). Dark canvas (`#0a0a0a`), electric yellow (`#faff69`), Inter + JetBrains Mono. **Always read before generating any UI.**

---

## Project structure

```
lcloud/
├── README.md                   Current state, setup
├── VISION.md                   Strategic vision
├── IDEAS.md                    Parking lot for future ideas (English)
├── AGENTS.md                   This file
├── CLAUDE.md                   One-liner: @AGENTS.md (Claude Code compatibility)
├── cmd/
│   └── server/main.go          Go entry point
├── internal/                   Go domain modules
│   ├── auth/                   JWT, user management, admin/user roles
│   ├── volume/                 Volume CRUD, file ops, .volume.json source of truth
│   ├── monitoring/             Stats, file type breakdown, multi-disk aggregation
│   ├── indexer/                VolumeIndexer interface + Bleve implementation
│   ├── plugin/                 go-plugin runtime, event bus, lifecycle
│   ├── task/                   gocron, macro executor, task CRUD
│   └── api/                    Gin routes and handlers (thin — call services only)
├── pkg/                        Exported shared utilities
├── web/                        React + Vite frontend
│   └── src/
│       ├── components/
│       ├── pages/
│       ├── hooks/
│       ├── lib/
│       └── store/
├── plugins/                    Plugin binaries (git-ignored, mounted in Docker)
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── .env.example
├── docs/
│   ├── project.md              Stack + architecture (source of truth)
│   ├── adr/                    Architecture Decision Records
│   └── features/               Feature-level docs (filled when needed)
└── design/
    ├── DESIGN.md               ClickHouse design system (getdesign.md)
    └── screenshots/
```

---

## Project-specific conventions

### Source of truth for Tech Stack and Project Structure
- **`docs/project.md` is authoritative.**
- When running `/spec` or any skill that requires Tech Stack / Project Structure information, **reference `docs/project.md` instead of duplicating** the content into SPEC.md.
- Example in SPEC.md: `## Tech Stack` → `See [docs/project.md](./docs/project.md#tech-stack).`

### Volume structure — immutable contract
The physical volume directory structure is a public contract. Any change to it (adding/removing directories, changing `.volume.json` schema) is an **architectural decision** — it must be documented as an ADR in `docs/adr/` and must not break portability (ability to re-import a volume by pointing to its root directory).

`.volume.json` is the source of truth for volume config. PostgreSQL is a performance cache. **Never write volume config exclusively to PostgreSQL.**

### VolumeIndexer interface — never bypass
All file search and indexing must go through the `VolumeIndexer` interface (`internal/indexer/`). Never call Bleve directly from outside the indexer module. This interface is what allows upgrading to content search without changing callers.

### Plugin event bus — subscribe, never call directly
No module outside `internal/plugin/` should call plugin functions directly. All plugin interactions go through the event bus. Plugins are isolated subprocesses — treat them as external systems.

### Macro vocabulary — predefined list is intentional
The predefined macro list in Phase 1.4 is closed for the MVP. Do not add macros outside the list without a `/spec` cycle. Macros are architecture surface — they extend the task system's public API.

### Deployment workflow (immutable)
- Short-lived branches (`feature/*`, `fix/*`, `docs/*`, `chore/*`, `refactor/*`) → PR → `main`
- Never commit directly to `main` unless explicitly requested
- Full details in `docs/project.md#deployment-workflow`

### Git & PR workflow

This repo uses agent-skills for workflow routing. For Git shipping, apply `/ship` or `git-workflow-and-versioning` with these local conventions.

Use branch names aligned with agent-skills:
- `feature/<short-topic>` for new user-facing features or project phases
- `fix/<short-topic>` for bug fixes
- `docs/<short-topic>` for documentation-only changes
- `chore/<short-topic>` for maintenance/config changes
- `refactor/<short-topic>` for behavior-preserving code simplification

Before editing files:
- if on `main`, create the appropriate short-lived branch;
- if already on a task branch, keep using it;
- if there are unrelated local changes, stop and ask.

When the user says `ship en PR` or `crée une PR`:
1. Review `git status` and the diff.
2. Stage only intended files.
3. Commit with a concise message.
4. Push the branch.
5. Open a draft PR to `main`.
6. Report the PR link and checks run.
7. Do not merge.

When the user says `merge la PR`:
1. Review the PR diff.
2. Check mergeability and conflicts.
3. Merge only if there is no blocker.
4. Fetch and update `main`.

### Post-MVP development
- MVP reached — bootstrap plan removed
- New work is **feature-scoped**; `/spec` takes a feature description directly
- Future ideas → IDEAS.md; do not revive a monolithic bootstrap roadmap file

### Documentation conventions
- **Architecture Decision Records** in `docs/adr/` — use `documentation-and-adrs` for significant technical decisions
- **Feature-level documentation** in `docs/features/<feature-name>/README.md` when a feature grows complex

### Design and visual references
- **`design/DESIGN.md`** is present — ClickHouse design system (getdesign.md). **Always read before generating any UI.** Source of truth for colors, typography, spacing, and components.
- No screenshots or mockups provided at project start — all visual direction comes from `design/DESIGN.md`.

---

## Task routing — suggesting the right entry point

This project expects agent-skills for all development work.

**→ For any non-trivial development task, start from `using-agent-skills`, then apply the project-specific conventions above. If routing is still unclear, `using-agent-skills` remains the source of truth.**

### How to route

When the user formulates a development request, suggest the appropriate entry point in 2 lines max, announce it as an assumption the user can override, then proceed.

**Routing table:**

| User's intent signal | Suggested entry point |
|---|---|
| User doesn't yet know what they want, intent unclear | `interview-me` |
| Vague idea, needs refinement before anything | `idea-refine` |
| Typo, single-line fix, unambiguous self-contained change | `/build` (skipping `/spec` + `/plan`) |
| Modify existing behavior, targeted improvement, clear scope | `/build` (skipping `/spec` + `/plan`) |
| New screen, new data model, new integration, ambiguous requirements | `/spec` (full pipeline) |
| Structural refactor (extract module, change architecture) | `/spec` |
| Cleanup refactor (same code, cleaner) | `/code-simplify` |
| Bug, error, unexpected behavior | `debugging-and-error-recovery` |
| Browser-based behavior to verify (UI flows, console, network) | `browser-testing-with-devtools` |
| Code review, verification pass | `/review` |
| Security review (input handling, auth, data storage) | `security-and-hardening` |
| Performance issue (latency, memory, throughput) | `performance-optimization` |
| Document a decision, write an ADR | `documentation-and-adrs` |
| Pre-ship web quality audit (perf, best practices) | `web-quality-audit` |
| Deploy, ship to production | `/ship` |

**Context-specific skills activate automatically.** `context-engineering` activates when starting a session, switching tasks, or when output quality drops. **UI work** activates `frontend-ui-engineering` automatically during `/build`. **API work** activates `api-and-interface-design`. **Unfamiliar library or API** activates `source-driven-development`. **High-stakes or unfamiliar-code changes** activate `doubt-driven-development`.

**Web quality skills augment work on this web project.** `performance` activates when work touches loading speed or resource optimization. `best-practices` activates when work touches security headers or HTTP-exposed surfaces. These are background skills that augment `/build` contextually rather than serving as entry points. The user-triggered `web-quality-audit` remains the explicit pre-ship audit.

**For UI work specifically:** always read `design/DESIGN.md` before generating any UI. It is the source of truth for visual language — colors, typography, components, spacing.

**Git and CI work activate dedicated skills.** `git-workflow-and-versioning` covers commits, branches, conflicts. `ci-cd-and-automation` activates when touching build/test/deploy pipelines.

**Suggestion format** (to the user, in their language):

> "Je pars sur `[entry-point]` — [reason in 1 line]. Dis-moi si tu préfères autre chose."

Then proceed. Do not block for explicit confirmation.

### When skipping `/spec` and `/plan`

Allowed only for single-line fixes, typos, or unambiguous self-contained changes. Quick check: is this really unambiguous, or a bug in disguise? When in doubt, per `using-agent-skills` rule #1: *"Check for an applicable skill before starting work."*

### When `/spec` is required (non-negotiable)

Always suggest `/spec` when the request implies:
- A new data model
- A new screen or module
- A third-party integration
- Requirements that are ambiguous or not self-contained
- A structural refactor that changes architecture
- Any change to the volume directory structure or `.volume.json` schema

### Before proposing `/spec` — a quick step-back

When a conversation has been exploratory and feels ready for `/spec`, propose a short step-back first:

> "Before we jump into `/spec`, want me to take a quick step back and check if what we've discussed is coherent, simple, and cleanly thought through?"

If the user accepts → examine through the lens of simplicity, robustness, elegance.
If the user declines → go directly to `/spec`.

### Post-MVP workflow

New work is feature-scoped. `/spec` takes a feature description directly; the agent reads VISION.md + docs/project.md (+ IDEAS.md when relevant).

---

## React skills (outside agent-skills)

These come from Vercel and are installed separately — they are NOT part of the agent-skills pipeline. Apply them in the background whenever writing, reviewing, or refactoring React code:

- **`react-best-practices`** — data fetching & waterfalls, bundle size, re-render optimization.
- **`composition-patterns`** — compound components, render props, context providers, avoiding boolean-prop bloat.

---

## What agents should NOT do

- Never modify `VISION.md` strategic structure (sections 1-5) without explicit user confirmation.
- Never duplicate Tech Stack or Project Structure content — always reference `docs/project.md`.
- Never commit secrets or real `.env` files.
- Never bypass the `VolumeIndexer` interface — all indexing through `internal/indexer/`.
- Never write volume config exclusively to PostgreSQL — `.volume.json` is the source of truth.
- Never change the volume directory structure or `.volume.json` schema without an ADR.
- Never call plugin functions directly from outside `internal/plugin/` — all interactions through the event bus.
- Never add macros to the predefined list without a `/spec` cycle.
- Never make an architectural decision without documenting it in `docs/adr/` via `documentation-and-adrs`.
- Never invoke the `accessibility` skill unless VISION.md explicitly states an accessibility compliance requirement (WCAG, EAA, etc.).
- Never put operator names in product/technical docs (VISION.md, IDEAS.md, docs/project.md). Operator names live only in the AGENTS.md "Operators" block.
