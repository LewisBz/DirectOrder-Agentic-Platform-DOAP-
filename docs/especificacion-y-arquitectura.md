# DirectOrder Agentic Platform (DOAP)

> Plataforma autónoma y multi-tenant de pedidos directos asistida por agentes de IA para comercios locales e independientes.

Este documento es el **punto de entrada** del producto. Las specs de feature de Spec Kit lo refinan; no lo contradicen. Si un plan de implementación discrepa de este archivo, manda este archivo.

---

## 1. Contexto y problemática

### 1.1 El dilema de los comercios locales

Los restaurantes, cadenas de comida rápida y tiendas locales se enfrentan a un compromiso insostenible:

- **Dependencia de agregadores masivos (ej. Rappi):** comisiones del 20 % al 30 % por orden, pérdida de la titularidad sobre datos y clientes, y nulo control de marca.
- **Canales propios saturados (WhatsApp manual):** empleados que responden mensajes no estructurados, con demoras de 20 a 40 minutos en horas pico, errores de personalización y abandono de carritos.

### 1.2 La solución: comercio directo agéntico

El proyecto no es un marketplace que compite en flota de repartidores. Es **infraestructura de software soberana** para que el comercio reciba, confirme, cobre y cumpla pedidos en su propio canal.

1. Entiende pedidos en lenguaje natural mediante **JEV AI**. El modelo solo propone; no fija precio, stock ni total.
2. Evita pedidos erróneos con widgets interactivos (**MCP Apps**, evolución de MCP-UI) cuando el pedido es ambiguo o sensible.
3. Expone las mismas operaciones de dominio en el navegador mediante **WebMCP** (`document.modelContext.registerTool`), con detección de soporte. WebMCP es un Draft Community Group Report, no un estándar W3C cerrado.
4. Ejecuta el negocio en un **monolito modular multi-tenant en Go**. La **API HTTP con OpenAPI** es el límite del sistema. WebMCP y MCP Apps son adaptadores.

### 1.3 Primer mercado

Colombia. Defectos del comercio: moneda **COP**, impuesto **IVA**, zona horaria **`America/Bogota`**. El servidor guarda instantes en UTC. El primer corte cobra en **efectivo**. **Mercado Pago** es la primera pasarela futura detrás de la interfaz de pagos; no se implementa en el primer corte.

---

## 2. Pila tecnológica

### Backend

- **Lenguaje:** Go 1.23+ / 1.24+.
- **Arquitectura:** monolito modular multi-tenant.
- **ORM y modelado:** Ent. El esquema Ent es el contrato. GORM no.
- **Driver:** pgx.
- **Validación de entradas:** `go-playground/validator/v10`.
- **Migraciones:** Atlas CLI (`ariga.io/atlas`).
- **Contrato de API:** OpenAPI.
- **Seguridad de staff:** bcrypt costo 12, JWT de vida corta más refresh rotativo, RBAC.

### Frontend y móvil

- **Framework:** Next.js (App Router) + TypeScript + Tailwind CSS.
- **Chat, widgets MCP Apps y registro WebMCP:** client components de esa app.
- **PWA:** `manifest.webmanifest` y service worker compatible con Next (Serwist o equivalente). El service worker no cachea respuestas autenticadas ni reenvía un pedido viejo sin red.
- **Empaquetado nativo (después del núcleo):** Trusted Web Activities (TWA / Bubblewrap) o Capacitor.

Astro y Vite 8.3 no forman parte de la pila. Next.js trae su propio bundler.

### Inteligencia y protocolos

- **JEV AI:** clasificador y orquestador de flujos de baja latencia. Puerto inyectable en el módulo `agent`. No es fuente de verdad de catálogo ni de dinero.
- **MCP Apps:** recursos `ui://` renderizados en iframe sandboxed para modificar, confirmar y aclarar el pedido.
- **WebMCP:** registro de herramientas en `document.modelContext`. Feature detection obligatoria. Las mismas operaciones existen en la API HTTP.

### Infraestructura y metodología

- **Contenedores:** Docker multi-stage para Go con certificados CA (la imagen puede llamar a JEV AI por TLS). Docker Compose para desarrollo: PostgreSQL + backend + frontend.
- **Metodología:** Spec-Driven Development con Spec Kit. Agente principal: **Cursor**. OpenCode queda instalado al lado; no es el agente por defecto.

---

## 3. Arquitectura del sistema

WebMCP y MCP Apps no son el backend. La API del monolito es el límite.

```
[ Canales: PWA ahora ; mensajería después ]
                    │
                    ▼
          [ Chat Next.js ]
           /            \
          ▼              ▼
 [ JEV AI no confiable ]  [ MCP Apps: solo UI ]
          │
          ▼
 [ Resolución contra catálogo ]
          │
          ▼
 [ Herramientas de dominio ]
          │
     ┌────┴────┐
     ▼         ▼
 [ API Go ]  [ WebMCP en client components ]
     │            │
     └─────┬──────┘
           ▼
 [ PostgreSQL + FORCE ROW LEVEL SECURITY ]
```

### 3.1 Frontera de confianza

| Superficie | Puede | No puede |
| --- | --- | --- |
| JEV AI | Proponer ítems, modificadores e intención | Fijar precio, IVA, SKU, stock o total |
| Widget MCP Apps | Recoger elección de la persona | Saltar la API ni mutar stock por su cuenta |
| WebMCP en el DOM | Invocar las mismas operaciones que la API, con confirmación humana en `submitOrder` | Ser la única vía de pedido |
| API HTTP | Calcular total, reservar stock, cobrar, transicionar estados | Confiar en totales enviados por el cliente o el modelo |

### 3.2 Identidades

Dos mundos distintos:

- **Invitado (storefront):** sesión de pedido sin cuenta. Identifica el carrito y el canal. No lleva rol de staff ni JWT de empleado.
- **Staff del comercio:** dueño, caja, cocina. Autenticación con bcrypt y JWT que lleva rol y `tenant_id`.
- **Admin de plataforma:** no lee todos los tenants por defecto. Cualquier acceso cross-tenant es explícito, auditado y fuera del primer corte.

### 3.3 Resolución de tenant

El tenant se resuelve por **host o slug** (subdominio o ruta del comercio). No basta un header que el cliente pueda falsificar. El middleware fija el tenant de la transacción; las consultas no aceptan `tenant_id` del cuerpo como autoridad.

### 3.4 Aislamiento en PostgreSQL

- Discriminador `tenant_id` en todas las filas de negocio.
- Políticas RLS con **`FORCE ROW LEVEL SECURITY`**.
- Rol de aplicación distinto del owner de las tablas. El owner ignoraría RLS sin `FORCE`.
- `SET LOCAL` del tenant dentro de la transacción (por ejemplo `app.tenant_id`).
- Tests obligatorios: el tenant A no lee ni muta filas del tenant B.

### 3.5 Puerto de canal

El núcleo de pedidos no conoce WhatsApp ni el DOM. Conoce un **puerto de canal**: mensaje de entrada, identidad de canal, y respuesta (texto, widget o tool call).

- Primer corte: adaptador PWA / chat web.
- Después, sin reescribir pedidos: adaptador WhatsApp.

---

## 4. Módulos y orden de construcción

Primer corte, en este orden:

1. `tenant` — comercios, slug/host, moneda, IVA, zona horaria, horario del local.
2. `auth` — staff (dueño, caja, cocina) y sesión de invitado.
3. `catalog` — productos, grupos de modificadores (mínimo/máximo), disponibilidad, precios en COP.
4. `cart` — carrito por sesión de invitado, recálculo de totales en servidor.
5. `orders` — pedido, máquina de estados, entrega o recoger en tienda, bitácora de transiciones.
6. `payments` — efectivo ahora; interfaz de proveedor; Mercado Pago después.
7. `agent` — puerto del clasificador, resolución contra catálogo, umbral de confianza. No registra herramientas en el DOM y no sirve de servidor MCP UI.

Después del núcleo, sin reescribirlo: canal WhatsApp, consola de cocina, notificaciones, Mercado Pago, fotos en object storage, jobs (`cmd/worker` además de `cmd/api`), TWA/APK.

### 4.1 Máquina de estados de la orden

Estados:

- `DRAFT` — carrito aún no enviado.
- `PENDING_PAYMENT` — enviado, esperando cobro (efectivo al recibir o al recoger).
- `CONFIRMED` — el comercio aceptó.
- `IN_KITCHEN`
- `READY_FOR_PICKUP` — recoger en tienda.
- `DISPATCHED` — envío a domicilio. No aplica a recoger en tienda.
- `DELIVERED`
- `CANCELLED`
- `REJECTED`
- `PAYMENT_FAILED`

Cada transición declara **actor permitido** (invitado, caja, cocina, dueño, sistema). El invitado no marca `IN_KITCHEN` ni `DELIVERED`. Caja o dueño confirman y cobran. Cocina avanza cocina y listo. El sistema puede expirar un carrito o un pago.

### 4.2 Idempotencia y stock

`submitOrder` exige **llave de idempotencia**. Reintentos del agente, doble toque y el service worker no crean dos pedidos. Reserva de stock y creación de la orden ocurren en la **misma transacción** ACID. Si el stock no alcanza, no hay orden.

### 4.3 Dinero

El servidor calcula precio unitario, modificadores, IVA y total a partir del catálogo vigente. Cualquier total, precio o SKU que llegue del modelo o del navegador se ignora como autoridad. Si la confianza del clasificador está por debajo del umbral, se muestra un widget o se escala a una persona.

---

## 5. Estructura de directorios (objetivo)

El árbol es el destino de la implementación. Spec Kit no lo genera. Hasta el primer módulo, el repo contiene metodología y esta especificación.

```
.
├── backend/
│   ├── cmd/
│   │   ├── api/main.go
│   │   └── worker/main.go          # timeouts de carrito, reintentos; no en el primer binario
│   ├── internal/
│   │   ├── platform/
│   │   │   ├── config/
│   │   │   ├── database/           # pgx, Ent, SET LOCAL tenant
│   │   │   ├── middleware/         # TenantResolver por host/slug, JWT staff, CORS, rate limit
│   │   │   └── validation/
│   │   ├── tenant/
│   │   ├── auth/
│   │   ├── catalog/
│   │   ├── cart/
│   │   ├── orders/
│   │   ├── payments/
│   │   ├── agent/                  # puerto clasificador + resolución; no DOM
│   │   └── channel/                # puerto de canal; adaptador web primero
│   ├── migrations/                 # Atlas
│   ├── Dockerfile                  # Go estático + CA certs
│   ├── go.mod
│   └── go.sum
├── frontend/                       # Next.js App Router
│   ├── app/
│   ├── src/
│   │   ├── components/
│   │   ├── mcp-apps/               # widgets ui://
│   │   └── webmcp/                 # document.modelContext.registerTool
│   ├── public/manifest.webmanifest
│   └── Dockerfile
├── docs/
│   └── especificacion-y-arquitectura.md
├── docker-compose.yml
└── .gitignore
```

---

## 6. Requisitos funcionales

- **RF-01:** Autenticación de staff con bcrypt costo 12 y JWT de vida corta que porta rol y `tenant_id`. Refresh rotativo.
- **RF-02:** Sesión de invitado para el storefront, independiente del staff.
- **RF-03:** Aislamiento multi-tenant por host o slug, RLS forzado, rol de aplicación no-owner, `SET LOCAL` por transacción.
- **RF-04:** Validación de DTOs con `validator/v10` antes de cualquier mutación.
- **RF-05:** Clasificación y extracción de entidades con JEV AI. El resultado se resuelve contra el catálogo antes de mutar.
- **RF-06:** Widgets MCP Apps cuando el pedido tiene campos variables o la confianza es baja (modificadores, horario, entrega vs recoger).
- **RF-07:** Registro WebMCP en `document.modelContext` con feature detection. `submitOrder` exige confirmación explícita de la persona. Las mismas herramientas existen en la API HTTP.
- **RF-08:** Carrito, recálculo de totales en servidor, IVA COP, entrega o recoger en tienda.
- **RF-09:** `submitOrder` idempotente; reserva de stock en la misma transacción.
- **RF-10:** Máquina de estados de la sección 4.1, con actor por transición.
- **RF-11:** Pago en efectivo en el primer corte; interfaz de proveedor lista para Mercado Pago.
- **RF-12:** Puerto de canal desacoplado del browser.

---

## 7. Requisitos no funcionales

- **RNF-01:** SLO separados. Lecturas de catálogo calientes frente a turnos del agente. No hay un p99 único de 15 ms para todo el backend.
- **RNF-02:** Cero secretos de base de datos o de proveedores en el frontend o en git.
- **RNF-03:** Transacciones ACID en PostgreSQL para stock y creación de orden.
- **RNF-04:** Lighthouse ≥ 90 en móviles en el storefront, sin cachear autenticación en el service worker.
- **RNF-05:** Imagen Docker del backend con CA certs para TLS de salida. No se usa `scratch` vacío.
- **RNF-06:** Rate limit y tope de costo en el endpoint de lenguaje natural.
- **RNF-07:** Tests de aislamiento de tenant, contratos OpenAPI y contratos de herramientas.
- **RNF-08:** Bitácora de transiciones de orden por tenant, sin volcar PII al log.
- **RNF-09:** Datos personales (teléfono, dirección) con retención del pedido acorde a la operación en Colombia. El detalle legal se especifica en una feature posterior; el modelo de datos no asume borrado inmediato del pedido cumplido.

---

## 8. Fuera del primer corte

Se nombra para no acoplar el núcleo:

- Adaptador WhatsApp sobre el puerto de canal.
- Consola de cocina / caja (superficie distinta del storefront).
- Notificaciones al local y al cliente.
- Mercado Pago.
- Object storage para fotos de catálogo.
- Worker de jobs: timeout de carrito, limpieza de tokens, reintentos.
- TWA / APK.
- Admin de plataforma con acceso cross-tenant explícito.

---

## 9. Flujo de trabajo Spec-Driven Development

Agente principal: Cursor (`cursor-agent`). OpenCode está instalado para el mismo repo; no cambia el default.

Orden de trabajo:

1. Constitución (`.specify/memory/constitution.md`) y este documento.
2. Spec de feature (`/speckit-specify`) por módulo, en el orden de la sección 4.
3. Plan, tareas, implementación. Compilar, `golangci-lint` / `tsc`, tests unitarios y de aislamiento antes de dar el módulo por cerrado.
4. Contratos OpenAPI y esquemas de herramientas **antes** de generar código de adaptadores WebMCP o MCP Apps.

No se genera el monolito en el bootstrap de Spec Kit. El código empieza en el primer módulo (`tenant`).
