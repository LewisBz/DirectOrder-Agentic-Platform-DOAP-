# Feature Specification: Autenticación de staff y sesión de invitado

**Feature Branch**: `002-staff-auth`

**Created**: 2026-09-21

**Status**: Draft

**Input**: User description: "seguir con el plan, creo que ahora viene el tema de el modulo de autenticacion me gustaria crear el crud basico con la vista en nextjs para probarlo y dsp del acceso una vista rapida asi sea un hola mundo para comprobar que todo sirva"

## Clarifications

### Session 2026-09-21

- Q: En este recorte, ¿el visitante de la vitrina debe obtener ya una sesión de invitado (sin poder entrar al saludo ni al CRUD de personal), o eso se deja para el módulo de carrito? → A: Emitir invitado ahora: sesión anónima por comercio, bloqueada en saludo y CRUD
- Q: ¿La sesión de invitado se crea sola al abrir la vitrina de un comercio conocido, o solo cuando la persona pulsa una acción explícita? → A: Automática al abrir la vitrina de un comercio conocido
- Q: Si el dueño desactiva a alguien del equipo, ¿puede volver a activar esa misma cuenta, o la baja es definitiva? → A: El dueño puede reactivar la misma cuenta; entonces vuelve a poder iniciar sesión
- Q: Al crear o editar personal, ¿el dueño puede asignar también el rol dueño a otra persona, o en este recorte solo caja y cocina? → A: No: altas y ediciones de rol solo caja o cocina; el dueño demo viene de la semilla
- Q: En el CRUD de prueba, ¿el dueño puede establecer una contraseña nueva al editar a caja o cocina, o la contraseña solo se define al crear la cuenta? → A: Al editar, el dueño puede (opcional) definir una contraseña nueva; si no la cambia, la anterior sigue

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Entrar al backoffice y ver que la sesión sirve (Priority: P1)

Una persona del comercio (dueño, caja o cocina) abre la pantalla de acceso del comercio, ingresa correo y contraseña, y llega a una pantalla mínima de confirmación: un saludo con su nombre, rol y el nombre del comercio. No hay panel de pedidos todavía; basta comprobar que el acceso quedó ligado a ese comercio.

**Why this priority**: Sin una prueba visible de “entré y el sistema me reconoce”, el módulo de autenticación no se puede demostrar ni usar como base del resto del backoffice.

**Independent Test**: Con un comercio de ejemplo y una cuenta de dueño sembrada, iniciar sesión y ver el saludo. Cerrar sesión e intentar abrir de nuevo la pantalla de saludo: se pide acceso otra vez.

**Acceptance Scenarios**:

1. **Given** un comercio de ejemplo con una cuenta de dueño válida, **When** la persona envía correo y contraseña correctos en la pantalla de acceso de ese comercio, **Then** ve el saludo con nombre, rol y comercio en pocos segundos.
2. **Given** una sesión ya iniciada, **When** recarga la pantalla de saludo, **Then** sigue autenticada sin volver a escribir la contraseña.
3. **Given** una sesión iniciada, **When** cierra sesión, **Then** la pantalla de saludo deja de estar disponible y se muestra de nuevo el acceso.
4. **Given** correo o contraseña incorrectos, **When** intenta entrar, **Then** no entra y ve un mensaje genérico que no revela si el correo existe.

---

### User Story 2 - CRUD de personal del comercio (Priority: P1)

El dueño del comercio lista, crea, edita, desactiva y reactiva cuentas de caja y cocina en pantallas de prueba ligadas a su comercio. El dueño de cada demo nace de la semilla; no se crea ni se cambia a dueño desde esas pantallas. Caja y cocina pueden entrar (historia 1) pero no administrar al resto del equipo.

**Why this priority**: El usuario necesita un CRUD visible para ejercitar el módulo; además el producto exige roles de staff por comercio, no un usuario global.

**Independent Test**: Como dueño de A, crear una cuenta de caja, verla en la lista, cambiarle el nombre o el rol, opcionalmente poner una contraseña nueva, desactivarla (ya no inicia sesión), reactivarla (vuelve a iniciar sesión). Como caja, intentar abrir la gestión de personal y ser rechazado.

**Acceptance Scenarios**:

1. **Given** un dueño autenticado en el comercio A, **When** crea una cuenta con correo, contraseña, nombre y rol (caja o cocina), **Then** esa persona aparece en la lista de A y puede iniciar sesión en A.
2. **Given** esa lista, **When** el dueño edita nombre o cambia el rol entre caja y cocina sin enviar contraseña nueva, **Then** el cambio se refleja al recargar y en el saludo de la siguiente sesión; el acceso sigue con la contraseña anterior.
3. **Given** esa cuenta, **When** el dueño guarda una edición con una contraseña nueva válida, **Then** el acceso anterior deja de servir y el nuevo funciona; no hay flujo de “olvidé mi contraseña” por correo.
4. **Given** una cuenta activa, **When** el dueño la desactiva, **Then** desaparece de las cuentas operativas, no puede iniciar sesión y las sesiones previas dejan de valer.
5. **Given** esa cuenta desactivada, **When** el dueño la reactiva, **Then** vuelve a aparecer como operativa y puede iniciar sesión de nuevo; las sesiones anteriores a la baja siguen sin valer.
6. **Given** una persona con rol caja o cocina autenticada, **When** intenta crear, editar, desactivar o reactivar staff, **Then** la operación se rechaza y no cambia datos.
7. **Given** un correo ya usado en el mismo comercio (activo o desactivado), **When** se intenta crear otra cuenta con ese correo, **Then** se rechaza con un error claro; el mismo correo puede existir en otro comercio.

---

### User Story 3 - El personal de un comercio no cruza al otro (Priority: P1)

Las cuentas, sesiones y pantallas de staff viven dentro del comercio resuelto por host o slug. El personal de A no lista, muta ni usa sesiones del personal de B. Un identificador de comercio enviado por el cliente no cambia el comercio de la sesión.

**Why this priority**: La constitución exige aislamiento comprobable en cada módulo de negocio. Auth es el primero que guarda personas.

**Independent Test**: Sembrar dueños en A y B. Autenticado como A, listar staff: solo A. Intentar leer o editar el identificador de una cuenta de B: no encontrado o vacío. Una sesión de A no abre el saludo de B aunque se cambie de host.

**Acceptance Scenarios**:

1. **Given** staff en A y en B, **When** el dueño de A lista el equipo, **Then** no aparecen cuentas de B.
2. **Given** el dueño de A, **When** intenta actualizar o desactivar una cuenta que solo existe en B, **Then** no hay efecto sobre B.
3. **Given** una sesión válida de A, **When** se abre el acceso o el saludo en el host o slug de B, **Then** no se trata como personal de B; debe autenticarse en B o ser rechazado.
4. **Given** cualquier petición de staff, **When** el cliente envía un identificador de comercio inventado, **Then** el comercio sigue siendo el del host o slug.

---

### User Story 4 - Invitado de vitrina, distinto del staff (Priority: P2)

Al abrir la vitrina de un comercio conocido, el visitante recibe una sesión de invitado sin cuenta de empleado ni clic extra. Esa sesión identifica el canal y servirá al carrito más adelante. No otorga el saludo de staff ni el CRUD de personal.

**Why this priority**: Staff e invitado son dos identidades del módulo de autenticación. En este recorte el invitado ya se emite y se aísla; no espera al carrito. El foco de demostración sigue siendo el CRUD y el saludo de staff.

**Independent Test**: Abrir la vitrina de A (sin pulsar un botón de “continuar”). Existe sesión de invitado de A. Abrir el saludo de staff: se exige acceso de empleado. Esa sesión no lista staff.

**Acceptance Scenarios**:

1. **Given** un visitante sin sesión previa, **When** abre la vitrina del comercio A, **Then** obtiene automáticamente una sesión de invitado de A, sin correo de empleado y sin una acción explícita.
2. **Given** solo sesión de invitado, **When** abre la pantalla de saludo o el CRUD de personal, **Then** no entra; se pide acceso de staff.
3. **Given** un invitado en A, **When** se usa esa sesión en el contexto de B, **Then** no se reutiliza como invitado de B.

---

### Edge Cases

- Comercio desconocido (host o slug inválido): no hay pantalla de acceso de un comercio “por defecto”; respuesta de no encontrado; no se emite sesión de invitado.
- Cuenta desactivada con contraseña correcta: mismo resultado de acceso denegado que una contraseña mala (mensaje genérico). Tras reactivar, el acceso con esa contraseña vuelve a ser válido.
- Contraseña demasiado corta o correo mal formado al crear staff, o al enviar una contraseña nueva en una edición: se rechaza antes de guardar. Si la edición no incluye contraseña nueva, se conserva la anterior.
- El dueño no puede dejar el comercio sin ningún dueño activo: no se desactiva ni se degrada el último dueño. En este recorte no hay alta ni promoción a dueño desde las pantallas; el dueño demo es semilla.
- Intento de crear o editar un miembro con rol dueño desde las pantallas: se rechaza.
- Cierre de sesión o renovación: un identificador de sesión ya rotado o cerrado no vuelve a abrir el saludo.
- Ausencia de contexto de comercio en el almacén: las filas de personal no se listan todas juntas.
- Intento de registro público de empleados: no existe; solo el dueño (o semilla local) crea staff.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST autenticar staff por correo y contraseña dentro del comercio resuelto por host o slug, nunca por un identificador de comercio que el cliente invente.
- **FR-002**: Las contraseñas de staff MUST almacenarse de forma no reversible, con un costo de hash alto (trabajo 12 o equivalente constitucional). El frontend MUST NOT contener secretos de firma ni de base de datos.
- **FR-003**: Tras un acceso correcto, el sistema MUST emitir una sesión de staff de vida corta que porta rol y comercio, más una renovación que invalida el identificador anterior (rotación).
- **FR-004**: El sistema MUST ofrecer pantallas de prueba para: acceso, listado/alta/edición/desactivación/reactivación de staff, y un saludo mínimo post-acceso (nombre, rol, comercio) suficiente para comprobar que la sesión sirve.
- **FR-005**: Solo el rol dueño MUST crear, editar, desactivar y reactivar caja y cocina de su comercio. Caja y cocina MUST poder autenticarse y ver el saludo, y MUST NOT administrar el equipo.
- **FR-006**: El correo MUST ser único por comercio, no global. Roles de cuenta: dueño, caja, cocina. En este recorte, altas y cambios de rol desde las pantallas MUST ser solo caja o cocina; el dueño de cada comercio demo MUST existir por semilla, no por el CRUD de prueba.
- **FR-007**: Desactivar un miembro MUST impedir nuevos accesos y MUST invalidar sesiones vigentes. Reactivarlo MUST permitir un nuevo acceso con las mismas credenciales; MUST NOT revalidar las sesiones emitidas antes de la baja. La baja es lógica: el correo sigue único en ese comercio (activo o no).
- **FR-008**: El sistema MUST sembrar, en el entorno local de ejemplo, al menos un dueño por comercio demo con credenciales documentadas solo en la plantilla de variables o en el quickstart, nunca como secreto de producción versionado.
- **FR-009**: Al abrir la vitrina de un comercio conocido, el sistema MUST emitir (sin acción explícita) una sesión de invitado independiente, sin rol de staff, acotada a ese comercio.
- **FR-010**: El personal y las sesiones de invitado MUST llevar discriminador de comercio; las políticas de fila forzadas MUST impedir que A lea o mute a B; un test automatizado lo demuestra.
- **FR-011**: Fallos de acceso MUST usar un mensaje genérico. Los registros de operación MUST NOT volcar contraseñas ni PII innecesaria.
- **FR-012**: Esta feature MUST NOT implementar recuperación de contraseña por correo, OAuth, admin de plataforma, catálogo, carrito, pedidos, pagos ni agente.
- **FR-013**: Al editar caja o cocina, el dueño MAY definir una contraseña nueva. Si no la envía, MUST conservarse la anterior. Si la envía, MUST sustituirla y MUST invalidar las sesiones vigentes de esa persona. Esto MUST NOT ser un flujo de recuperación por correo.

### Key Entities

- **Miembro de staff**: Persona operativa de un comercio. Correo, nombre visible, rol (dueño / caja / cocina), estado activo o desactivado, credencial no reversible, pertenencia exclusiva a un comercio.
- **Sesión de staff**: Acceso de corta duración ligado a un miembro, su rol y su comercio; se puede renovar rotando el identificador de renovación; se puede cerrar.
- **Sesión de invitado**: Identidad anónima de vitrina ligada a un comercio y a un canal; no es empleado.
- **Comercio**: Ya existe; esta feature no lo crea. Es el límite de todas las cuentas y sesiones.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Un dueño de ejemplo completa el acceso y ve el saludo correcto (nombre, rol, comercio) en menos de 30 segundos en un entorno local ya arrancado.
- **SC-002**: Un dueño crea una cuenta de caja y esa caja inicia sesión y ve su propio saludo en el mismo comercio, en menos de 2 minutos, sin pasos fuera de las pantallas de prueba.
- **SC-003**: En 100 % de las pruebas de aislamiento, el dueño de A no lista ni altera staff de B, y una sesión de A no autentica el saludo de B.
- **SC-004**: El 100 % de los intentos con contraseña incorrecta, cuenta desactivada o sesión de invitado sobre pantallas de staff terminan en acceso denegado, sin fuga de si el correo existe.
- **SC-005**: Tras cerrar sesión, el 100 % de las recargas de la pantalla de saludo piden de nuevo el acceso.
- **SC-006**: Tras abrir la vitrina de A, el visitante ya tiene sesión de invitado y no puede completar ninguna acción de administración de personal.

## Assumptions

- El esqueleto `001-monolith-scaffold` ya resuelve el comercio por host o slug y siembra Demo A y Demo B.
- “CRUD básico + hola mundo” es el recorte de demostración: staff (acceso, CRUD, saludo) más emisión de sesión de invitado por comercio, sin carrito ni backoffice final.
- No hay auto-registro público de empleados. La semilla local cubre el único dueño de cada demo; el CRUD de prueba no crea dueños.
- El admin de plataforma y el acceso cruzado entre comercios quedan fuera.
- Recuperación de contraseña por correo, segundo factor y proveedores externos quedan fuera. El dueño sí puede asignar una contraseña nueva en la edición de caja o cocina.
- La vitrina de prueba vive en las rutas del comercio ya usadas (host o slug); no se pide un producto de diseño.
- El carrito usará la sesión de invitado en una feature posterior; aquí solo se emite y se rechaza su uso como staff.
- Colombia (COP, IVA, Bogotá) no cambia en esta feature; no hay datos fiscales nuevos en el personal.
- Las credenciales de demo son placeholders locales (`changeme` o equivalente documentado), rotables, no usadas en producción.
