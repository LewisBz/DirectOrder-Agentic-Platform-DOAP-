#!/usr/bin/env bash
set -euo pipefail

echo "waiting for postgres..."
until pg_isready -h "${POSTGRES_HOST}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" >/dev/null 2>&1; do
  sleep 1
done

export PGPASSWORD="${POSTGRES_PASSWORD}"

psql -v ON_ERROR_STOP=1 -h "${POSTGRES_HOST}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${POSTGRES_OWNER_USER}') THEN
    EXECUTE format('CREATE ROLE %I LOGIN PASSWORD %L', '${POSTGRES_OWNER_USER}', '${POSTGRES_OWNER_PASSWORD}');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${POSTGRES_APP_USER}') THEN
    EXECUTE format('CREATE ROLE %I LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE', '${POSTGRES_APP_USER}', '${POSTGRES_APP_PASSWORD}');
  END IF;
END
\$\$;
ALTER ROLE ${POSTGRES_APP_USER} NOBYPASSRLS;
GRANT ${POSTGRES_OWNER_USER} TO ${POSTGRES_USER};
SQL

psql -v ON_ERROR_STOP=1 -h "${POSTGRES_HOST}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -f /migrations/schema.sql
psql -v ON_ERROR_STOP=1 -h "${POSTGRES_HOST}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -f /migrations/seed.sql

SEED_PW="${STAFF_SEED_PASSWORD:-changeme_staff}"
SEED_PW_ESC=${SEED_PW//\'/\'\'}
psql -v ON_ERROR_STOP=1 -h "${POSTGRES_HOST}" -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" <<SQL
INSERT INTO staff (id, tenant_id, email, name, role, password_hash, active)
VALUES
  ('33333333-3333-4333-8333-333333333333', '11111111-1111-1111-1111-111111111111', 'owner@demo-a.local', 'Dueño A', 'owner', crypt('${SEED_PW_ESC}', gen_salt('bf', 12)), true),
  ('44444444-4444-4444-8444-444444444444', '22222222-2222-2222-2222-222222222222', 'owner@demo-b.local', 'Dueño B', 'owner', crypt('${SEED_PW_ESC}', gen_salt('bf', 12)), true)
ON CONFLICT (tenant_id, email) DO NOTHING;
SQL

echo "atlas sql applied"
