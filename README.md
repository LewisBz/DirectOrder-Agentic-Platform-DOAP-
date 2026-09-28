# DirectOrder Agentic Platform (DOAP)

Plataforma autónoma y multi-tenant de pedidos directos asistida por agentes de IA para comercios locales e independientes.

La especificación de entrada está en [docs/especificacion-y-arquitectura.md](docs/especificacion-y-arquitectura.md). El trabajo se hace con Spec Kit; Cursor es el agente principal.

## Arranque local

Requisitos: Docker Compose v2. Puertos 5432, 8080 y 3000.

```powershell
copy .env.example .env
git check-ignore -v .env
docker compose down -v
docker compose up --build
```

Espera `migrate` en `Exited (0)`, `api listening` y `web Ready`. El `-v` vuelve a sembrar dueños demo.

Cuando `api` / `web` estén arriba:

```powershell
curl.exe http://localhost:8080/healthz
curl.exe http://localhost:8080/readyz
```

Ambos deben devolver `{"status":"ok"}`.

Swagger (contrato OpenAPI, Try it out): [http://localhost:8080/docs](http://localhost:8080/docs). YAML: [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml). En el navegador el header `Host` no se puede fijar; Swagger manda `X-Forwarded-Host: demo-a.localhost` si lo dejas vacío. Authorize con el `access_token` de `POST /v1/auth/login`.

Para ver comercios por host, añade en el archivo `hosts` de Windows:

```text
127.0.0.1 demo-a.localhost demo-b.localhost
```

Luego `http://demo-a.localhost:3000` o `http://localhost:3000/t/demo-a`.

### Staff (feature 002)

Credenciales semilla (password = `STAFF_SEED_PASSWORD` en `.env`, ejemplo `changeme_staff`):

| Comercio | Correo | Rol |
| --- | --- | --- |
| Demo A | `owner@demo-a.local` | dueño |
| Demo B | `owner@demo-b.local` | dueño |

- Login / saludo: `http://localhost:3000/t/demo-a/login` → `/t/demo-a/hello`
- CRUD caja/cocina: `http://localhost:3000/t/demo-a/staff` (solo dueño)
- Vitrina (invitado automático): `http://localhost:3000/t/demo-a`

Detalle de curls y aislamiento: [specs/002-staff-auth/quickstart.md](specs/002-staff-auth/quickstart.md). Esqueleto de tenant: [specs/001-monolith-scaffold/quickstart.md](specs/001-monolith-scaffold/quickstart.md).

Los tests de Go (incluido RLS) se corren **en el host** (no en la imagen distroless):

```powershell
cd backend
go test ./...
```
