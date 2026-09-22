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
