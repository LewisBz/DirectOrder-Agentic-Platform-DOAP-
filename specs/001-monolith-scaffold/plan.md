# Implementation Plan: Esqueleto del monolito multi-tenant

**Branch**: `001-monolith-scaffold` | **Date**: 2026-09-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-monolith-scaffold/spec.md`

## Summary

Levantar el árbol del monolito modular, Compose (Postgres + API + vitrina), secretos fuera de git, esquema de comercios con RLS forzado y rol de aplicación distinto del owner, más el módulo `tenant` (resolver por Host, fallback slug). Auth, catálogo, carrito, pedidos, pagos y agente quedan como carpetas vacías.

## Technical Context

**Language/Version**: Go 1.24 (API); TypeScript (Next.js App Router) en la vitrina

**Primary Dependencies**: Ent, Atlas, pgx, go-playground/validator/v10, chi o net/http; Next.js + Tailwind. GORM no. Astro no.

**Storage**: PostgreSQL 16. Roles `doap_owner` (migraciones) y `doap_app` (API, NOBYPASSRLS). FORCE RLS + `SET LOCAL app.tenant_id`.

**Testing**: `go test` de integración contra Postgres (RLS A-vs-B); comprobaciones HTTP del contrato OpenAPI; `tsc` en la vitrina.

**Target Platform**: desarrollo local en contenedores (Windows host + Linux containers). No producción.

**Project Type**: web (backend modular + frontend)

**Performance Goals**: señal de vida y `GET /v1/tenants/current` usables en local; no hay SLO de 15 ms en este corte.

**Constraints**: secretos fuera de git; imagen API con CA certs (`distroless/base`, no scratch vacío); el cliente no es autoridad del tenant; OpenAPI es el límite.

**Scale/Scope**: dos comercios semilla; una tabla canario; vitrina de una página; cero dominio de pedidos.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Este corte | Estado |
| --- | --- | --- |
| I Comercio dueño / Colombia | Defaults COP, IVA, America/Bogota en entidad y semilla | Pass |
| II API HTTP es el límite | [contracts/openapi.yaml](contracts/openapi.yaml); sin WebMCP ni MCP Apps | Pass |
| III Modelo sin autoridad de dinero | Sin clasificador ni precios | Pass (N/A) |
| IV Aislamiento comprobable | FORCE RLS, `doap_app` ≠ owner, SET LOCAL, test A-vs-B, resolve por host/slug | Pass |
| V Pedido atómico | Fuera de alcance | Pass (N/A) |
| Stack Ent/Atlas/pgx + Next + CA | research.md | Pass |
| Código empieza en tenant | Plataforma + `tenant` reales; resto esqueleto | Pass |
| Pedidos/WhatsApp/Mercado Pago | Explicitamente no | Pass |

Post-diseño (Phase 1): la tabla `isolation_canaries` no es un módulo extra de producto; es el instrumento del test de IV. La función `SECURITY DEFINER` es el único camino para resolver el comercio sin abrir RLS. Sin violaciones que rellenar en Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/001-monolith-scaffold/
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
├── cmd/api/main.go
├── cmd/worker/main.go          # hueco: log "not implemented" y exit 0 o 1 documentado
├── internal/platform/config/
├── internal/platform/database/ # pool pgx, Tx con SET LOCAL
├── internal/platform/middleware/
├── internal/platform/validation/
├── internal/tenant/
├── internal/auth/              # .gitkeep
├── internal/catalog/
├── internal/cart/
├── internal/orders/
├── internal/payments/
├── internal/agent/
├── internal/channel/
├── ent/                        # schema Tenant, IsolationCanary
├── migrations/                 # Atlas: roles, FORCE RLS, resolve_tenant, seed
├── Dockerfile
├── go.mod
└── go.sum

frontend/
├── app/                        # página / y /t/[slug]
├── src/components/
├── src/mcp-apps/               # .gitkeep
├── src/webmcp/                 # .gitkeep
├── public/manifest.webmanifest
├── Dockerfile
└── package.json

docker-compose.yml
.env.example
README.md
docs/especificacion-y-arquitectura.md
```

**Structure Decision:** Árbol de [docs/especificacion-y-arquitectura.md](../../docs/especificacion-y-arquitectura.md) sección 5, más `backend/ent` (Ent exige su paquete) y servicio Compose `migrate`.

## Complexity Tracking

Ninguna violación de constitución. `isolation_canaries` y `resolve_tenant` SECURITY DEFINER son el diseño mínimo para cumplir el gate IV sin implementar catálogo.

## Phase 0 / Phase 1

Hecho: [research.md](research.md), [data-model.md](data-model.md), [contracts/openapi.yaml](contracts/openapi.yaml), [quickstart.md](quickstart.md).
