# Atlas SQL for DOAP scaffold

`atlas.hcl` + `schema.sql` are the source of truth.

Local Compose applies the same SQL with `apply.sh` after creating `doap_owner` / `doap_app` from env passwords (Atlas cannot inject those secrets into `CREATE ROLE`). The API connects only as `doap_app` (`NOBYPASSRLS`). Table owner is `doap_owner`.

To apply as owner against a running database:

```bash
atlas schema apply --env local --auto-approve
```

`OWNER_DATABASE_URL` must be a superuser or `doap_owner` URL after roles exist.

## Plantilla para tablas de negocio futuras

```sql
CREATE TABLE example (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants (id) ON DELETE CASCADE
    -- columnas de dominio
);

ALTER TABLE example OWNER TO doap_owner;
ALTER TABLE example ENABLE ROW LEVEL SECURITY;
ALTER TABLE example FORCE ROW LEVEL SECURITY;

CREATE POLICY example_isolation ON example
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE example TO doap_app;
```

No conceder estas tablas al owner de migraciones como rol de la API. La transacción de negocio hace `SET LOCAL app.tenant_id`.

## Staff, refresh e invitado (002)

Las tablas `staff`, `staff_refresh_tokens` y `guest_sessions` usan la misma plantilla: `tenant_id`, `OWNER TO doap_owner`, `ENABLE` + `FORCE ROW LEVEL SECURITY`, policy `tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid` en `USING` y `WITH CHECK`, `GRANT` solo a `doap_app`.

Particularidades:

- `staff`: único `(tenant_id, email)` (incluye inactivos). Roles `owner` | `cashier` | `kitchen`. El lookup de comercio (`resolve_tenant`) es `SECURITY DEFINER` con `row_security = off`; el SELECT de staff no.
- `staff_refresh_tokens`: hash SHA-256 del refresh opaco; revocar al rotar, logout, desactivar o cambiar contraseña.
- `guest_sessions`: hash SHA-256 del token de vitrina (`channel = pwa`). Reutilizar en el mismo tenant; otro host emite otra fila. Nunca autoriza `/v1/auth/me` ni `/v1/staff`.

La API abre cada consulta de esas tablas con `WithTenantTx` → `SET LOCAL app.tenant_id`. Sin ese setting, `FORCE RLS` devuelve 0 filas al rol `doap_app`.

`resolve_tenant` es `SECURITY DEFINER` y **no** usa `SET row_security = off` (en Postgres 16 eso exige `BYPASSRLS` y si no está, el lookup revienta con SQLSTATE 42501). Tampoco devuelve `SETOF tenants` (RLS del invocador). Devuelve `TABLE(...)`. Policy `tenants_owner_lookup` (`FOR SELECT TO doap_owner USING (true)`) cubre el SELECT interno bajo `FORCE RLS`. `doap_app` no tiene `BYPASSRLS`.
