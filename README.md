# lcloud

A self-hosted, open-source, and extensible cloud storage platform. Deploy it on your local network or a VPS with a single Docker command, store your files on your own hardware, and extend every behavior through a first-class plugin system.

---

## Current state

**Current phase:** Phase 1.1 — Volume management
**Target MVP:** End of Phase 1.4

> Full startup plan: [STARTUP.md](./STARTUP.md)
> Strategic vision: [VISION.md](./VISION.md)
> Stack and architecture: [docs/project.md](./docs/project.md)
> Parked ideas: [IDEAS.md](./IDEAS.md)

Phase 1.1 delivers volume CRUD, multi-disk selection, file upload/download, extension filters, quotas, thumbnails, and Bleve metadata indexing.

---

## Local setup

### Prerequisites
- Go 1.25+
- Docker & Docker Compose
- Node.js 20+ (for frontend development)

### Installation
```bash
git clone https://github.com/theo-henon/lcloud
cd lcloud
cp .env.example .env
mkdir -p ./data/disks/ssd ./data/disks/hdd1 ./data/disks/hdd2
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
STORAGE_BASE_PATH=./data/storage
STORAGE_DISK_PATHS=/data/disks/ssd,/data/disks/hdd1,/data/disks/hdd2
STORAGE_DISK_LABELS=SSD,HDD 1,HDD 2
MAX_UPLOAD_BYTES=104857600
```

### Multi-disk setup

Each physical disk is mounted independently into the Docker container. Example from `.env.example`:

```
DISK_SSD=/mnt/ssd          # host path
DISK_HDD1=/mnt/hdd1
DISK_HDD2=/mnt/hdd2
STORAGE_DISK_PATHS=/data/disks/ssd,/data/disks/hdd1,/data/disks/hdd2
```

`docker-compose.yml` maps host paths to container paths under `/data/disks/*`. At volume creation, pick one registered disk from the UI.

If `STORAGE_DISK_PATHS` is empty, lcloud falls back to scanning subdirectories of `STORAGE_BASE_PATH` (local dev convenience).

### Launch
```bash
make embed          # build frontend and embed into Go binary
docker compose up -d --build
```

Or without Make:
```bash
cd web && npm install && npm run build
cp -R web/dist/. cmd/server/static/
docker compose up -d --build
```

App accessible on http://localhost:8080.

For frontend development:
```bash
cd web && npm install && npm run dev
```

To run the Go server locally with the embedded UI:
```bash
make embed
go run ./cmd/server
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
