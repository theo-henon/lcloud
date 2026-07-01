# lcloud • Vision

> Language: English
> Last updated: 2026-07-01
>
> This document evolves with the project. It may start lean and grow richer
> over time. Sections marked "to complete later" are invitations, not obligations.

---

## 1. Reason for being

lcloud exists because data ownership shouldn't require compromise.

Today, most people store their data on services they don't control — Google Drive, iCloud, Dropbox. Convenient, yes. But the data lives on someone else's hardware, subject to someone else's rules, priced by someone else, and one policy change away from being unavailable. The people who understand this and want out have options, but those options often impose their own constraints: heavy to install, opinionated in how they work, closed to extension.

lcloud is built for the people who know exactly what they want. Your data lives where you decide — on your hardware, in your network, under your rules. When the built-in tools aren't enough, you extend lcloud exactly the way you want. No workarounds. No waiting for a maintainer to merge your feature. You build it, you install it, it works.

The plugin system is not a feature bolt-on. It is the foundation. Every internal capability of lcloud is exposed as events in a shared bus. External plugins — yours, the community's — subscribe to those events, respond to them, and emit new ones. This is what makes lcloud genuinely extensible: not a settings page with toggles, but a programmable infrastructure you can shape to your exact needs.

---

## 2. Guiding principles

### Principle 1 • Ownership first
Your data never leaves your infrastructure unless you decide it does. lcloud runs on your hardware. Volumes live on disks you choose. Metadata stays inside the volume — not in a third-party cloud, not in a SaaS database. If you shut lcloud down tomorrow, every byte is still exactly where you left it, in a structure you can read without lcloud.

### Principle 2 • Extensibility by design
The plugin system is not a feature — it's the foundation everything is built on. The event bus is the nervous system of lcloud. Every operation — file upload, volume creation, task execution — emits events. Plugins subscribe to those events and can emit new ones. lcloud's behavior is shapeable by anyone with the knowledge and motivation to do it.

### Principle 3 • Built for people who know what they're running
lcloud is honest and technical. It does not abstract away disk selection, volume structure, or plugin behavior to protect non-technical users from complexity. The target user understands that `./userdata` is a folder on a real disk, that a plugin is a subprocess with a defined interface, and that a task is a cron job with access to volume primitives. lcloud speaks plainly to that user.

### Principle 4 • Community-first
lcloud is open-source with no commercial intent. Plugins are first-class citizens — writing a plugin is writing lcloud. The community plugin registry is as important as the core. Contributions of all kinds — plugins, bug fixes, documentation, translations — are the product's lifeblood.

### Principle 5 • One command to start
`docker compose up`. That is the full installation process on a fresh machine. No wizard, no guided setup tour, no 30-step tutorial. A single command, a running instance, a working admin interface. Self-hosting should be accessible to anyone technical enough to run a Docker container.

---

## 3. Long-term vision — the rise

**Stage 1 • Personal cloud** — lcloud runs at home. One user, or a family on the same local network. Volumes on local disks. A first plugin. Tasks that keep things tidy. A dashboard that shows what's happening. The core is working and trustworthy.

**Stage 2 • Extensible platform** — The plugin ecosystem grows. Community plugins handle media, documents, backups, notifications, and integrations. WebDAV and FTP expose volumes to native OS clients. A plugin registry makes discovery natural. lcloud becomes less a single tool and more a platform that different people configure into very different things.

**Stage 3 • Infrastructure primitive** — lcloud is the storage layer under larger systems. AI tools read and write to volumes. VS Code integrates directly. Organizations run it as the backbone of their internal file infrastructure. The plugin API is stable enough that third parties build on it without asking permission.

Each stage makes the next one possible. The plugin ecosystem built in Stage 2 is what attracts the integrations and production deployments of Stage 3.

---

## 4. What lcloud is not

**Not a consumer cloud.** lcloud is not designed to be a Dropbox replacement for non-technical users. There is no hosted version, no free tier, no mobile app with swipe gestures. The target user opens a terminal.

**Not a managed service.** There is no lcloud cloud. There is no company that runs lcloud for you. lcloud runs on your hardware, and that is both a constraint and the entire point. Hosting it yourself is not a limitation — it is the feature.

**Not an enterprise product.** lcloud has no licensing model, no enterprise tier, no vendor relationship. It is open-source infrastructure. Organizations that adopt it do so because it works for them technically, not because a salesperson showed them a roadmap.

---

## 5. Conceptual architecture

**Volume** — the fundamental storage primitive. A directory on a physical disk, self-describing (`.volume.json`), portable (readable without lcloud), structured (`userdata/`, `cache/`, `plugins/`). A volume is the atom — everything else operates on volumes.

**Event Bus** — the nervous system. Every meaningful action in lcloud emits a typed event. File uploaded. Volume created. Task executed. Plugin registered. The bus is what makes extensibility real: plugins don't hook into lcloud's internals — they listen to the bus and respond.

**Plugin Runtime** — the extension layer. A plugin is an external binary that implements a defined Go interface and communicates with lcloud via stdin/stdout (go-plugin protocol). Any language that compiles to a binary can write a plugin. The runtime manages lifecycle, restarts, and isolation.

**Task Engine** — the automation layer. Tasks are scheduled operations (periodic or event-triggered) that execute predefined macros against volumes. The macro vocabulary is the public API between automation and the volume primitive.

**VolumeIndexer** — the search layer. An interface that abstracts the indexing engine. The MVP implementation uses Bleve, indexing file metadata. The same interface will support full-text content indexing without changing callers.

---

## Note on the evolution of this document

VISION.md is a living document. It grows richer as the project matures.
The operational detail does not live here. See:
- **STARTUP.md** for the bootstrap plan to MVP (temporary document)
- **docs/project.md** for stack and technical architecture
- **IDEAS.md** for future ideas not yet planned
