INSERT INTO tenants (id, slug, host, name, currency, tax_name, timezone, opening_hours)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'demo-a', 'demo-a.localhost', 'Demo A', 'COP', 'IVA', 'America/Bogota', '{}'::jsonb),
  ('22222222-2222-2222-2222-222222222222', 'demo-b', 'demo-b.localhost', 'Demo B', 'COP', 'IVA', 'America/Bogota', '{}'::jsonb)
ON CONFLICT (slug) DO UPDATE
SET host = EXCLUDED.host,
    name = EXCLUDED.name,
    currency = EXCLUDED.currency,
    tax_name = EXCLUDED.tax_name,
    timezone = EXCLUDED.timezone,
    updated_at = now();
