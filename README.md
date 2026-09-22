# DirectOrder Agentic Platform (DOAP)

Plataforma autónoma y multi-tenant de pedidos directos asistida por agentes de IA para comercios locales e independientes.

La especificación de entrada está en [docs/especificacion-y-arquitectura.md](docs/especificacion-y-arquitectura.md). El trabajo se hace con Spec Kit; Cursor es el agente principal.

## Arranque local (esqueleto)

Requisitos: Docker Compose v2. Puertos 5432, 8080 y 3000.

```powershell
copy .env.example .env
git check-ignore -v .env
docker compose up --build
```

Cuando `migrate` termine y `api` / `web` estén arriba:

```powershell
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

Ambos deben devolver `{"status":"ok"}`.

Para ver comercios por host, añade en el archivo `hosts` de Windows:

```text
127.0.0.1 demo-a.localhost demo-b.localhost
```

Luego `http://demo-a.localhost:3000` o `http://localhost:3000/t/demo-a`.

Los tests de Go (incluido RLS) se corren **en el host**:

```powershell
cd backend
go test ./...
```

Detalle: [specs/001-monolith-scaffold/quickstart.md](specs/001-monolith-scaffold/quickstart.md).
