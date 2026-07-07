# 📝 GoPost API

API RESTful para gestión de posts y usuarios construida con Go, siguiendo principios de arquitectura limpia.

## 🚀 Características

- ✅ Autenticación JWT
- ✅ CRUD completo de Posts
- ✅ Gestión de Usuarios
- ✅ Arquitectura en capas (Handlers → Services → Repositories)
- ✅ Seguridad con bcrypt para contraseñas
- ✅ Validación de datos
- ✅ Relaciones usuario-posts
- ✅ Errores de dominio tipados (sentinel errors) mapeados a códigos HTTP correctos (404/403, no un 400 genérico)
- ✅ Apagado ordenado (graceful shutdown) y timeouts HTTP configurados
- ✅ Sin fuga de errores internos al cliente en respuestas 500 (se registran en el log del servidor)

## 📋 Requisitos

- **Go**: 1.25.5 o superior
- **MySQL**: 8.0 o superior
- **Git**: Para clonar el repositorio

## 🛠️ Instalación

### 1. Clonar el repositorio

```bash
git clone https://github.com/alexroel/gopost-api-rest.git
cd gopost-api-rest
```

### 2. Instalar dependencias

```bash
go mod download
```

### 3. Configurar base de datos

Crear base de datos en MySQL:

```sql
CREATE DATABASE gopost;
USE gopost;
```

Ejecutar el schema:

```bash
mysql -u root -p gopost < database/schema.sql
```

### 4. Configurar variables de entorno

Crear archivo `.env` en la raíz del proyecto:

```env
PORT=:5050
JWT_SECRET=tu_secreto_super_seguro_aqui_minimo_32_caracteres
DATABASE_URL=root:password@tcp(localhost:3306)/gopost
```

> ⚠️ **IMPORTANTE**: Genera un JWT_SECRET seguro para producción. Puedes usar:
> ```bash
> openssl rand -base64 32
> ```

### 5. Ejecutar la aplicación

```bash
go run main.go
```

El servidor se iniciará en `http://localhost:5050`

### 🧪 Probar con Postman

Este repo incluye [`GoPost-API.postman_collection.json`](GoPost-API.postman_collection.json), lista para importar (Postman → *Import* → arrastrar el archivo).

- Variable `baseUrl` ya viene configurada a `http://localhost:5050`.
- Ejecuta **Sign Up** y **Login** primero: el token JWT se guarda automáticamente en la variable de colección `token` y el resto de requests protegidos lo usan solos (Bearer Auth).
- **Create Post** guarda el `id` creado en la variable `postId`, que reutilizan **Get/Update/Delete Post By ID**.

## 📚 Endpoints de la API

### 🔓 Públicos

#### Health Check
```http
GET /health
```

#### Registro de Usuario
```http
POST /signup
Content-Type: application/json

{
  "name": "Juan Pérez",
  "email": "juan@example.com",
  "password": "password123"
}
```

#### Login
```http
POST /login
Content-Type: application/json

{
  "email": "juan@example.com",
  "password": "password123"
}
```

**Respuesta:**
```json
{
  "message": "Inicio de sesión exitoso",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### Obtener todos los posts
```http
GET /posts
```

#### Obtener un post específico
```http
GET /posts/{id}
```

#### Obtener posts de un usuario
```http
GET /users/{id}/posts
```

### 🔐 Protegidos (Requieren autenticación)

> **Nota:** Incluir el token JWT en el header:
> ```
> Authorization: Bearer {token}
> ```

#### Obtener perfil del usuario autenticado
```http
GET /me
Authorization: Bearer {token}
```

#### Crear un post
```http
POST /posts
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "Mi primer post",
  "content": "Este es el contenido de mi post"
}
```

#### Obtener mis posts
```http
GET /posts/me
Authorization: Bearer {token}
```

#### Actualizar un post
```http
PUT /posts/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "Título actualizado",
  "content": "Contenido actualizado"
}
```

#### Eliminar un post
```http
DELETE /posts/{id}
Authorization: Bearer {token}
```

## 🏗️ Arquitectura del Proyecto

```
gopost-api/
├── config/          # Configuración de la aplicación
├── database/        # Conexión y schema de BD
├── handlers/        # Controladores HTTP
├── middleware/      # Middleware de autenticación
├── models/          # Estructuras de datos
├── repositories/    # Capa de acceso a datos
├── server/          # Servidor HTTP personalizado
├── services/        # Lógica de negocio
├── main.go          # Punto de entrada
├── go.mod           # Dependencias
└── .env             # Variables de entorno
```

### Flujo de datos

```
Request → Handler → Service → Repository → Database
                      ↓
                 Validación
                      ↓
Response ← Handler ← Service ← Repository ← Database
```

## 🔒 Seguridad

- **Contraseñas**: Hasheadas con bcrypt (cost 10)
- **JWT**: Tokens con expiración de 72 horas, firmados HS256; el middleware verifica explícitamente el algoritmo de firma para prevenir ataques de confusión de algoritmo
- **SQL Injection**: Prevenido con prepared statements (sin `SELECT *`, columnas explícitas)
- **Validación**: Validación de emails y longitud de contraseñas
- **Autorización**: Verificación de propiedad en operaciones de posts (solo el autor puede modificar/eliminar su post, devolviendo 403 si no le pertenece)
- **Enumeración de usuarios**: el login responde siempre el mismo mensaje genérico ("credenciales incorrectas"), sin distinguir email inexistente de contraseña incorrecta
- **Sin fuga de información interna**: los errores de infraestructura (fallos de base de datos, etc.) se registran en el log del servidor y nunca se reenvían tal cual al cliente en respuestas 5xx

## ⚠️ Manejo de errores

Las respuestas de error siguen un formato uniforme:

```json
{
  "error": "Not Found",
  "message": "post no encontrado",
  "code": 404
}
```

Los errores de negocio de `PostService` (post inexistente, post de otro usuario) son errores tipados (*sentinel errors*) que el handler traduce a códigos HTTP precisos:

| Situación | Código |
|---|---|
| Datos de entrada inválidos o campos faltantes | `400 Bad Request` |
| Token ausente, inválido o expirado | `401 Unauthorized` |
| El post existe pero pertenece a otro usuario | `403 Forbidden` |
| Recurso (post/usuario) no encontrado | `404 Not Found` |
| Fallo interno (base de datos, etc.) | `500 Internal Server Error` (mensaje genérico; el detalle solo queda en el log) |

## 🧪 Validaciones

### Usuarios
- ✅ Email: Formato válido requerido
- ✅ Password: Mínimo 6 caracteres
- ✅ Name, Email, Password: Campos obligatorios

### Posts
- ✅ Title y Content: Campos obligatorios
- ✅ Autorización: Solo el autor puede modificar/eliminar sus posts

## 📦 Dependencias

```go
require (
    github.com/go-sql-driver/mysql v1.9.3
    github.com/golang-jwt/jwt/v5 v5.3.0
    github.com/joho/godotenv v1.5.1
    golang.org/x/crypto v0.46.0
)
```

## 🐛 Solución de Problemas

### Error: "JWT_SECRET es requerido"
Asegúrate de tener un archivo `.env` con la variable `JWT_SECRET` configurada.

### Error de conexión a MySQL
Verifica que:
- MySQL esté corriendo
- Las credenciales en `DATABASE_URL` sean correctas
- La base de datos `gopost` exista

### Error: "email ya está registrado"
El email ya existe en la base de datos. Usa otro email o inicia sesión.

## 📝 Variables de Entorno

| Variable | Descripción | Valor por Defecto | Requerido |
|----------|-------------|-------------------|-----------|
| `PORT` | Puerto del servidor | `:5050` | No |
| `JWT_SECRET` | Clave secreta para JWT | - | **Sí** |
| `DATABASE_URL` | URL de conexión MySQL | `root:password@tcp(localhost:3306)/gopost` | No |

## 🚀 Próximas Mejoras

- [x] Graceful shutdown (implementado: `server.RunServer` captura SIGINT/SIGTERM y drena conexiones antes de cerrar)
- [ ] Tests unitarios e integración (requiere extraer interfaces de los repositorios para poder mockearlos)
- [ ] Paginación en listado de posts
- [ ] Rate limiting
- [ ] CORS configurado
- [ ] Swagger/OpenAPI documentation
- [ ] Docker y Docker Compose
- [ ] Migraciones de base de datos
- [ ] Logging estructurado (`log/slog` o similar, con niveles y correlación por request)
- [ ] Migrar el router propio a un framework web (ej. Fiber) si el proyecto crece
- [ ] CI/CD pipeline

Ver [`GUIA_DE_DESARROLLO.md`](GUIA_DE_DESARROLLO.md) para el detalle de cómo abordar cada una de estas mejoras.

## 👨‍💻 Desarrollo

### Ejecutar en modo desarrollo
```bash
go run main.go
```

### Compilar para producción
```bash
go build -o gopost-api
./gopost-api
```

### Formato de código
```bash
go fmt ./...
```

### Linting
```bash
go vet ./...
```

## 📚 Documentación adicional

- [`GUIA_DE_DESARROLLO.md`](GUIA_DE_DESARROLLO.md): tutorial paso a paso de cómo se construyó esta API, explicando las decisiones arquitectónicas, los patrones de diseño aplicados (Repository, DTO, Decorator, Sentinel Errors, Composition Root, etc.) y una sección de mejoras futuras en detalle.
- [`CLAUDE.md`](CLAUDE.md): guía de arquitectura y convenciones del proyecto pensada para agentes de IA (Claude Code) que colaboren en el repositorio.

## 📄 Licencia

Este proyecto es de código abierto.

---

⭐ Si este proyecto te fue útil, considera darle una estrella en GitHub!
