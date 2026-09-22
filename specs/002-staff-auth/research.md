# Research: 002-staff-auth

## 1. Tokens de staff (JWT + refresh rotativo)

**Decision:** Access JWT HS256, **15 minutos**, claims `sub` (staff id), `tenant_id`, `role` (`owner` | `cashier` | `kitchen`), `typ=access`. Refresh opaco (32 bytes random), **7 días**, guardado como SHA-256 en `staff_refresh_tokens`. Cada refresh invalida el hash anterior (rotación). Logout y desactivar/cambiar contraseña revocan todas las filas de refresh de esa persona. Firma con `AUTH_JWT_SECRET` (≥32 caracteres), solo env de API.

**Rationale:** Constitución: JWT corto + refresh rotativo, rol y tenant en el token. Vida fija evita el “NEEDS CLARIFICATION” de duración. El frontend no firma ni guarda el secreto.

**Alternatives considered:** Sesión solo en servidor (más simple, no es el contrato constitucional). Refresh en el JWT (no se puede revocar al desactivar). Access de 24 h (exceso para backoffice).

## 2. Dónde vive el token en el navegador

**Decision:** La API responde JSON (`access_token`, `refresh_token`, `expires_in`). Next App Router, en route handlers del mismo origen que la vitrina, guarda cookies httpOnly (`doap_access`, `doap_refresh`, `doap_guest`) y llama a la API con `Authorization: Bearer`. El servidor de Next usa `API_INTERNAL_URL` (Compose: `http://api:8080`).

**Rationale:** `localhost:3000` y `localhost:8080` no comparten cookies. CORS `*` del esqueleto no sirve con credenciales. El límite de sistema sigue siendo OpenAPI; Next no calcula precios ni roles.

**Alternatives considered:** Cookie Set-Cookie desde la API (rota en local). `localStorage` (XSS lee el JWT). Proxy inverso único (más Compose del que hace falta ahora).

## 3. Hash de contraseña y validación

**Decision:** `bcrypt` costo **12**. Contraseña **mínimo 8** caracteres al crear o al enviar una nueva en PATCH. Correo normalizado a lowercase, formato básico. En PATCH, omitir `password` deja el hash intacto; si viene, se rehash y se revocan refresh.

**Rationale:** Costo 12 es constitucional. Longitud 8 cierra el hueco del clarify. Edición opcional de clave es Q5 de la spec.

**Alternatives considered:** argon2 (válido, no es lo escrito en la constitución). Coste 10 (más rápido, más débil).

## 4. Roles en el cable vs pantallas

**Decision:** API usa `owner`, `cashier`, `kitchen`. UI en español: dueño, caja, cocina. POST/PATCH de staff desde pantallas **solo** `cashier` | `kitchen`. Semilla inserta `owner`. Intento de crear/editar `owner` por API de CRUD → `403`/`422`. Un `owner` no se desactiva ni se cambia de rol en este recorte (único dueño demo).

**Rationale:** Spec Q4. Evita comercio sin dueño sin lógica extra de “último owner”.

**Alternatives considered:** Varios dueños por CRUD (aparcado). Roles en español en JSON (peor para código).

## 5. Lookup de tenant y FORCE RLS

**Decision:** `CREATE FUNCTION resolve_tenant … SET row_security = off` (sigue `SECURITY DEFINER`, owner `doap_owner`). Login: resolver comercio → `SET LOCAL app.tenant_id` → SELECT staff por email. Tablas `staff`, `staff_refresh_tokens`, `guest_sessions` con FORCE RLS `tenant_id = current_setting('app.tenant_id')::uuid`.

**Rationale:** Con FORCE, el definer tampoco ve `tenants`; por eso `/v1/tenants/current/by-slug/demo-a` responde `tenant_not_found` con seed presente. Sin este arreglo el login no arranca.

**Alternatives considered:** Política permisiva en `tenants` para `doap_app` (filtra mal). `BYPASSRLS` en `doap_app` (viola constitución).

## 6. Invitado automático y reutilización

**Decision:** Al renderizar la vitrina de un comercio conocido, Next llama `POST /v1/guest/sessions` (canal `pwa`). Si hay cookie `doap_guest` válida **para ese tenant**, se reutiliza y se actualiza `last_seen`. Si la cookie es de otro tenant o inválida, se emite otra. Header `X-Guest-Token` o cookie hacia la API; **nunca** autoriza rutas `/v1/auth/me` ni `/v1/staff`.

**Rationale:** Clarify Q1–Q2 (emitir ahora, automático). Reutilizar en recargas evita inflar `guest_sessions` y cierra el punto deferred del clarify.

**Alternatives considered:** Nueva fila en cada GET (ruido). Botón “continuar” (rechazado). Aplazar al carrito (rechazado).

## 7. Autorización de peticiones staff

**Decision:** Middleware: Host/slug resuelve tenant; Bearer JWT; `jwt.tenant_id` MUST igualar tenant resuelto; si no, 401. CRUD staff exige `role=owner`. `GET /v1/auth/me` y saludo: cualquier staff activo. Login no lleva Bearer.

**Rationale:** El cliente no manda `tenant_id` de autoridad. Un JWT de A en host de B no es personal de B.

**Alternatives considered:** Confiar solo en el JWT (salta el host). Header `X-Tenant-Id` (prohibido).

## 8. Rate limit de login (local)

**Decision:** 10 intentos / minuto / clave `(ip, tenant_id, email)` en memoria del proceso API. 429 con el mismo mensaje genérico de fallo de acceso cuando se pueda; si no, `rate_limited`. No Redis en este corte.

**Rationale:** Tope razonable del clarify deferred. Un nodo Compose basta.

**Alternatives considered:** Sin tope (abuso trivial en demo). Redis (operación de más).

## 9. Semilla y secretos

**Decision:** `.env.example`: `AUTH_JWT_SECRET`, `STAFF_SEED_PASSWORD=changeme_staff`. Seed SQL: `owner@demo-a.local` y `owner@demo-b.local` (hashes bcrypt generados en `apply.sh` o script one-shot, no un hash hardcodeado obsoleto). Documentar en quickstart. Nunca commitear `.env`.

**Rationale:** FR-008. Dos dueños para el test A-vs-B.

**Alternatives considered:** Password en claro en SQL (inaceptable). Un solo dueño (no prueba aislamiento de login).

## 10. Tests

**Decision:** Contratos HTTP: login ok/genérico, me, logout, refresh rotativo (el refresh viejo falla), CRUD cashier, 403 caja, 409 email duplicado, guest no pasa staff. Aislamiento: SET LOCAL A no ve staff de B; login A + host B → 401. `tsc` en frontend.

**Rationale:** SC-003–SC-006.

**Alternatives considered:** Solo tests de bcrypt (no prueba RLS ni host).
