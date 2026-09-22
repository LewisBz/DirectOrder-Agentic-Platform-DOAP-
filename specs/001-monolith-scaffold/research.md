# Research: 001-monolith-scaffold

## 1. Resolución de comercio (host vs slug)

**Decision:** El Host de la petición (sin puerto) es la autoridad. Si no hay coincidencia, se intenta el slug en la ruta `/t/{slug}`. Si Host y slug contradicen, gana el Host. Se ignoran `X-Tenant-Id`, `tenant_id` en query y cuerpo.

**Rationale:** FR-005 y la constitución prohíben un header inventable. En Compose local, `demo-a.localhost` y `demo-b.localhost` apuntan a la vitrina; la API recibe `Host` o `X-Forwarded-Host` puesto por el proxy de desarrollo. El fallback `/t/{slug}` cubre quien no pueda mapear hosts.

**Alternatives considered:** Solo slug en path (más simple, pero el producto pide host o slug y el host gana). Header `X-Tenant-Slug` (fácil de falsificar si algún día hay un proxy mal configurado).

## 2. RLS, roles y Ent

**Decision:** Dos roles de Postgres: `doap_owner` (migraciones, dueño de tablas) y `doap_app` (la API, `NOBYPASSRLS`). Tablas de negocio con `ENABLE` + `FORCE ROW LEVEL SECURITY`. Política: `tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid`. La API abre cada transacción de negocio con `SET LOCAL app.tenant_id`. La resolución de comercio usa una función `resolve_tenant(host text, slug text)` `SECURITY DEFINER` owned by `doap_owner`, para no leer `tenants` sin contexto. Ent modela entidades; Atlas aplica el SQL de roles, GRANTs, función y políticas (Ent no genera FORCE RLS).

**Rationale:** El owner ignora RLS salvo FORCE. El rol de aplicación no debe ser owner. SET LOCAL no sobrevive al commit, así no se filtra el tenant entre peticiones del pool.

**Alternatives considered:** Un solo rol postgres (viola constitución). RLS solo en software (un `WHERE` olvidado filtra mal). GORM (prohibido).

## 3. Tabla canario de aislamiento

**Decision:** Tabla `isolation_canaries` (`id`, `tenant_id`, `label`) semilla una fila por comercio de ejemplo. No es producto; existe para que el test A-vs-B tenga una fila de negocio distinta de `tenants`.

**Rationale:** `tenants` se resuelve con SECURITY DEFINER; no demuestra RLS de filas de negocio.

**Alternatives considered:** Probar RLS solo sobre `tenants` (falso positivo). Crear catálogo ahora (fuera de alcance).

## 4. Imagen de la API y certificados

**Decision:** Multi-stage: `golang:1.24` compile estático → runtime `gcr.io/distroless/base-debian12` (incluye CA). No `scratch` vacío.

**Rationale:** FR-011 y constitución: TLS de salida hacia JEV AI más adelante.

**Alternatives considered:** `scratch` + copiar `ca-certificates.crt` (válido, más frágil). Alpine (más superficie).

## 5. Compose y migraciones al arrancar

**Decision:** `docker-compose.yml` con `postgres:16`, `api`, `web`. Un servicio `migrate` (mismo Dockerfile de API, entrypoint Atlas) corre antes de `api` (`depends_on` + healthcheck de Postgres). Semilla SQL idempotente de `demo-a` y `demo-b` + canarios.

**Rationale:** Un solo comando documentado cumple SC-001. La API no corre migraciones con el rol `doap_app`.

**Alternatives considered:** Migrar a mano (rompe el quickstart). Auto-migrate de Ent en el proceso API (mezcla owner y app).

## 6. Vitrina mínima

**Decision:** Next.js App Router (TypeScript, Tailwind) con una página que muestra el comercio actual vía `GET /v1/tenants/current` usando Host. Sin login, chat ni PWA worker agresivo. Carpetas `src/mcp-apps` y `src/webmcp` vacías con `.gitkeep`.

**Rationale:** FR-001 pide vitrina viva; FR-012 prohíbe pedido y agente.

**Alternatives considered:** Solo un nginx estático (no es el stack mandado). App completa de chat (fuera de alcance).

## 7. Tests

**Decision:** Tests de integración Go contra Postgres de Compose (o Testcontainers si Compose no está). Un test de contrato HTTP contra el OpenAPI (healthz, current tenant, 404, header `X-Tenant-Id` ignorado). Un test de RLS: rol `doap_app`, SET LOCAL A, SELECT canario B → 0 filas; intento UPDATE → 0.

**Rationale:** SC-002 exige batería automatizada, no una demo manual.

**Alternatives considered:** Solo tests unitarios del resolver (no prueba FORCE RLS).

## 8. Variables de entorno

**Decision:** `.env.example` con `POSTGRES_OWNER_USER`, `POSTGRES_OWNER_PASSWORD`, `POSTGRES_APP_USER`, `POSTGRES_APP_PASSWORD`, `POSTGRES_DB`, `POSTGRES_HOST`, `API_ADDR`, `NEXT_PUBLIC_API_URL`. Valores de ejemplo solo para local (`changeme`), nunca producción. `.env` ignorado.

**Rationale:** FR-002. Dos usuarios distintos son el aislamiento.

**Alternatives considered:** Una sola `DATABASE_URL` (el migrador y la API compartirían owner).
