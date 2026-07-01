# lcloud

A self-hosted, open-source, and extensible cloud storage platform. Deploy it on your local network or a VPS with a single Docker command, store your files on your own hardware, and extend every behavior through a first-class plugin system.

---

## Current state

**Current phase:** Phase 0 — Technical foundations
**Target MVP:** End of Phase 1.4

> Full startup plan: [STARTUP.md](./STARTUP.md)
> Strategic vision: [VISION.md](./VISION.md)
> Stack and architecture: [docs/project.md](./docs/project.md)
> Parked ideas: [IDEAS.md](./IDEAS.md)

---

## Local setup

### Prerequisites
- Go 1.25+
- Docker & Docker Compose
- Node.js 20+ (for frontend development)

### Installation
```bash
git clone https://github.com/<username>/lcloud
cd lcloud
cp .env.example .env
```

### Environment variables
Copy `.env.example` to `.env` and fill in:
```
POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_DB=
JWT_SECRET=
ADMIN_EMAIL=
ADMIN_PASSWORD=
STORAGE_BASE_PATH=
```

### Launch
```bash
docker compose up -d
```

App accessible on http://localhost:8080.

For frontend development:
```bash
cd web && npm install && npm run dev
```

---

## Deployment

**Default workflow:** short-lived branches (`feature/*`, `fix/*`, `docs/*`, `chore/*`, `refactor/*`) → PR → `main`.

For full deployment details: see [docs/project.md](./docs/project.md#deployment-workflow).

---

## Tests

```bash
# Backend
go test ./...

# Frontend
cd web && npm run test
```

---

## Project structure

```
lcloud/
├── cmd/server/             ← Go entry point
├── internal/               ← Go modules (auth, volume, monitoring, indexer, plugin, task, api)
├── web/                    ← React + Vite frontend
├── plugins/                ← Plugin binaries (git-ignored, mounted in Docker)
├── docker-compose.yml
├── Dockerfile
├── .env.example
├── README.md
├── VISION.md
├── STARTUP.md              ← temporary, deleted after MVP
├── IDEAS.md
├── AGENTS.md
├── CLAUDE.md
├── docs/
│   ├── project.md
│   ├── adr/
│   └── features/
└── design/
    └── DESIGN.md
```

---

## Contribute

This project uses **agent-skills** (https://github.com/addyosmani/agent-skills) installed globally. See [AGENTS.md](./AGENTS.md) for project conventions and agent pointers.
