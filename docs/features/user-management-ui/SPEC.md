# Spec: User management UI (admin)

> **Status:** Draft — pending review  
> **Scope:** Replace the `/settings/users` placeholder with a full admin user-management experience — list accounts, create users, change roles, disable access, and reset passwords — backed by new admin API endpoints.  
> **Prerequisite:** Phase 0 auth shipped (`POST /api/admin/users`, JWT, roles, admin seed). MVP complete.  
> **Sources:** [VISION.md](../../../VISION.md), [docs/project.md](../../project.md), [IDEAS.md — User management UI (admin)](../../../IDEAS.md), [design/DESIGN.md](../../../design/DESIGN.md)

---

## Assumptions (correct me now or I proceed)

1. **Admin-provisioned accounts only** — no public signup, no invitation links, no OAuth/SSO (IDEAS.md product constraint).
2. **Login page stays sign-in only** — account creation lives exclusively on `/settings/users`.
3. **Soft-disable, not hard delete** — v1 revokes access via `disabled_at` timestamp; no `DELETE /api/admin/users/:id`. Hard delete deferred (orphaned volumes, audit trail).
4. **Disable revokes sessions** — on disable, all refresh tokens for that user are revoked immediately; existing access tokens expire naturally (JWT stateless).
5. **Last-admin safeguard** — cannot demote or disable the sole remaining `admin` account.
6. **Self-actions allowed with guardrails** — an admin may demote or disable themselves only if another active admin exists; UI warns before self-disable/demote.
7. **Password reset is admin-set** — admin types a new password in the UI; no email pipeline.
8. **No per-volume ACL** — volume ownership stays as today (separate IDEAS.md idea); this spec does not add sharing.
9. **Self-service password change is out of scope** — covered by IDEAS.md *User profile — change own password*.
10. **GORM AutoMigrate** — add `disabled_at` column via model change + AutoMigrate in dev; production SQL note in implementation PR.
11. **Unified Spotlight search** — user search in global search remains deferred until this API ships (`IDEAS.md` dependency).
12. **First admin is env-seeded, not UI-created** — on first startup with an empty `users` table, lcloud creates one `admin` from `ADMIN_EMAIL` / `ADMIN_PASSWORD` (already shipped). The Users UI manages accounts *after* that bootstrap only.
13. **Additional admins are UI-managed** — any logged-in `admin` may create a new account with `role: admin` or promote an existing `user` to `admin` via the role-change action. Env vars do not gate who can become admin afterward.
14. **Email change is intentionally deferred** — v1 treats email as read-only identity; a follow-up feature will allow admins to change a user's email (with uniqueness check). Documented in *Future scope* below — not a permanent product decision.

---

## Account & admin model

This section answers how accounts enter the system and who can be admin. It is **product policy**, not implementation detail.

### How accounts are created

| Mécanisme | Qui | Quand |
|---|---|---|
| Variables d'environnement `ADMIN_EMAIL` + `ADMIN_PASSWORD` | Premier compte admin uniquement | Au démarrage du serveur, **seulement si la table `users` est vide** |
| Interface `/settings/users` (ou `POST /api/admin/users`) | Un admin déjà connecté | Tous les comptes suivants — `user` ou `admin` au choix |
| Page de connexion / inscription publique | — | **N'existe pas** et ne sera pas ajoutée |

En résumé : **seul un admin crée des comptes**. Pas d'auto-inscription, pas de lien d'invitation, pas d'OAuth.

### Qui est admin ?

1. **Premier admin** — défini une fois dans `.env` (`ADMIN_EMAIL`, `ADMIN_PASSWORD`). C'est le compte avec lequel tu te connectes après `docker compose up` sur une instance neuve. Ce mécanisme ne se réexécute pas si des utilisateurs existent déjà.

2. **Admins supplémentaires** — **oui, un admin peut en nommer d'autres** :
   - à la création : formulaire « Créer un utilisateur » avec rôle `admin` ;
   - plus tard : action « Changer le rôle » sur un compte `user` existant → `admin`.

   Il n'y a pas de liste blanche d'emails admin dans l'environnement après le bootstrap. La confiance est portée par le rôle en base de données.

3. **Garde-fou** — on ne peut pas rétrograder ou désactiver le **dernier** admin actif. Il faut toujours au moins un admin pour gérer l'instance.

### Email : lecture seule en v1, modifiable plus tard

En v1, l'email **n'est pas modifiable** via l'interface ni via `PATCH`. C'est un choix de scope (moins de risque de collision, sessions JWT inchangées), pas une règle définitive.

**Prévu pour une itération ultérieure** (hors cette spec) :
- action admin « Modifier l'email » ;
- vérification d'unicité (`409` si déjà pris) ;
- éventuellement déconnexion forcée de l'utilisateur concerné après changement.

Cette évolution fera l'objet d'une spec ou d'un ajout ciblé à IDEAS.md quand on la priorisera.

### Bootstrap admin : peut-on le supprimer ?

**Scénario typique :** déploiement initial → connexion avec l'admin `.env` → création d'un second admin (ex. compte « équipe ») → désactivation du premier admin bootstrap.

**Verdict : oui, ce scénario a du sens** — et c'est le comportement **voulu** pour lcloud.

| Option | Verdict | Pourquoi |
|---|---|---|
| **A. Bootstrap = amorçage unique** *(recommandé, comportement actuel)* | ✅ Oui | Après le premier démarrage, tous les admins sont **égaux** en base. L'admin `.env` n'est pas spécial : on peut le désactiver comme n'importe quel autre admin, tant qu'il en reste au moins un actif. |
| **B. Recréer l'admin `.env` à chaque redémarrage** | ❌ Très mauvaise idée | Porte dérobée permanente : le mot de passe `.env` reste valide pour toujours, même après « suppression » en UI. Impossible de révoquer l'accès bootstrap. Surprend l'opérateur (« pourquoi ce compte revient ? »). |
| **C. Admin `.env` indestructible** (jamais désactivable) | ❌ Mauvaise idée | Bloque la passation propre de l'instance. Compte « sacré » sans justification une fois un second admin en place. Casse le cas « j'ai mis le mauvais email dans `.env` ». |
| **D. Synchroniser `.env` → base à chaque boot** (mot de passe/email) | ❌ Mauvaise idée | `.env` devient source de vérité permanente — contredit le modèle « PostgreSQL = état runtime ». Modifier `.env` réinitialiserait un mot de passe sans action admin explicite. |

**Comportement actuel du code** (`SeedAdmin`) :

```
Démarrage → users vide ? → créer admin depuis .env
         → users non vide ? → ne rien faire
```

Donc :
- Désactiver le bootstrap admin **fonctionne** dès qu'un autre admin est actif.
- Changer `ADMIN_EMAIL` / `ADMIN_PASSWORD` dans `.env` **plus tard** n'a **aucun effet** (sauf reset manuel de la base).
- Redémarrer lcloud **ne recrée pas** l'admin bootstrap.

**Risque à connaître (accepté) :** si tu désactives le bootstrap admin **et** que tu perds l'accès au seul admin restant (mot de passe oublié), il n'y a pas de bouton magique dans l'UI. Récupération = accès opérateur (PostgreSQL, ou procédure documentée de reset). C'est cohérent avec une cible « utilisateurs expérimentés » (VISION.md) : l'instance t'appartient, la responsabilité aussi.

**Protections retenues (v1)** — pas de traitement spécial du bootstrap :

1. **Dernier admin** — impossible de désactiver/rétrograder le seul admin actif (déjà dans la spec).
2. **Confirmation UI** — avant désactivation d'un admin : « Assurez-vous qu'au moins un autre admin actif existe. »
3. **Documentation** — README / spec : `.env` = amorçage du **premier** démarrage uniquement.

**Explicitement hors scope v1 :** flag `seeded_from_env` en base, resync `.env`, compte bootstrap immortel, recréation automatique au boot.

---

## Objective

Today, admins create accounts only via `curl` against `POST /api/admin/users`. The web UI shows a placeholder at `/settings/users` with a "Soon" badge. For a family or small team deployment, this is unacceptable friction.

**Who:** Instance administrator (homelab operator, small-team admin).

**Why:** Manage who has access to the instance from the control panel — see accounts, create users, assign `admin` / `user` roles, revoke access, reset forgotten passwords — without leaving the browser.

**Pain points addressed:**

| Today | Target |
|---|---|
| User creation via API/curl only | Create-user form on `/settings/users` |
| No user list in UI | Table: email, role, status, created date |
| No role change flow | Inline role selector or action menu |
| No revoke-access flow | Disable toggle with confirmation |
| No password reset in UI | Admin-set new password (modal) |
| Sidebar "Soon" badge | Remove badge; page is live |

**User stories:**

| ID | Story |
|---|---|
| UM1 | As an admin, I see a list of all accounts on `/settings/users`. |
| UM2 | As an admin, I create a new account with email, password, and role (`user` default). |
| UM3 | As an admin, I promote a `user` to `admin` or demote an `admin` to `user` (including at account creation with role `admin`). |
| UM4 | As an admin, I disable an account so it can no longer sign in. |
| UM5 | As an admin, I re-enable a previously disabled account. |
| UM6 | As an admin, I set a new password for a user who forgot theirs. |
| UM7 | As an admin, I cannot demote or disable the last remaining admin — the API and UI block it. |
| UM8 | As a regular user, I cannot access `/settings/users` or admin user APIs (existing `AdminRoute` + `RequireAdmin`). |
| UM9 | As a disabled user, login fails with a clear message. |

**Out of scope (v1):**

- Public signup, registration page, invitation links
- OAuth / social login / SSO
- Password-reset email workflow
- Hard delete of user records
- Per-volume ACL / volume sharing
- Self-service "change my password" on Settings
- Bulk import/export of users
- User activity audit log
- Avatar / display name / profile fields beyond email
- **Change user email** — deferred; see *Account & admin model* and *Future scope*

### Future scope (explicitly not v1)

- **Admin changes user email** — `PATCH` field `email` + UI action; uniqueness validation; optional session revoke on change
- Hard delete of user records (with volume-ownership policy TBD)
- Self-service password change on Settings (separate IDEAS.md idea)

---

## Tech Stack

See [docs/project.md — Tech Stack](../../project.md#tech-stack).

This feature extends:

| Layer | Change |
|---|---|
| Backend | `internal/auth/` — `ListUsers`, `UpdateUser`, `SetPassword`, `DisableUser`/`EnableUser`, `disabled_at` on `User` |
| Backend | `internal/api/router.go` — `GET /api/admin/users`, `PATCH /api/admin/users/:id` |
| Frontend | `web/src/pages/UsersPage.tsx` — replace `PlaceholderPage` |
| Frontend | `web/src/components/users/` — table, form, modals |
| Frontend | `web/src/hooks/useUsers.ts` — React Query hooks |
| Frontend | `web/src/lib/api.ts` — admin user API methods |
| Database | PostgreSQL `users.disabled_at` (nullable timestamp) |

No new dependencies expected.

---

## Commands

### Development

```bash
# Backend — auth is critical path (80% coverage target)
go test ./internal/auth/...
go test ./internal/api/... -run AdminUser

# Frontend
cd web && npm run test
cd web && npm run build

# Full stack smoke
make embed
docker compose up -d --build
# → http://localhost:8080/settings/users (admin login)
```

### Lint / typecheck

```bash
go test ./...
cd web && npm run build
```

---

## Project Structure

See [docs/project.md — Project Structure](../../project.md#project-structure).

Feature-specific layout:

```
internal/auth/
├── model.go              → add DisabledAt *time.Time to User; extend UserResponse
├── service.go            → ListUsers, UpdateUser, SetPassword, disable/enable helpers
├── service_test.go       → last-admin, disable, login-blocked tests
└── errors.go             → ErrLastAdmin, ErrCannotDisableSelf (optional), ErrUserDisabled

internal/api/
├── router.go             → GET/PATCH admin users routes
└── router_test.go        → admin user list/patch tests

web/src/
├── pages/UsersPage.tsx
├── components/users/
│   ├── UserTable.tsx
│   ├── CreateUserForm.tsx
│   ├── ResetPasswordDialog.tsx
│   └── RoleBadge.tsx
├── hooks/useUsers.ts
└── lib/api.ts            → listAdminUsers, patchAdminUser, createAdminUser (exists)

docs/features/user-management-ui/
└── SPEC.md               → this file
```

---

## API Design

### `GET /api/admin/users`

**Auth:** Admin only.

**Response `200`:**

```json
{
  "users": [
    {
      "id": "uuid",
      "email": "member@example.com",
      "role": "user",
      "created_at": "2026-07-01T10:00:00Z",
      "disabled_at": null
    }
  ]
}
```

- Sorted by `created_at` ascending (stable, predictable).
- `disabled_at` null = active; non-null = disabled.

### `POST /api/admin/users` *(existing)*

Unchanged. Body `{ email, password, role? }`. Returns `UserResponse` without password.

### `PATCH /api/admin/users/:id`

**Auth:** Admin only.

**Body** (all fields optional; at least one required):

```json
{
  "role": "admin",
  "disabled": true,
  "password": "newpassword1"
}
```

**Behavior:**

| Field | Effect |
|---|---|
| `role` | Change role (`admin` \| `user`). Blocked if target is sole active admin and new role is `user`. |
| `disabled` | `true` → set `disabled_at = now()`, revoke all refresh tokens. `false` → clear `disabled_at`. Blocked if sole active admin. |
| `password` | bcrypt-hash and replace password (min 8 chars, same validation as create). |

**Response `200`:** Updated `UserResponse` (include `disabled_at`).

**Errors:**

| Status | Condition |
|---|---|
| `400` | Invalid body, invalid role, password too short, empty patch |
| `403` | Last-admin safeguard triggered |
| `404` | User ID not found |
| `409` | N/A for patch (email not patchable in v1) |

### Login behavior change

`auth.Service.Login` returns `ErrUserDisabled` (mapped to `401` + `"account disabled"`) when `disabled_at` is set.

---

## Code Style

Follow existing Go service + Gin handler + React Query patterns.

**UserResponse extension:**

```go
type UserResponse struct {
    ID         uuid.UUID  `json:"id"`
    Email      string     `json:"email"`
    Role       Role       `json:"role"`
    CreatedAt  time.Time  `json:"created_at"`
    DisabledAt *time.Time `json:"disabled_at"`
}
```

**Last-admin check (service layer):**

```go
func (s *Service) countActiveAdmins(excludeID uuid.UUID) (int64, error) {
    var count int64
    err := s.db.Model(&User{}).
        Where("role = ? AND disabled_at IS NULL AND id != ?", RoleAdmin, excludeID).
        Count(&count).Error
    return count, err
}
```

Before demoting or disabling user `id`, if user is `admin` and `countActiveAdmins(id) == 0`, return `ErrLastAdmin`.

**Frontend hook pattern** (match `useTasks.ts`):

```tsx
export function useAdminUsers() {
  return useQuery({
    queryKey: ["admin", "users"],
    queryFn: () => api.listAdminUsers(),
    staleTime: 15_000,
  });
}
```

**UI conventions** (per [design/DESIGN.md](../../../design/DESIGN.md)):

| Element | Style |
|---|---|
| Page header | `Header` component — title "Users", description about team access |
| Primary CTA | Yellow `Button` — "Create user" with `ActionIcon action="create"` |
| Role badges | `admin` → `Badge` with `text-primary` border; `user` → muted outline |
| Disabled row | `text-muted` + `Status: Disabled` badge (`text-accent-rose`) |
| Destructive confirm | Disable uses confirm dialog; copy explains revoked sessions |
| Table | Card-wrapped; columns: Email, Role, Status, Created, Actions |
| Forms | Card sections; email + password + role select; inline validation errors from API |

---

## Testing Strategy

See [docs/project.md — Coverage targets](../../project.md#coverage-targets).

| Layer | Focus | Location |
|---|---|---|
| Backend unit | `ListUsers`, disable/enable, last-admin guard, login blocked when disabled | `internal/auth/service_test.go` |
| Backend unit | `SetPassword` validation | `internal/auth/service_test.go` |
| Backend integration | `GET/PATCH /api/admin/users` auth + happy paths | `internal/api/router_test.go` |
| Backend integration | Regular user forbidden on admin routes | `internal/api/router_test.go` |
| Frontend unit | `RoleBadge` renders correct label | optional |
| Manual | Full admin flow in browser | checklist below |

**Manual smoke checklist:**

1. Log in as admin → `/settings/users` loads user table (includes seeded admin).
2. Sidebar "Users" link has no "Soon" badge.
3. Create user with role `user` → appears in table; can log in as that user.
4. Promote user to `admin` → badge updates; user sees admin nav items after re-login.
5. Disable user → user cannot log in ("account disabled"); row shows Disabled status.
6. Re-enable user → login works again.
7. Reset password → user logs in with new password.
8. Attempt to demote/disable sole admin → API error + UI shows message.
9. Log in as regular user → `/settings/users` redirects to dashboard.
10. Disabled user's existing session: refresh fails after disable (tokens revoked).

---

## Boundaries

- **Always:** Run `go test ./internal/auth/...` and `cd web && npm run build` before PR; enforce last-admin safeguard in service layer (not UI-only); revoke refresh tokens on disable; map `disabled_at` in `UserResponse`.
- **Ask first:** Hard delete endpoint; email patch; OAuth; changing JWT expiry strategy; ADR if we later tie user deletion to volume ownership transfer.
- **Never:** Add public signup route; store plaintext passwords; bypass `RequireAdmin()` middleware; write user config exclusively outside PostgreSQL (users are DB-only, not `.volume.json`); commit secrets.

---

## UI Specification

### Page layout (`UsersPage`)

```
┌─────────────────────────────────────────────────────────┐
│ Header: Users — Manage who can access this instance.    │
├─────────────────────────────────────────────────────────┤
│ [Create user]                              (top-right)  │
├─────────────────────────────────────────────────────────┤
│ ┌─ User table (Card) ─────────────────────────────────┐ │
│ │ Email          │ Role   │ Status  │ Created │ ⋮    │ │
│ │ admin@…        │ admin  │ Active  │ Jul 1   │ …    │ │
│ │ member@…       │ user   │ Active  │ Jul 3   │ …    │ │
│ └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Create user form

- Triggered by "Create user" button → inline Card below header (same pattern as TasksPage form toggle) or slide-over Card.
- Fields: Email (required), Password (required, min 8), Role (select: `user` default, `admin`).
- Submit → `POST /api/admin/users` → invalidate list → collapse form.
- Errors: "email already taken" (409), validation messages inline.

### Row actions (dropdown menu)

| Action | Behavior |
|---|---|
| Change role | Toggle or select `admin` / `user`; confirm if demoting admin |
| Disable / Enable | Toggle; confirm dialog on disable |
| Reset password | Opens `ResetPasswordDialog` — new password + confirm field |

Current user row: show "(you)" next to email; disable/demote self allowed only when another active admin exists (otherwise action disabled + tooltip).

### Empty state

When only the seeded admin exists and list is otherwise empty (edge case after deletes in future): show centered `Users` icon + "No additional users yet — create one to share access."

---

## Implementation plan (vertical slices)

Dependency order — backend list + disable model before UI.

### Slice 1 — Data model + list API (UM1, UM8)

**Files:** `internal/auth/model.go`, `service.go`, `router.go`, `router_test.go`, `api.ts`, `useUsers.ts`

- Add `DisabledAt` to `User` + `UserResponse`.
- Implement `ListUsers()`.
- `GET /api/admin/users` handler + tests.
- Frontend: fetch hook only (no page yet) — optional dev verification via console.

**Checkpoint:** `go test ./internal/auth/... ./internal/api/...` green.

### Slice 2 — Disable + login guard (UM4, UM5, UM9)

**Files:** `service.go`, `service_test.go`, `router.go`, `router_test.go`

- `UpdateUser` with `disabled` field; revoke refresh tokens on disable.
- `Login` rejects disabled accounts.
- Last-admin safeguard for disable/demote.
- `PATCH /api/admin/users/:id` (disable/enable only) + tests.

**Checkpoint:** Login-blocked test passes; last-admin test passes.

### Slice 3 — Role change + password reset API (UM3, UM6, UM7)

**Files:** `service.go`, `router.go`, tests

- Extend `PATCH` for `role` and `password`.
- `SetPassword` with bcrypt + min length validation.

**Checkpoint:** Full PATCH coverage in router tests.

### Slice 4 — Users page UI (UM1–UM7)

**Files:** `UsersPage.tsx`, `components/users/*`, `useUsers.ts`, `api.ts`, `App.tsx`, `Sidebar.tsx`

- Replace `PlaceholderPage` with `UsersPage`.
- User table + create form + row actions + reset password dialog.
- Remove `soon: true` from Sidebar Users item.
- React Query mutations with list invalidation.

**Checkpoint:** Manual smoke checklist items 1–9 pass.

### Slice 5 — Polish + edge cases

- Self-action warnings ("You are about to disable your own account").
- Loading/error states matching TasksPage patterns.
- Confirm dialogs for destructive actions.

**Checkpoint:** Full manual smoke checklist; `cd web && npm run test && npm run build` green.

---

## Success Criteria

User management UI is **done** when all of the following pass:

- [ ] **SC-UM1** `GET /api/admin/users` returns all users with `disabled_at`; admin-only; covered by tests.
- [ ] **SC-UM2** `PATCH /api/admin/users/:id` supports `role`, `disabled`, `password`; last-admin safeguard enforced server-side.
- [ ] **SC-UM3** Disabled users cannot log in; refresh tokens revoked on disable.
- [ ] **SC-UM4** `/settings/users` shows user table with create form and row actions.
- [ ] **SC-UM5** Sidebar Users link has no "Soon" badge.
- [ ] **SC-UM6** Admin can create, change role, disable, re-enable, and reset password entirely from the UI.
- [ ] **SC-UM7** Regular users cannot access page or APIs (403/redirect).
- [ ] **SC-UM8** `go test ./internal/auth/...` and `cd web && npm run build` pass.
- [ ] **SC-UM9** UI follows [design/DESIGN.md](../../../design/DESIGN.md) tokens; role badges and confirm dialogs present.

---

## Open Questions

| ID | Question | Default if no answer |
|---|---|---|
| OQ1 | Hard delete in v1 or soft-disable only? | **Soft-disable only** (assumption #3) |
| OQ2 | Can admin patch their own role/disable themselves? | **Yes, if another active admin exists**; otherwise blocked |
| OQ3 | Create form: inline Card (Tasks pattern) or modal? | **Inline Card toggle** — matches TasksPage |
| OQ4 | Show `updated_at` column in table? | **No** — email, role, status, created is enough for v1 |
| OQ5 | Email editable via PATCH in v1? | **No** — deferred to future scope; admin email change is planned, not ruled out |
| OQ6 | Sort order for user list? | **`created_at` ascending** — admin first (seeded), then chronological |
| OQ7 | Password reset: require confirm-password field? | **Yes** — client-side match before submit |

---

## Relationship to other work

| Related | Relationship |
|---|---|
| [IDEAS.md — User profile — change own password](../../../IDEAS.md) | Complements this spec; self-service on Settings page |
| [IDEAS.md — In-instance volume sharing](../../../IDEAS.md) | Depends on User management UI for ACL grants |
| [IDEAS.md — Unified Spotlight search](../../../IDEAS.md) | Users category blocked until `GET /api/admin/users` ships |
| [IDEAS.md — Control panel iconography](../../../IDEAS.md) | Users nav icon already mapped; page gets empty-state icon |
| Phase 0 `POST /api/admin/users` | Reused by create form — no API change needed |

---

## Revision history

| Date | Change |
|---|---|
| 2026-07-03 | Initial draft from IDEAS.md — User management UI (admin) |
| 2026-07-03 | Bootstrap admin lifecycle: env = one-shot seed; bootstrap admin may be disabled like any other |
