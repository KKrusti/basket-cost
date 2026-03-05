# PROGRESS.md

Resumenes automáticos generados por agentes tras completar una tarea.

Este archivo debe contener un resumen claro del estado tras cada tarea completada, permitiendo retomar el proyecto desde cero en caso de interrupciones de la sesión.

---

## Sesión: Multi-tenancy + JWT Auth (2026-03-05)

**Branch:** `refactor/security-solid-dry-improvements`

### Estado actual: COMPLETADO (backend)

Se ha implementado soporte multi-tenant completo en el backend. Todos los tests del backend pasan.

### Cambios realizados en esta sesión

1. **`cmd/seed/main.go`**: Corregido para pasar `userID=0` a `imp.Import()` (datos de seed son anónimos).

2. **`internal/store/store.go`**:
   - Añadido helper `nullableUserID(int64) any` que convierte `0` → `nil` (SQL NULL).
   - `UpsertPriceRecord` y `UpsertPriceRecordBatch`: usan `nullableUserID` al insertar `user_id`.
   - `SearchProducts`, `GetMostPurchased`, `GetBiggestPriceIncreases`: cuando `userID=0` usan `IS NULL`.
   - `IsFileProcessed` y `MarkFileProcessed`: mismo patrón NULL para `userID=0`.
   - `SetProductImageURLManual`: nuevo método que marca `image_url_locked=1`.

3. **`internal/handlers/handlers.go`**:
   - `UserIDContextKey` exportado para que el middleware en `cmd/server` lo pueda usar.
   - `RegisterHandler` y `LoginHandler`: nuevos handlers de autenticación JWT.
   - `ProductImageHandler`: nuevo handler PATCH para actualizar imagen manualmente.
   - Handlers existentes actualizados para pasar `userID` desde el contexto.

4. **`internal/auth/auth.go`** (nuevo): bcrypt + JWT (HS256, TTL 72h).

5. **`internal/auth/auth_test.go`** (nuevo): 5 tests para hash, check y tokens JWT.

6. **`internal/database/db.go`**: Migraciones m4 (tabla `users`), m5 (`user_id` nullable en `price_records`/`processed_files`), m6 (`image_url_locked` en `products`).

7. **`internal/models/models.go`**: Tipo `User` y campo `ImageURLLocked` en `Product`.

8. **`cmd/server/main.go`**:
   - Middleware `optionalAuthMiddleware`: JWT opcional (userID=0 si no hay token).
   - Rutas nuevas: `POST /api/auth/register`, `POST /api/auth/login`.
   - CORS actualizado para soportar `PATCH` y header `Authorization`.

9. **`internal/handlers/handlers_test.go`**: 10 nuevos tests auth + corrección `IsFileProcessed`.

10. **`internal/store/store_test.go`**: Todos los tests actualizados al patrón multi-tenant.

### Arquitectura de autenticación

- **Anónimo** (`userID=0`): datos con `user_id IS NULL`. Retrocompatible con seed.
- **Autenticado** (`userID>0`): datos filtrados por `user_id`. Cada usuario ve sus propios tickets.
- El frontend aún **no** envía JWT. Pendiente: integración frontend.

### Próximos pasos (frontend)

- Pantalla de login/registro.
- Almacenar JWT en `localStorage`.
- Incluir `Authorization: Bearer <token>` en llamadas a la API.

### Tests (todos pasan)

```
ok  basket-cost/internal/auth      (5 tests)
ok  basket-cost/internal/handlers  (todos pasan, 10 nuevos auth tests)
ok  basket-cost/internal/store     (todos pasan)
ok  basket-cost/internal/ticket    (todos pasan)
ok  basket-cost/internal/enricher  (todos pasan)
```
