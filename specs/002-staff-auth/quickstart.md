# Quickstart: 002-staff-auth

Prueba login de staff, CRUD de caja/cocina, saludo e invitado. No hay catálogo ni pedidos.

Depende de Compose y comercios demo de `001-monolith-scaffold`. Contrato: [contracts/openapi.yaml](contracts/openapi.yaml). Modelo: [data-model.md](data-model.md).

## Prerrequisitos

- Docker Compose v2, puertos 5432 / 8080 / 3000
- Variables nuevas en `.env` (copia desde `.env.example`): `AUTH_JWT_SECRET` (≥32), `STAFF_SEED_PASSWORD`, `API_INTERNAL_URL` (en Compose, `http://api:8080`)

```powershell
copy .env.example .env
docker compose down -v
docker compose up --build
```

Espera `migrate` `Exited` (éxito), `api listening`, `web Ready`. Recrear con `-v` para sembrar dueños.

## Tenant (regresión del esqueleto)

```powershell
curl.exe http://localhost:8080/healthz
curl.exe -H "Host: demo-a.localhost" http://localhost:8080/v1/tenants/current
curl.exe http://localhost:8080/v1/tenants/current/by-slug/demo-a
```

Los dos current deben devolver Demo A. Si el slug sigue en `tenant_not_found`, `resolve_tenant` aún no tiene `row_security = off`.

## Login y saludo (US1)

Credenciales semilla: `owner@demo-a.local` / valor de `STAFF_SEED_PASSWORD` (ejemplo `changeme_staff`).

```powershell
curl.exe -s -H "Host: demo-a.localhost" -H "Content-Type: application/json" `
  -d "{\"email\":\"owner@demo-a.local\",\"password\":\"changeme_staff\"}" `
  http://localhost:8080/v1/auth/login
```

Guarda `access_token`. `GET /v1/auth/me` con `Authorization: Bearer` y el mismo Host debe devolver nombre, `role: owner`, comercio Demo A.

Mal password → 401 `invalid_credentials` (mensaje genérico).

Pantallas: `http://localhost:3000/t/demo-a/login` → `http://localhost:3000/t/demo-a/hello` (saludo). Logout vuelve al login.

## CRUD (US2)

Con el Bearer del dueño A:

1. `POST /v1/staff` body cashier (email, password ≥8, name, role `cashier`) → 201
2. `GET /v1/staff` → incluye la caja, no a B
3. `PATCH /v1/staff/{id}` sin password (nombre) y luego con password nueva
4. `POST /v1/staff/{id}/deactivate` → login de esa caja falla
5. `POST /v1/staff/{id}/reactivate` → login vuelve
6. Login como caja → `POST /v1/staff` → 403
7. `POST` role `owner` → 422/403

UI: `http://localhost:3000/t/demo-a/staff`

## Aislamiento (US3)

Login dueño B. El Bearer de A en `Host: demo-b.localhost` → 401. `GET /v1/staff` como A no lista a B.

```powershell
cd backend
go test ./internal/auth ./internal/tenant
```

## Invitado (US4)

Abre `http://localhost:3000/t/demo-a` (vitrina). Sin botón extra hay cookie de invitado. `GET /v1/auth/me` o `/v1/staff` con solo guest → 401. `http://localhost:3000/t/demo-a/hello` pide login.

## Recrear

```powershell
docker compose down -v
docker compose up --build
```

Dueños semilla otra vez. Cajas creadas a mano no sobreviven al `-v`.
