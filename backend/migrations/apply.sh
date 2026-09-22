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

echo "atlas sql applied"
