# Feature Specification: Esqueleto del monolito multi-tenant

**Feature Branch**: `001-monolith-scaffold`

**Created**: 2026-09-21

**Status**: Draft

**Input**: User description: "quiero ir implementando la estructura del proyecto ir creando las carpetas el docker el .env el .env.example el schema de la base de datos obviamente todo dockerizado implementado el monolito modular con el multi tenant en el folder docs/ esta el plan inicial"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Arrancar el entorno local completo (Priority: P1)

Una persona del equipo clona el repositorio, copia las variables documentadas, levanta el entorno empaquetado y ve tres piezas vivas: almacén de datos, aplicación de negocio y vitrina. No tiene que instalar la base de datos ni el runtime de la aplicación a mano.

**Why this priority**: Sin un entorno repetible no se puede construir el resto de módulos ni demostrar el producto.

**Independent Test**: En una máquina limpia (salvo la herramienta de contenedores), seguir el README, copiar el ejemplo de variables y arrancar. En menos de 15 minutos hay señal de vida de datos, aplicación y vitrina.

**Acceptance Scenarios**:

1. **Given** un clone sin artefactos locales, **When** se copian las variables de ejemplo y se arranca el entorno documentado, **Then** el almacén de datos acepta conexiones, la aplicación responde “estoy viva” y la vitrina carga una página mínima.
2. **Given** el entorno ya arrancado, **When** se detiene y se vuelve a arrancar, **Then** los datos de comercios de ejemplo siguen ahí.
3. **Given** alguien abre el repositorio, **When** busca secretos de base de datos o de proveedores, **Then** no hay valores reales versionados; sí hay una plantilla que nombra cada variable necesaria.

---

### User Story 2 - Dos comercios aislados desde el primer día (Priority: P1)

La plataforma nace con al menos dos comercios de ejemplo (slug y host distintos). Cada uno tiene moneda, impuesto y zona horaria por defecto de Colombia. Una consulta hecha en el contexto del comercio A no ve ni cambia datos del comercio B.

**Why this priority**: El aislamiento de inquilino es regla de la constitución. Si se pospone, cada módulo posterior lo reescribe mal.

**Independent Test**: Crear o sembrar dos comercios, fijar el contexto al A, intentar leer filas del B, y comprobar que falla o devuelve vacío. Lo mismo al revés. Automatizable.

**Acceptance Scenarios**:

1. **Given** comercios A y B sembrados, **When** una petición llega con el host o slug de A, **Then** el sistema trabaja solo sobre A y no usa un identificador de inquilino enviado por el cliente como autoridad.
2. **Given** una fila de negocio de A, **When** una sesión de aplicación actúa como B, **Then** esa fila no se lee ni se muta.
3. **Given** un comercio nuevo, **When** se registra sin indicar mercado, **Then** hereda COP, IVA y zona horaria de Bogotá.

---

### User Story 3 - Mapa de módulos listo para el orden de construcción (Priority: P2)

El repositorio muestra el monolito modular: piezas compartidas (configuración, datos, resolución de comercio) y carpetas vacías o esqueleto para `tenant`, `auth`, `catalog`, `cart`, `orders`, `payments`, `agent` y el puerto de canal. Esta feature implementa de verdad solo la pieza `tenant` más la plataforma compartida. El resto existe para no reordenar el árbol después.

**Why this priority**: Evita que el siguiente módulo invente otra estructura. No entrega pedido ni chat todavía.

**Independent Test**: Inspeccionar el árbol y comprobar que coincide con el documento de entrada en `docs/`. Arrancar la aplicación sin auth, catálogo ni pedidos y ver que no falla por módulos ausentes.

**Acceptance Scenarios**:

1. **Given** el repositorio arrancable, **When** se recorre la estructura, **Then** existen las carpetas de plataforma y de cada módulo del primer corte, más canal y el binario de API.
2. **Given** esa estructura, **When** se llama a la señal de vida, **Then** responde sin exigir catálogo, carrito, pago ni agente.
3. **Given** el módulo `tenant`, **When** se lista comercios por host o slug, **Then** se obtiene el comercio correcto o un no encontrado claro.

---

### User Story 4 - Esquema de datos que fuerza el aislamiento (Priority: P1)

El almacén aplica el modelo de inquilino: cada comercio es una fila propia; las tablas de negocio futuras llevan el discriminador de comercio; las políticas de fila están forzadas; el rol con el que corre la aplicación no es el dueño de las tablas.

**Why this priority**: Sin esto, “multi-tenant” es solo un comentario en el código.

**Independent Test**: Conectar como rol de aplicación e intentar leer sin contexto de comercio; no hay fugas. Un test automatizado cubre A contra B.

**Acceptance Scenarios**:

1. **Given** el esquema aplicado, **When** la aplicación abre una transacción de negocio, **Then** el comercio queda fijado en esa transacción y las lecturas quedan filtradas a ese comercio.
2. **Given** el rol de aplicación, **When** se omite el contexto de comercio, **Then** no se listan filas de todos los comercios juntos.
3. **Given** migraciones versionadas, **When** se recrea el entorno desde cero, **Then** el esquema y los dos comercios de ejemplo quedan iguales.

---

### Edge Cases

- Variable de entorno obligatoria ausente: el arranque falla con mensaje que nombra la variable; no usa un secreto inventado en silencio.
- Host o slug desconocido: respuesta de no encontrado, sin filtrar a un comercio por defecto.
- Host y slug que no coinciden en la misma petición: gana el host público; no se mezcla contexto.
- Recrear el entorno borra datos locales no versionados; el ejemplo de variables y las migraciones bastan para volver al estado semilla.
- Intento de crear un segundo comercio con el mismo slug u host: se rechaza.
- La vitrina en este corte no autentica ni toma pedidos; si alguien llama rutas de pedido, no existen aún.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST ofrecer un entorno local empaquetado que levante almacén de datos, aplicación de negocio y vitrina con un procedimiento documentado en el README.
- **FR-002**: El repositorio MUST incluir una plantilla de variables (`env` de ejemplo) que liste cada secreto y ajuste necesario, sin valores reales. El archivo de valores locales MUST quedar fuera de git (ya cubierto por el ignore de la raíz).
- **FR-003**: El sistema MUST persistir comercios con slug único, host único, nombre, moneda, impuesto, zona horaria y horario del local.
- **FR-004**: Un comercio nuevo MUST heredar por defecto COP, IVA y `America/Bogota` si no se indica otro mercado.
- **FR-005**: El sistema MUST resolver el comercio activo por host o slug público. MUST NOT tratar un identificador de inquilino enviado por el cliente (cuerpo o cabecera inventable) como autoridad.
- **FR-006**: Toda fila de negocio MUST llevar el discriminador del comercio. Las políticas de aislamiento MUST estar forzadas. El rol de aplicación MUST ser distinto del dueño de las tablas. El comercio MUST fijarse por transacción.
- **FR-007**: El sistema MUST sembrar al menos dos comercios de ejemplo y MUST incluir un test automatizado que demuestre que el comercio A no lee ni muta al comercio B.
- **FR-008**: El árbol del repositorio MUST seguir el mapa de directorios del documento de entrada: aplicación modular, vitrina, migraciones versionadas, entorno empaquetado y documentación.
- **FR-009**: Esta feature MUST implementar la plataforma compartida (configuración tipada, conexión y transacción con contexto de comercio, resolución de comercio) y el módulo `tenant`. MUST crear el esqueleto de `auth`, `catalog`, `cart`, `orders`, `payments`, `agent`, `channel` y el hueco del trabajador de jobs, sin su lógica de negocio.
- **FR-010**: La aplicación MUST exponer una señal de vida y MUST arrancar aunque los módulos posteriores estén vacíos.
- **FR-011**: La imagen de la aplicación MUST poder hacer salidas cifradas hacia servicios externos (certificados de confianza incluidos). MUST NOT publicarse una imagen sin esos certificados.
- **FR-012**: Esta feature MUST NOT implementar login de staff, sesión de invitado, catálogo, carrito, pedidos, cobros, clasificador, widgets ni mensajería.

### Key Entities

- **Comercio (tenant)**: negocio local que usa la plataforma. Identidad pública (slug, host), nombre, mercado (moneda, impuesto, zona horaria), horario. Es la raíz de aislamiento.
- **Contexto de comercio**: comercio resuelto para una petición o transacción. Se obtiene del host o slug, no de un id que el cliente imponga.
- **Rol de aplicación**: identidad con la que el software consulta datos. No es dueña de las tablas; ve solo el comercio del contexto.
- **Semilla de desarrollo**: dos comercios ficticios para probar aislamiento y arranque.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Una persona que ya tiene la herramienta de contenedores deja el entorno usable (datos, aplicación y vitrina con señal de vida) en menos de 15 minutos siguiendo solo el README.
- **SC-002**: El 100 % de los intentos automatizados de que el comercio A lea o cambie datos del comercio B fallan (cero fugas en la batería de aislamiento).
- **SC-003**: Una búsqueda en el historial versionado no encuentra contraseñas, cadenas de conexión con clave, ni tokens; la plantilla de variables cubre el 100 % de lo que el arranque exige.
- **SC-004**: Recrear el entorno desde cero restaura los dos comercios de ejemplo y la señal de vida sin pasos no documentados.
- **SC-005**: Un revisor puede señalar en el árbol cada módulo del primer corte; ninguno falta. Pedido, pago y agente no son usables aún (respuesta de no implementado o ausencia de rutas).

## Assumptions

- El documento [docs/especificacion-y-arquitectura.md](../../docs/especificacion-y-arquitectura.md) y la constitución mandan. Esta feature es el arranque de ese árbol, no un producto de pedidos completo.
- Entorno empaquetado: Compose con almacén PostgreSQL, aplicación (Go, Ent, Atlas, pgx) y vitrina (Next.js App Router, TypeScript, Tailwind) según ese documento.
- Variables: `.env.example` versionado; `.env` local ignorado.
- Esquema de esta feature: tabla de comercios más la maquinaria de aislamiento que heredarán las tablas futuras. No se modelan productos, carritos, órdenes ni pagos aquí.
- Semilla: dos comercios colombianos de ejemplo (por ejemplo `demo-a` y `demo-b`) con hosts de desarrollo.
- Resolución local: host (`demo-a.localhost`) y/o ruta de slug; el plan de implementación elige el detalle sin contradecir FR-005.
- El binario trabajador de jobs existe como hueco (`cmd/worker`) y no corre en este corte.
- No se publica a un registro de imágenes ni a producción.

## Out of Scope

- WhatsApp, consola de cocina, notificaciones, Mercado Pago, fotos, TWA/APK, admin cross-tenant.
- Autenticación, catálogo, carrito, máquina de estados de pedido, agente JEV, MCP Apps, WebMCP.
- Lighthouse, rate limit del clasificador y bitácora de transiciones de orden (llegan con sus módulos).
