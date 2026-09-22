---
description: "Task list for 002-staff-auth"
---

# Tasks: Autenticación de staff y sesión de invitado

**Input**: Design documents from `/specs/002-staff-auth/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/openapi.yaml](./contracts/openapi.yaml), [quickstart.md](./quickstart.md)

**Tests**: Incluidos. Spec y constitución exigen login/CRUD/aislamiento A-vs-B y contratos OpenAPI.

**Organization**: Por historia de usuario. Setup y fundación bloquean todas las historias.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Paralelo (archivos distintos, sin depender de tareas incompletas del mismo lote)
- **[Story]**: US1–US4 según spec.md
- Cada descripción incluye ruta de archivo

## Path Conventions

- Backend: `backend/`
- Frontend: `frontend/`
- Raíz: `docker-compose.yml`, `.env.example`, `README.md`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Dependencias y secretos de auth. Sin lógica de login todavía.

- [x] T001 Add `AUTH_JWT_SECRET`, `STAFF_SEED_PASSWORD`, and Compose `API_INTERNAL_URL` to `.env.example` and pass them in `docker-compose.yml` (`api` + `web`)
- [x] T002 [P] Add `github.com/golang-jwt/jwt/v5`, promote `golang.org/x/crypto`, and `github.com/go-playground/validator/v10` in `backend/go.mod`
- [x] T003 [P] Replace `backend/internal/auth/README.md` with a short module note (package lives here; no domain yet)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Arreglar `resolve_tenant`, tablas RLS de staff/guest/refresh, config, CORS, middleware JWT. Ninguna historia empieza antes.

**CRITICAL**: US1–US4 dependen de esta fase

- [x] T004 Set `resolve_tenant` to `SECURITY DEFINER` plus `SET row_security = off` in `backend/migrations/schema.sql` so Host/slug lookup works under FORCE RLS
- [x] T005 [P] Add Ent schemas `Staff`, `StaffRefreshToken`, and `GuestSession` in `backend/ent/schema/staff.go`, `backend/ent/schema/staff_refresh_token.go`, and `backend/ent/schema/guest_session.go` per [data-model.md](./data-model.md)
- [x] T006 Add SQL for `staff`, `staff_refresh_tokens`, `guest_sessions` (FORCE RLS, GRANTs to `doap_app`, unique `(tenant_id, email)`) in `backend/migrations/schema.sql`
- [x] T007 Hash and insert seed owners `owner@demo-a.local` / `owner@demo-b.local` from `STAFF_SEED_PASSWORD` in `backend/migrations/apply.sh` (and `backend/migrations/seed.sql` if needed)
- [x] T008 Require `AUTH_JWT_SECRET` (≥32 chars) in `backend/internal/platform/config/config.go` with fail-fast named key
- [x] T009 Restrict CORS to the web origin and allow `Authorization` in `backend/internal/platform/http/router.go` (replace `Access-Control-Allow-Origin: *`)
- [x] T010 Add staff JWT middleware (Bearer, `tenant_id` must match Host) in `backend/internal/platform/middleware/staff_jwt.go`
- [x] T011 Add JSON errors `invalid_credentials`, `forbidden`, `rate_limited` in `backend/internal/platform/http/errors.go`
- [x] T012 Add in-memory login rate limit (10/min per ip+tenant+email) in `backend/internal/auth/ratelimit.go`

**Checkpoint**: Recreate Compose; `GET /v1/tenants/current/by-slug/demo-a` returns Demo A. API still has no login routes until US1.

---

## Phase 3: User Story 1 - Entrar al backoffice y ver que la sesión sirve (Priority: P1) MVP

**Goal**: Login dueño demo, JWT + refresh rotativo, `/v1/auth/me`, pantallas `/t/{slug}/login` y `/hello`, logout.

**Independent Test**: Login `owner@demo-a.local` + seed password → me con rol owner y Demo A; logout; hello pide acceso otra vez. Mal password → 401 genérico.

### Tests for User Story 1

- [ ] T013 [P] [US1] Write failing contract tests for `POST /v1/auth/login`, `POST /v1/auth/refresh`, `POST /v1/auth/logout`, `GET /v1/auth/me` in `backend/internal/auth/login_contract_test.go` against [contracts/openapi.yaml](./contracts/openapi.yaml)
- [ ] T014 [P] [US1] Write failing unit tests for bcrypt cost 12 and refresh rotation in `backend/internal/auth/tokens_test.go`

### Implementation for User Story 1

- [ ] T015 [US1] Implement staff lookup + bcrypt verify in `backend/internal/auth/store.go` (after `resolve_tenant` + `SET LOCAL`)
- [ ] T016 [US1] Implement access JWT (15m) and rotating hashed refresh (7d) in `backend/internal/auth/tokens.go`
- [ ] T017 [US1] Implement login, refresh, logout, me handlers in `backend/internal/auth/http.go`
- [ ] T018 [US1] Mount auth routes on the chi router in `backend/internal/platform/http/router.go`
- [ ] T019 [US1] Add Next route handlers that set httpOnly cookies from the API token pair in `frontend/app/api/auth/login/route.ts`, `frontend/app/api/auth/refresh/route.ts`, and `frontend/app/api/auth/logout/route.ts`
- [ ] T020 [P] [US1] Add login screen in `frontend/app/t/[slug]/login/page.tsx`
- [ ] T021 [US1] Add post-login hello screen (name, role, tenant) in `frontend/app/t/[slug]/hello/page.tsx` using `frontend/lib/api.ts`

**Checkpoint**: Quickstart sección Login y saludo. MVP demostrable.

---

## Phase 4: User Story 2 - CRUD de personal del comercio (Priority: P1)

**Goal**: Dueño lista/crea/edita/desactiva/reactiva caja y cocina. Caja/cocina 403. Sin crear owner. Password opcional en PATCH.

**Independent Test**: Crear cashier, editar nombre, cambiar password, desactivar (no login), reactivar (login). Login cashier → POST staff 403. POST role owner → 422/403.

### Tests for User Story 2

- [ ] T022 [P] [US2] Write failing contract tests for `GET/POST /v1/staff`, `PATCH /v1/staff/{id}`, deactivate/reactivate in `backend/internal/auth/staff_contract_test.go`

### Implementation for User Story 2

- [ ] T023 [US2] Implement staff list/create/patch/deactivate/reactivate (unique email, bcrypt on create/optional patch, revoke refresh) in `backend/internal/auth/store.go`
- [ ] T024 [US2] Implement staff HTTP handlers and owner-only guard in `backend/internal/auth/http.go` and register in `backend/internal/platform/http/router.go`
- [ ] T025 [US2] Reject `role=owner` on create/patch and reject deactivate of owner in `backend/internal/auth/store.go`
- [ ] T026 [US2] Add staff CRUD screen in `frontend/app/t/[slug]/staff/page.tsx`

**Checkpoint**: Quickstart sección CRUD.

---

## Phase 5: User Story 3 - El personal de un comercio no cruza al otro (Priority: P1)

**Goal**: Staff y tokens de A no sirven en B. RLS FORCE. `X-Tenant-Id` ignorado.

**Independent Test**: Dueño A no lista ni muta staff de B. Bearer A + Host B → 401. `go test` aislamiento.

### Tests for User Story 3

- [ ] T027 [P] [US3] Write failing isolation tests (A cannot SELECT/UPDATE staff of B; no `app.tenant_id` → 0 rows) in `backend/internal/auth/isolation_test.go`
- [ ] T028 [P] [US3] Write failing test JWT of A on Host B returns 401 in `backend/internal/auth/host_mismatch_test.go`

### Implementation for User Story 3

- [ ] T029 [US3] Ensure every staff/refresh query runs inside `WithTenantTx` in `backend/internal/auth/store.go` and middleware in `backend/internal/platform/middleware/staff_jwt.go`
- [ ] T030 [US3] Confirm `X-Tenant-Id` is still stripped in `backend/internal/platform/middleware/tenant.go` for auth routes

**Checkpoint**: SC-003. Isolation tests green with Compose Postgres.

---

## Phase 6: User Story 4 - Invitado de vitrina, distinto del staff (Priority: P2)

**Goal**: Al abrir `/t/{slug}` se asegura sesión guest `pwa`. No abre hello ni CRUD.

**Independent Test**: Abrir vitrina demo-a sin botón; cookie guest. `/hello` y `/staff` piden login. Guest + `/v1/staff` → 401.

### Tests for User Story 4

- [ ] T031 [P] [US4] Write failing contract tests for `POST /v1/guest/sessions` (reuse same tenant, new token other tenant, 404 unknown host) in `backend/internal/auth/guest_contract_test.go`

### Implementation for User Story 4

- [ ] T032 [US4] Implement guest ensure/reuse hashed token in `backend/internal/auth/guest.go` and handler in `backend/internal/auth/http.go`
- [ ] T033 [US4] Call guest ensure from the storefront in `frontend/app/t/[slug]/page.tsx` and `frontend/app/page.tsx` via `frontend/app/api/auth/guest/route.ts`
- [ ] T034 [US4] Block guest cookies from `frontend/app/t/[slug]/hello/page.tsx` and `frontend/app/t/[slug]/staff/page.tsx` (redirect to login)

**Checkpoint**: Quickstart sección Invitado. SC-006.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Docs, lints, quickstart e2e

- [ ] T035 [P] Align root `README.md` with [quickstart.md](./quickstart.md) (seed owners, `/t/demo-a/login`)
- [ ] T036 [P] Document staff/guest RLS template in `backend/migrations/README.md`
- [ ] T037 Run `go test ./...` in `backend/` and `npx tsc --noEmit` in `frontend/`
- [ ] T038 Execute [quickstart.md](./quickstart.md) on Compose (`down -v` + up) and fix gaps

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sin dependencias
- **Foundational (Phase 2)**: depende de Setup — BLOQUEA historias
- **US1 (Phase 3)**: depende de Phase 2 — MVP
- **US2 (Phase 4)**: depende de US1 (CRUD usa sesión owner)
- **US3 (Phase 5)**: depende de US1 (tokens) y preferible US2 (filas staff de prueba)
- **US4 (Phase 6)**: depende de Phase 2; puede ir en paralelo con US2 si el BFF de cookies ya existe
- **Polish (Phase 7)**: después de las historias que se entreguen

### User Story Dependencies

- **US1**: después de Phase 2
- **US2**: después de US1 (dueño autenticado)
- **US3**: después de US1; mejor con US2 para mutar staff ajeno
- **US4**: después de Phase 2 + pantallas de US1 para demostrar el bloqueo

### Parallel Opportunities

- T002 / T003
- T005 vs T008–T012 (archivos distintos; T004+T006+T007 en `schema.sql`/`apply.sh` son sequential)
- T013 / T014
- T020 con T019 si el contrato de cookies está acordado
- T027 / T028
- T035 / T036

---

## Parallel Example: User Story 1

```text
Task: "Contract tests in backend/internal/auth/login_contract_test.go"
Task: "Token unit tests in backend/internal/auth/tokens_test.go"
```

Luego sequential: store → tokens → http → router → BFF → hello.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 + 2 (incluye arreglo `resolve_tenant`)
2. Phase 3 US1
3. STOP: login + hello con dueño semilla
4. Luego US2 CRUD visible
5. US3 aislamiento
6. US4 invitado

### Incremental Delivery

1. Fundación → by-slug Demo A funciona
2. US1 → demo login
3. US2 → demo CRUD
4. US3 → tests A-vs-B
5. US4 → vitrina emite guest

---

## Notes

- Implementar con `/speckit-implement`, no con `/speckit-plan`
- Isolation tests en el host (`go test`), no en la imagen distroless
- Contraseñas y JWT secret nunca en el frontend ni en git
- Commit por tarea o grupo lógico si el usuario lo pide
