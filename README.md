# DirectOrder Agentic Platform (DOAP)

Plataforma autónoma y multi-tenant de pedidos directos asistida por agentes de IA para comercios locales e independientes.

La especificación de entrada está en [docs/especificacion-y-arquitectura.md](docs/especificacion-y-arquitectura.md). El trabajo se hace con Spec Kit; Cursor es el agente principal.

## Arranque local (esqueleto)

Requisitos: Docker Compose v2. Puertos 5432, 8080 y 3000.

```powershell
copy .env.example .env
docker compose up --build
```

Cuando `migrate` termine y `api` / `web` estén arriba:

```powershell
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

Ambos deben devolver `{"status":"ok"}`. La vitrina está en `http://localhost:3000`.

Los tests de Go se corren **en el host** (la imagen `api` es distroless y no trae el compilador):

```powershell
cd backend
go test ./...
```

Detalle: [specs/001-monolith-scaffold/quickstart.md](specs/001-monolith-scaffold/quickstart.md).
