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
SET row_security = off
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

CREATE TABLE IF NOT EXISTS staff (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    email text NOT NULL,
    name text NOT NULL,
    role text NOT NULL,
    password_hash text NOT NULL,
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT staff_email_lower CHECK (email = lower(email)),
    CONSTRAINT staff_name_len CHECK (char_length(name) BETWEEN 1 AND 120),
    CONSTRAINT staff_role_ok CHECK (role IN ('owner', 'cashier', 'kitchen')),
    UNIQUE (tenant_id, email)
);

CREATE TABLE IF NOT EXISTS staff_refresh_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    staff_id uuid NOT NULL REFERENCES staff (id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS guest_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    channel text NOT NULL DEFAULT 'pwa',
    created_at timestamptz NOT NULL DEFAULT now(),
    last_seen timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT guest_channel_ok CHECK (channel IN ('pwa'))
);

ALTER TABLE staff OWNER TO doap_owner;
ALTER TABLE staff_refresh_tokens OWNER TO doap_owner;
ALTER TABLE guest_sessions OWNER TO doap_owner;

ALTER TABLE staff ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff FORCE ROW LEVEL SECURITY;
ALTER TABLE staff_refresh_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_refresh_tokens FORCE ROW LEVEL SECURITY;
ALTER TABLE guest_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE guest_sessions FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS staff_isolation ON staff;
CREATE POLICY staff_isolation ON staff
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

DROP POLICY IF EXISTS staff_refresh_tokens_isolation ON staff_refresh_tokens;
CREATE POLICY staff_refresh_tokens_isolation ON staff_refresh_tokens
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

DROP POLICY IF EXISTS guest_sessions_isolation ON guest_sessions;
CREATE POLICY guest_sessions_isolation ON guest_sessions
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE staff TO doap_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE staff_refresh_tokens TO doap_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE guest_sessions TO doap_app;

