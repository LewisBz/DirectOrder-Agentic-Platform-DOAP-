# Data model: 001-monolith-scaffold

## Tenant (comercio)

Tabla `tenants`. No es una fila “de negocio” filtrada por sí misma: se busca con `resolve_tenant`. Tras resolver, `SET LOCAL app.tenant_id`.

| Campo | Tipo | Reglas |
| --- | --- | --- |
| id | UUID PK | Generado en servidor |
| slug | text | Único, `[a-z0-9-]{2,63}`, no empieza ni termina en `-` |
| host | text | Único, hostname sin puerto ni esquema, lowercase |
| name | text | 1–120 caracteres |
| currency | char(3) | Default `COP` |
| tax_name | text | Default `IVA` |
| timezone | text | IANA, default `America/Bogota` |
| opening_hours | jsonb | Default `{}`; forma libre en este corte |
| created_at | timestamptz | UTC |
| updated_at | timestamptz | UTC |

Unicidad: `slug`, `host`. Insertar duplicado → conflicto.

RLS en `tenants` (rol `doap_app`): SELECT/UPDATE solo `id = current_setting('app.tenant_id')::uuid`. INSERT de comercios nuevos en este corte solo por semilla/migración (owner), no por API pública.

## IsolationCanary

Tabla `isolation_canaries`. Fila de negocio para demostrar RLS. No se expone en la vitrina.

| Campo | Tipo | Reglas |
| --- | --- | --- |
| id | UUID PK | |
| tenant_id | UUID FK → tenants | NOT NULL, ON DELETE CASCADE |
| label | text | Único por tenant |

RLS: `tenant_id = current_setting('app.tenant_id')::uuid` con FORCE. Sin `app.tenant_id`, cero filas.

## Roles Postgres

| Rol | Uso |
| --- | --- |
| `doap_owner` | Dueño de tablas, funciones definer, migraciones Atlas |
| `doap_app` | LOGIN, `NOBYPASSRLS`, GRANT SELECT/INSERT/UPDATE/DELETE sobre tablas de negocio según política |

La API usa solo `doap_app`. Atlas/Compose init usa `doap_owner` (o superuser de bootstrap que crea ambos roles y luego cede ownership).

## Función `resolve_tenant(p_host text, p_slug text)`

`SECURITY DEFINER`, `SET search_path = public`. Devuelve una fila de `tenants` o ninguna.

Orden: si `p_host` no es vacío, busca `host = lower(p_host)` (host sin puerto). Si hay fila, la devuelve aunque `p_slug` diga otra cosa. Si no, busca `slug = lower(p_slug)` si viene.

## Semilla

| slug | host | name |
| --- | --- | --- |
| demo-a | demo-a.localhost | Demo A |
| demo-b | demo-b.localhost | Demo B |

Ambos: COP, IVA, America/Bogota. Un canario `label = seed` por comercio.

## Fuera de este modelo

Productos, usuarios, carritos, órdenes, pagos, sesiones. Las tablas futuras MUST incluir `tenant_id` y la misma política FORCE RLS (plantilla documentada en migraciones).
