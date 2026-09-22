# Data model: 002-staff-auth

## Staff

Tabla `staff`. Fila de negocio. FORCE RLS por `tenant_id`.

| Campo | Tipo | Reglas |
| --- | --- | --- |
| id | UUID PK | Servidor |
| tenant_id | UUID FK → tenants | NOT NULL, ON DELETE CASCADE |
| email | text | lowercase, único por `(tenant_id, email)` |
| name | text | 1–120 |
| role | text | `owner` \| `cashier` \| `kitchen` |
| password_hash | text | bcrypt costo 12; nunca en JSON de respuesta |
| active | boolean | default true; baja lógica |
| created_at, updated_at | timestamptz | UTC |

Unicidad: `(tenant_id, email)` incluye inactivos (no se recrea el mismo correo; se reactiva).

Transiciones: `active true → false` (desactivar, revoca refresh). `false → true` (reactivar; no restaura refresh viejos). En este corte, filas `role=owner` no se crean/editan/desactivan por API.

## StaffRefreshToken

Tabla `staff_refresh_tokens`. FORCE RLS por `tenant_id`.

| Campo | Tipo | Reglas |
| --- | --- | --- |
| id | UUID PK | |
| tenant_id | UUID FK | NOT NULL |
| staff_id | UUID FK → staff | ON DELETE CASCADE |
| token_hash | bytea o text | SHA-256 del refresh opaco; único |
| expires_at | timestamptz | emisión + 7 días |
| revoked_at | timestamptz | null = vigente |
| created_at | timestamptz | |

Rotación: insertar nuevo + `revoked_at=now()` en el usado. Logout / desactivar / password nueva: revocar todos los de `staff_id`.

## GuestSession

Tabla `guest_sessions`. FORCE RLS por `tenant_id`. No es staff.

| Campo | Tipo | Reglas |
| --- | --- | --- |
| id | UUID PK | |
| tenant_id | UUID FK | NOT NULL |
| token_hash | text | SHA-256 del opaco de cookie |
| channel | text | este corte: `pwa` |
| created_at, last_seen | timestamptz | |

Reutilizar si el hash coincide y el tenant de la petición es el mismo. Otro tenant → nueva fila, nueva cookie.

## Relación con Tenant

`tenants` no cambia de columnas. `resolve_tenant` se altera a `SET row_security = off` para que el lookup vuelva a ver filas bajo FORCE RLS.

Tras resolver: `SET LOCAL app.tenant_id`. Todas las consultas de staff/guest/refresh van en esa transacción.

## Semilla

| Comercio | email | role | password (local) |
| --- | --- | --- | --- |
| demo-a | owner@demo-a.local | owner | `STAFF_SEED_PASSWORD` (`changeme_staff`) |
| demo-b | owner@demo-b.local | owner | igual |

No hay caja/cocina de semilla; el dueño las crea en el CRUD de prueba.

## Políticas RLS (plantilla)

```text
USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
WITH CHECK (igual)
```

GRANT a `doap_app`: SELECT/INSERT/UPDATE/DELETE en las tres tablas. Owner de tabla: `doap_owner`.

## Fuera de este modelo

Catálogo, carrito (usará `guest_sessions` después), pedidos, OAuth, reset por correo, admin de plataforma.
