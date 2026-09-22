---
description: "Task list for 001-monolith-scaffold"
---

# Tasks: Esqueleto del monolito multi-tenant

**Input**: Design documents from `/specs/001-monolith-scaffold/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/openapi.yaml](./contracts/openapi.yaml), [quickstart.md](./quickstart.md)

**Tests**: Incluidos. La spec exige test automatizado A-vs-B (FR-007, SC-002) y el plan pide tests de contrato HTTP.

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

**Purpose**: Árbol del repo, módulos Go/Next y plantilla de secretos. Sin lógica de negocio.

- [x] T001 Create the modular tree with `.gitkeep` in `backend/internal/auth/`, `backend/internal/catalog/`, `backend/internal/cart/`, `backend/internal/orders/`, `backend/internal/payments/`, `backend/internal/agent/`, `backend/internal/channel/`, `frontend/src/mcp-apps/`, `frontend/src/webmcp/`, and stub `backend/cmd/worker/main.go` that prints not implemented
- [x] T002 Initialize Go module `github.com/doap/platform` (or module path matching the repo) in `backend/go.mod` with chi, ent, pgx/v5, validator/v10
- [x] T003 [P] Scaffold Next.js App Router + TypeScript + Tailwind in `frontend/package.json`, `frontend/app/layout.tsx`, `frontend/app/page.tsx`
- [x] T004 [P] Write `.env.example` at repo root with `POSTGRES_OWNER_USER`, `POSTGRES_OWNER_PASSWORD`, `POSTGRES_APP_USER`, `POSTGRES_APP_PASSWORD`, `POSTGRES_DB`, `POSTGRES_HOST`, `API_ADDR`, `NEXT_PUBLIC_API_URL` (local placeholders only)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Config, pool, Ent, Atlas, Docker, router. Ninguna historia de usuario empieza antes.

**CRITICAL**: US1–US4 dependen de esta fase

- [x] T005 Implement fail-fast typed config in `backend/internal/platform/config/config.go` (missing env names the key; no silent default secrets)
- [x] T006 Implement pgx pool and `WithTenantTx` (`SET LOCAL app.tenant_id`) in `backend/internal/platform/database/database.go`
- [x] T007 [P] Create Ent schemas `Tenant` and `IsolationCanary` in `backend/ent/schema/tenant.go` and `backend/ent/schema/isolation_canary.go` per [data-model.md](./data-model.md)
- [x] T008 Add Atlas project and SQL migrations for roles `doap_owner`/`doap_app`, `ENABLE`+`FORCE ROW LEVEL SECURITY`, policy on `tenant_id`/`id`, and `resolve_tenant(host, slug)` SECURITY DEFINER in `backend/migrations/`
- [x] T009 Wire HTTP router, JSON errors (`tenant_not_found`, `validation_error`), and ignore `X-Tenant-Id` as authority in `backend/internal/platform/middleware/tenant.go` and `backend/cmd/api/main.go`
- [x] T010 Add `backend/Dockerfile` (Go 1.24 → `gcr.io/distroless/base-debian12`), `frontend/Dockerfile`, and `docker-compose.yml` (postgres:16, migrate, api, web) with healthchecks

**Checkpoint**: `docker compose build` succeeds; API still has no domain routes except what US1 adds

---

## Phase 3: User Story 1 - Arrancar el entorno local completo (Priority: P1) MVP

**Goal**: Un clone + `.env` + Compose deja vivos datos, API y vitrina. Secretos no versionados.

**Independent Test**: `copy .env.example .env`; `docker compose up --build`; `/healthz` y `/readyz` `ok`; vitrina carga; `.env` no está en git.

### Tests for User Story 1

- [x] T011 [P] [US1] Add contract tests for `GET /healthz` and `GET /readyz` in `backend/internal/platform/http/health_contract_test.go` against [contracts/openapi.yaml](./contracts/openapi.yaml)
- [x] T012 [US1] Add a test that the process exits with a named missing key when config is incomplete in `backend/internal/platform/config/config_test.go`

### Implementation for User Story 1

- [x] T013 [US1] Implement `GET /healthz` (no DB) and `GET /readyz` (DB ping) in `backend/internal/platform/http/health.go` and register in `backend/cmd/api/main.go`
- [x] T014 [US1] Add Compose `migrate` service that runs Atlas as owner before `api` starts, using `docker-compose.yml` and `backend/migrations/`
- [x] T015 [US1] Point the storefront at `NEXT_PUBLIC_API_URL` with a minimal live page in `frontend/app/page.tsx`
- [x] T016 [US1] Document copy env + `docker compose up --build` in `README.md` (link [quickstart.md](./quickstart.md))

**Checkpoint**: Quickstart steps 1–señales de vida pasan

---

## Phase 4: User Story 2 - Dos comercios aislados desde el primer día (Priority: P1)

**Goal**: Semilla demo-a/demo-b (COP, IVA, Bogotá). Host manda; slug es fallback; header de tenant se ignora.

**Independent Test**: `curl -H "Host: demo-a.localhost"` vs demo-b; `X-Tenant-Id` inventado no cambia A; host desconocido 404.

### Tests for User Story 2

- [x] T017 [P] [US2] Add contract tests for `GET /v1/tenants/current` (Host, X-Forwarded-Host, ignored X-Tenant-Id, 404) and `GET /v1/tenants/current/by-slug/{slug}` in `backend/internal/tenant/http_contract_test.go`

### Implementation for User Story 2

- [x] T018 [US2] Implement `resolve_tenant` usage and Host-then-slug middleware in `backend/internal/platform/middleware/tenant.go` (host stripped of port; host wins over slug)
- [x] T019 [US2] Implement tenant handlers returning `TenantPublic` in `backend/internal/tenant/http.go` for `/v1/tenants/current` and `/v1/tenants/current/by-slug/{slug}`
- [x] T020 [US2] Add idempotent seed of demo-a (`demo-a.localhost`) and demo-b (`demo-b.localhost`) with COP/IVA/`America/Bogota` in `backend/migrations/`
- [x] T021 [US2] Show current commerce name on `frontend/app/page.tsx` and `frontend/app/t/[slug]/page.tsx` using Host or slug fallback

**Checkpoint**: Quickstart secciones comercio por host, id inventado, fallback slug

---

## Phase 5: User Story 3 - Mapa de módulos listo (Priority: P2)

**Goal**: El árbol coincide con docs; la API arranca sin auth/catálogo/pedidos; tenant por slug/host ya existe (US2).

**Independent Test**: Inspeccionar carpetas de plan.md; `/healthz` ok; no hay rutas de pedido; worker es hueco.

### Implementation for User Story 3

- [x] T022 [P] [US3] Confirm stub modules remain empty of business logic and add a one-line README in `backend/internal/auth/README.md`, `backend/internal/catalog/README.md`, `backend/internal/cart/README.md`, `backend/internal/orders/README.md`, `backend/internal/payments/README.md`, `backend/internal/agent/README.md`, `backend/internal/channel/README.md` stating not in this feature
- [x] T023 [US3] Keep `backend/cmd/worker/main.go` as a documented no-op (not started by `docker-compose.yml`)
- [x] T024 [US3] Assert API has no order/cart/payment/agent routes (test in `backend/internal/platform/http/no_domain_routes_test.go`)

**Checkpoint**: Revisor señala cada módulo; pedido no es usable

---

## Phase 6: User Story 4 - Esquema que fuerza el aislamiento (Priority: P1)

**Goal**: FORCE RLS, rol app ≠ owner, SET LOCAL, canarios, test A no lee/muta B. Recrear Compose restaura semilla.

**Independent Test**: `go test` de canarios: contexto A → 0 filas de B; sin `app.tenant_id` → 0 filas de negocio; `compose down -v` + up restaura dos comercios.

### Tests for User Story 4

- [x] T025 [US4] Write failing-then-passing isolation tests in `backend/internal/tenant/isolation_test.go`: as `doap_app`, SET LOCAL A, SELECT/UPDATE canary B yields 0 rows; missing setting yields 0 business rows

### Implementation for User Story 4

- [x] T026 [US4] Seed one `isolation_canaries` row per demo tenant in `backend/migrations/`
- [x] T027 [US4] Verify Atlas SQL uses FORCE RLS and GRANTs only to `doap_app` (not table owner as API user) in `backend/migrations/` and document the future-table policy template in `backend/migrations/README.md`
- [x] T028 [US4] Run recreate path `docker compose down -v` && `up` and record expected seed in [quickstart.md](./quickstart.md)

**Checkpoint**: SC-002 y SC-004

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Alinear docs, lints y validación del quickstart

- [x] T029 [P] Align `README.md` with [quickstart.md](./quickstart.md) (hosts file note for `demo-a.localhost`)
- [x] T030 [P] Add `frontend/public/manifest.webmanifest` stub (no caching authenticated responses; no SW that replays orders)
- [x] T031 Run `go test ./...` in `backend/` and `npx tsc --noEmit` in `frontend/`
- [ ] T032 Execute [quickstart.md](./quickstart.md) end-to-end on Compose and fix gaps

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sin dependencias
- **Foundational (Phase 2)**: depende de Setup; BLOQUEA historias
- **US1 (Phase 3)**: después de fundación — MVP
- **US2 (Phase 4)**: después de US1 (API y migrate vivos); usa semilla y resolver
- **US3 (Phase 5)**: después de Setup (carpetas); el test de “sin rutas” espera el router de Phase 2
- **US4 (Phase 6)**: después de US2 (semilla de tenants); canarios + test RLS
- **Polish**: después de las historias que se entreguen

### User Story Dependencies

- **User Story 1 (P1)**: solo fundación
- **User Story 2 (P1)**: conviene después de US1 (Compose + migrate)
- **User Story 3 (P2)**: paralelo a US2 en carpetas; el test de rutas espera router
- **User Story 4 (P1)**: después de US2 (necesita demo-a/demo-b)

### Within Each User Story

- Tests primero y en rojo
- Migraciones/modelo antes de handlers
- Handlers antes de vitrina
- Historia cerrada antes de subir de prioridad

### Parallel Opportunities

- T003 y T004 en paralelo tras T001
- T007 en paralelo a T005–T006 si el módulo Go ya existe
- T011 y T012 en paralelo
- T022 en paralelo a T023
- T029 y T030 en paralelo

---

## Parallel Example: User Story 1

```text
Task: "Contract tests healthz/readyz in backend/internal/platform/http/health_contract_test.go"
Task: "Config missing-key test in backend/internal/platform/config/config_test.go"
```

## Parallel Example: User Story 2

```text
Task: "Contract tests current tenant in backend/internal/tenant/http_contract_test.go"
```

(Handlers T018–T021 son secuenciales sobre el mismo middleware/router.)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 Setup
2. Phase 2 Fundación
3. Phase 3 US1
4. STOP: `docker compose up` + healthz/readyz

### Incremental Delivery

1. Setup + Fundación
2. US1 → entorno vivo
3. US2 → dos comercios por Host
4. US4 → RLS comprobado (no saltar: constitución IV)
5. US3 → mapa de módulos visible
6. Polish + quickstart completo

### Parallel Team Strategy

Tras Phase 2: una persona en US1 Dockerfile/Compose, otra en Ent/Atlas (sigue siendo fundación si no terminó). Tras US1: resolver/handlers (US2) y stubs de módulos (US3) en paralelo. US4 cuando la semilla exista.

---

## Notes

- No implementar auth, catálogo, carrito, pedidos, pagos, agente, WhatsApp, Mercado Pago
- No usar GORM ni scratch sin CA
- La API usa `doap_app`; Atlas usa `doap_owner`
- Host sin puerto gana sobre slug; `X-Tenant-Id` nunca es autoridad
