# Quickstart: 001-monolith-scaffold

Prueba el esqueleto local. No implementa pedidos ni login.

## Prerrequisitos

- Docker Compose v2
- Puerto 5432, 8080 y 3000 libres (o ajusta `.env`)

## Arranque

```powershell
copy .env.example .env
git check-ignore -v .env
docker compose up --build
```

`git check-ignore` debe mencionar `.gitignore`. Espera a que `migrate` termine, `api` pase `/readyz` y `web` sirva `/`.

## Señales de vida

```powershell
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

Ambos deben devolver `{"status":"ok"}`.

## Comercio por host (autoridad)

En `C:\Windows\System32\drivers\etc\hosts` (como administrador):

```text
127.0.0.1 demo-a.localhost demo-b.localhost
```

```powershell
curl -H "Host: demo-a.localhost" http://localhost:8080/v1/tenants/current
curl -H "Host: demo-b.localhost" http://localhost:8080/v1/tenants/current
```

Cada respuesta trae `slug` `demo-a` o `demo-b`, `currency` COP, `tax_name` IVA, `timezone` America/Bogota.

## El id inventado no manda

```powershell
curl -H "Host: demo-a.localhost" -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000099" http://localhost:8080/v1/tenants/current
```

Sigue siendo Demo A. Host desconocido → 404 `tenant_not_found`.

## Fallback por slug

```powershell
curl http://localhost:8080/v1/tenants/current/by-slug/demo-a
```

## Vitrina

Abre `http://localhost:3000/t/demo-a` (o `http://demo-a.localhost:3000` si añadiste hosts). Debe mostrar el nombre del comercio. `http://localhost:3000` sin host de demo muestra el aviso de host desconocido: no hay catálogo, carrito ni login en este esqueleto.

El log de Next `Local: http://<id>:3000` es el hostname interno del contenedor. En el navegador del host usa `http://localhost:3000`. `migrate` en `Exited` es correcto: aplica schema/seed y termina. En Compose, Next llama a la API con `API_INTERNAL_URL=http://api:8080` (no `localhost` dentro de `web`). Recrea `web` si cambias esa variable.

## Aislamiento

Desde el **host** (la imagen `api` es distroless; no ejecuta `go test`):

```powershell
cd backend
go test ./...
```

Con Compose arriba, `TestIsolation*` usa `doap_app` en localhost:5432. Contexto A no lee ni actualiza canarios de B. Sin `app.tenant_id`, `isolation_canaries` devuelve 0 filas. El rol de la API no es dueño de las tablas.

## Recrear

```powershell
docker compose down -v
docker compose up --build
curl -H "Host: demo-a.localhost" http://localhost:8080/v1/tenants/current
curl -H "Host: demo-b.localhost" http://localhost:8080/v1/tenants/current
```

Semilla otra vez: Demo A, Demo B, `/readyz` ok, un canario `seed` por comercio.

## Secretos

`.env` no se commitea. Si falta una variable obligatoria, `api` no arranca y el log nombra la clave. Contrato: [contracts/openapi.yaml](contracts/openapi.yaml). Modelo: [data-model.md](data-model.md).
