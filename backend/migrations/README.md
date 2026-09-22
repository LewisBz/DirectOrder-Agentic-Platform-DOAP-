# Atlas SQL for DOAP scaffold

`atlas.hcl` + `schema.sql` are the source of truth.

Local Compose applies the same SQL with `apply.sh` after creating `doap_owner` / `doap_app` from env passwords (Atlas cannot inject those secrets into `CREATE ROLE`).

To apply as owner against a running database:

```bash
atlas schema apply --env local --auto-approve
```

`OWNER_DATABASE_URL` must be a superuser or `doap_owner` URL after roles exist.

Future business tables MUST:

1. Include `tenant_id uuid NOT NULL REFERENCES tenants(id)`
2. `ENABLE` + `FORCE ROW LEVEL SECURITY`
3. Policy: `tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid`
4. `OWNER TO doap_owner` and `GRANT` to `doap_app` only
