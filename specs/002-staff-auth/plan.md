# Implementation Plan: Autenticación de staff y sesión de invitado

**Branch**: `002-staff-auth` | **Date**: 2026-09-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/002-staff-auth/spec.md`

## Summary

Módulo `auth`: login de staff (bcrypt 12, JWT corto + refresh rotativo), CRUD de caja/cocina por dueño, saludo mínimo post-acceso, sesión de invitado automática en vitrina, aislamiento A-vs-B. Dueños demo solo por semilla. Sin OAuth, recovery por correo, catálogo ni pedidos.

## Technical Context

**Language/Version**: Go 1.24 (API); TypeScript (Next.js App Router) en vitrina y pantallas de prueba de staff

**Primary Dependencies**: chi, pgx, validator/v10, `golang.org/x/crypto/bcrypt`, `github.com/golang-jwt/jwt/v5`; Next.js + Tailwind. Ent schemas opcionales; persistencia como `tenant` (SQL Atlas + pgx). GORM no.

**Storage**: PostgreSQL 16. Tablas `staff`, `staff_refresh_tokens`, `guest_sessions` con `tenant_id`, FORCE RLS, `doap_app` ≠ owner, `SET LOCAL app.tenant_id`. Función `resolve_tenant` con `row_security = off` (el FORCE actual deja el lookup en cero filas).

**Testing**: `go test` contratos OpenAPI (login, me, CRUD, 401/403, aislamiento); integración RLS staff/guest; `tsc --noEmit`. Isolation tests en el host, no en la imagen distroless.

**Target Platform**: Compose local (Windows host + Linux containers). No producción.

**Project Type**: web (backend modular + frontend)

**Performance Goals**: login y saludo en < 30 s en entorno ya arrancado (SC-001); CRUD caja + login en < 2 min (SC-002). Sin SLO de 15 ms.

**Constraints**: `AUTH_JWT_SECRET` solo en env, no en el frontend; CORS no puede seguir `*`; tenant por host/slug, JWT `tenant_id` debe coincidir; acceso 15 min, refresh 7 días rotativo; contraseña ≥ 8; tope de login 10/min por IP+correo+comercio (memoria, local).

**Scale/Scope**: dos dueños semilla; pantallas login / equipo / saludo; invitado PWA; cero catálogo.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Este corte | Estado |
| --- | --- | --- |
| I Comercio dueño / Colombia | Sin cambios de mercado; staff por comercio | Pass |
| II API HTTP es el límite | [contracts/openapi.yaml](contracts/openapi.yaml); pantallas Next son clientes | Pass |
| III Modelo sin autoridad de dinero | Sin precios ni agente | Pass (N/A) |
| IV Aislamiento comprobable | RLS en staff/guest/refresh; test A no lee/muta B; host gana a `X-Tenant-Id` | Pass |
| V Pedido atómico | Fuera de alcance | Pass (N/A) |
| Identidades staff + invitado | JWT staff + sesión opaca invitado; no mezclar | Pass |
| bcrypt 12 + JWT corto + refresh rotativo | research.md | Pass |
| Admin plataforma / OAuth / recovery | Explicitamente no | Pass |
| Orden de módulos | `auth` después de `tenant` | Pass |

Post-diseño (Phase 1): cookies httpOnly las pone Next en su origen (BFF) porque `localhost:3000` no recibe `Set-Cookie` de `:8080`. Los JWT viven en el contrato JSON de la API. `resolve_tenant` + `row_security = off` no relaja RLS de tablas de negocio. Sin violaciones en Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/002-staff-auth/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/openapi.yaml
└── tasks.md              # /speckit-tasks — no creado aquí
```

### Source Code (repository root)

```text
backend/
├── cmd/api/main.go                         # AUTH_JWT_SECRET obligatorio
├── internal/platform/http/router.go        # monta auth; CORS con origen web
├── internal/platform/middleware/           # JWT staff, guest cookie/header
├── internal/auth/                          # login, refresh, logout, me, staff CRUD, guest
├── ent/schema/                             # Staff, StaffRefreshToken, GuestSession (schema only)
└── migrations/                             # staff, refresh, guest, seed dueños, GRANT, RLS

frontend/
├── app/t/[slug]/login/page.tsx
├── app/t/[slug]/hello/page.tsx
├── app/t/[slug]/staff/page.tsx
├── lib/api.ts                              # API_INTERNAL_URL + cookies BFF
└── app/api/auth/*                          # route handlers: login, logout, refresh, guest ensure
```

**Structure Decision:** Mismo monolito que 001. Auth deja de ser `.gitkeep`. Pantallas de prueba bajo `/t/[slug]/…` (funciona sin archivo hosts). Host `demo-a.localhost` sigue siendo autoridad en la API.

## Complexity Tracking

Ninguna violación. El BFF de cookies en Next no es un segundo backend de dominio: solo transporta los tokens del contrato OpenAPI.

## Phase 0 / Phase 1

Hecho: [research.md](research.md), [data-model.md](data-model.md), [contracts/openapi.yaml](contracts/openapi.yaml), [quickstart.md](quickstart.md).
