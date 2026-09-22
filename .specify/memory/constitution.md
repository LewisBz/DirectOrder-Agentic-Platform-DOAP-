# Constitución DOAP

Este archivo resume las reglas vinculantes del proyecto. El detalle de producto y arquitectura está en [docs/especificacion-y-arquitectura.md](../../docs/especificacion-y-arquitectura.md). Si este resumen y esa especificación discrepan, **manda la especificación**.

## Core Principles

### I. El comercio es dueño del pedido y del cliente

DOAP no es un marketplace y no compite en flota de repartidores. El software recibe, confirma, cobra y cumple pedidos en el canal del comercio. El primer mercado es Colombia (COP, IVA, `America/Bogota`). El primer corte cobra en efectivo. Mercado Pago queda detrás de la interfaz de pagos y no se implementa hasta que el núcleo cierre.

### II. La API HTTP es el límite del sistema

OpenAPI describe las operaciones de dominio. WebMCP (`document.modelContext.registerTool`, con feature detection) y MCP Apps (widgets `ui://` en iframe) son adaptadores. No son el backend. `submitOrder` exige confirmación explícita de la persona. El núcleo de pedidos habla un puerto de canal: la PWA es el primer adaptador; WhatsApp entra después sin reescribir órdenes.

### III. El modelo no tiene autoridad sobre el dinero

JEV AI clasifica y propone. El servidor resuelve contra el catálogo y calcula precio, IVA, disponibilidad y total. Si la confianza es baja, se pregunta con un widget o se escala a una persona. Totales, precios o SKU enviados por el modelo o el navegador no son autoridad.

### IV. Aislamiento de tenant comprobable

Toda fila de negocio lleva `tenant_id`. El tenant se resuelve por host o slug, no por un header que el cliente pueda inventar. PostgreSQL usa `FORCE ROW LEVEL SECURITY`, un rol de aplicación distinto del owner, y `SET LOCAL` del tenant en la transacción. Un módulo no se cierra sin un test de que el tenant A no lee ni muta al tenant B.

### V. Pedido atómico e idempotente

`submitOrder` lleva llave de idempotencia. Reserva de stock y creación de la orden ocurren en la misma transacción ACID. La máquina de estados declara actor por transición e incluye cancelación, rechazo, fallo de pago y listo para recoger. `DISPATCHED` no aplica a recoger en tienda.

## Stack y módulos

- Backend: Go, Ent, Atlas, pgx, validator/v10. GORM no. Imagen Docker con CA certs (TLS de salida).
- Frontend: Next.js App Router, TypeScript, Tailwind. Astro y Vite 8.3 no.
- Identidades: sesión de invitado en el storefront; staff (dueño, caja, cocina) con bcrypt costo 12 y JWT de vida corta más refresh rotativo. El admin de plataforma no lee todos los tenants por defecto.
- Primer corte, en este orden: `tenant` → `auth` → `catalog` → `cart` → `orders` → `payments` → `agent`.
- El módulo `agent` es el puerto del clasificador y la resolución contra catálogo. No registra herramientas en el DOM y no sirve de host MCP Apps.

## Calidad y operación

- SLO separados: lecturas de catálogo frente a turnos del agente. No hay un p99 único de 15 ms para todo el backend.
- Secretos fuera del frontend y de git.
- Rate limit y tope de costo en el endpoint de lenguaje natural.
- Tests: aislamiento de tenant, contratos OpenAPI y contratos de herramientas, además de linters (`golangci-lint`, `tsc`).
- El service worker de la PWA no cachea respuestas autenticadas ni reenvía un pedido viejo.
- Bitácora de transiciones por tenant, sin PII en el log.

## Flujo de desarrollo

Spec-Driven Development con Spec Kit. Agente principal: Cursor (`cursor-agent`). OpenCode está instalado al lado y no es el default. Contratos OpenAPI y esquemas de herramientas se escriben antes del código de adaptadores. No se genera el monolito en el bootstrap; el código empieza en el módulo `tenant`.

## Governance

Las specs de feature, planes y código deben cumplir esta constitución y la especificación de entrada. Una excepción se documenta en la spec de la feature (qué se relaja, por qué, y cómo se recupera). Complejidad extra (nuevo módulo, nueva pasarela, nuevo canal) se justifica contra el orden de construcción de la sección de módulos.

**Version**: 1.0.0 | **Ratified**: 2026-09-21 | **Last Amended**: 2026-09-21
