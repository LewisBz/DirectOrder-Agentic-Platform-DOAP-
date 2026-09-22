CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS tenants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text NOT NULL UNIQUE,
    host text NOT NULL UNIQUE,
    name text NOT NULL,
    currency character(3) NOT NULL DEFAULT 'COP',
    tax_name text NOT NULL DEFAULT 'IVA',
    timezone text NOT NULL DEFAULT 'America/Bogota',
    opening_hours jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT tenants_slug_format CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$'),
    CONSTRAINT tenants_name_len CHECK (char_length(name) BETWEEN 1 AND 120)
);

CREATE TABLE IF NOT EXISTS isolation_canaries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    label text NOT NULL,
    UNIQUE (tenant_id, label)
);

ALTER TABLE tenants OWNER TO doap_owner;
ALTER TABLE isolation_canaries OWNER TO doap_owner;

ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenants FORCE ROW LEVEL SECURITY;
ALTER TABLE isolation_canaries ENABLE ROW LEVEL SECURITY;
ALTER TABLE isolation_canaries FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenants_isolation ON tenants;
CREATE POLICY tenants_isolation ON tenants
    USING (id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

DROP POLICY IF EXISTS isolation_canaries_isolation ON isolation_canaries;
CREATE POLICY isolation_canaries_isolation ON isolation_canaries
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

CREATE OR REPLACE FUNCTION resolve_tenant(p_host text, p_slug text)
RETURNS SETOF tenants
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    h text := nullif(lower(split_part(btrim(coalesce(p_host, '')), ':', 1)), '');
    s text := nullif(lower(btrim(coalesce(p_slug, ''))), '');
BEGIN
    IF h IS NOT NULL THEN
        RETURN QUERY SELECT * FROM tenants WHERE host = h;
        IF FOUND THEN
            RETURN;
        END IF;
    END IF;
    IF s IS NOT NULL THEN
        RETURN QUERY SELECT * FROM tenants WHERE slug = s;
    END IF;
END;
$$;

ALTER FUNCTION resolve_tenant(text, text) OWNER TO doap_owner;

REVOKE ALL ON FUNCTION resolve_tenant(text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION resolve_tenant(text, text) TO doap_app;

GRANT SELECT, UPDATE ON TABLE tenants TO doap_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE isolation_canaries TO doap_app;
