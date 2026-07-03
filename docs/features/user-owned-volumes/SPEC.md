# Spec: User-owned volumes and deletion requests

> **Status:** Implemented  
> **Scope:** Let `user` role create and manage their own volumes (masked disk picker); restrict volume deletion to `admin`; phase-A deletion request queue.  
> **Sources:** [IDEAS.md — User-owned volumes](../../../IDEAS.md), [docs/project.md](../../project.md), [design/DESIGN.md](../../../design/DESIGN.md)

---

## Objective

Regular users cannot create volumes today (`RequireAdmin` on `POST /api/volumes`). Family/small-team deployments need per-user storage without granting admin. Destructive delete stays admin-only; users **request** deletion (phase A: admin queue, no bell notifications).

**User stories:**

| ID | Story |
|---|---|
| UOV1 | As a `user`, I create a volume on a masked storage (`Storage 1` + free space). |
| UOV2 | As a `user`, I see and use only my volumes. |
| UOV3 | As a `user`, I cannot delete a volume — I can request deletion. |
| UOV4 | As an `admin`, I see pending deletion requests and dismiss or delete the volume. |
| UOV5 | As an `admin`, I still create/delete volumes and see real disk paths when unmasked. |

**Out of scope v1:** per-user volume/quota limits; in-app notifications (phase B); volume sharing; email.

---

## Permissions

| Action | `user` | `admin` |
|---|---|---|
| `POST /api/volumes` | Own volumes (`owner_id` = self) | Yes |
| `DELETE /api/volumes/:id` | **403** | Yes (+ force) |
| `POST /api/volumes/:id/deletion-request` | Own volumes | N/A (can delete directly) |
| `GET /api/admin/volume-deletion-requests` | 403 | List pending |
| `DELETE /api/admin/volume-deletion-requests/:id` | 403 | Dismiss |

---

## Disk picker (masked)

- Every disk in `GET /api/disks` has stable `id`: `storage-1`, `storage-2`, … (index in registry).
- **Masked** (`mask_disk_names` + non-admin): `path` empty, `label` = `Storage N`, show `free_bytes` / `total_bytes`.
- **Create** accepts `disk_id` (required when `path` hidden) or `disk_path` (admin unmasked). Server resolves `disk_id` → real path via `DiskRegistry`.

---

## Deletion requests (phase A)

Model `volume_deletion_requests`: `id`, `volume_id`, `user_id`, `status` (`pending` | `dismissed`), timestamps.

- One pending request per volume (409 if duplicate).
- Admin dismisses via `DELETE /api/admin/volume-deletion-requests/:id`.
- Admin deletes volume → pending requests for that volume auto-dismissed.

---

## Success criteria

- [ ] SC-UOV1 User can create volume with `disk_id` on masked disk list
- [ ] SC-UOV2 User DELETE volume returns 403
- [ ] SC-UOV3 User can POST deletion-request; admin sees queue
- [ ] SC-UOV4 Volumes empty state is role-aware
- [ ] SC-UOV5 `go test ./internal/...` and `cd web && npm run build` pass
