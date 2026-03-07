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

Completado en sesión 2026-03-07 — ver entrada siguiente.

### Tests (todos pasan)

```
ok  basket-cost/internal/auth      (5 tests)
ok  basket-cost/internal/handlers  (todos pasan, 10 nuevos auth tests)
ok  basket-cost/internal/store     (todos pasan)
ok  basket-cost/internal/ticket    (todos pasan)
ok  basket-cost/internal/enricher  (todos pasan)
```

---

## Sesión: Household UI + e2e + Auto-traducción catalan (2026-03-07)

### Estado actual: COMPLETADO — no hay tareas pendientes

### Tareas completadas

#### basket-cost-2wk + basket-cost-3st (cerradas como ya implementadas)
Los componentes `HouseholdSection.tsx` y `AcceptInviteModal.tsx` ya estaban implementados con sus tests unitarios y estilos CSS. Se cerraron al verificar que el trabajo estaba completo.

#### basket-cost-mvo — e2e: household invitation and shared purchases tests

**Archivos modificados:**
- `frontend/e2e/helpers.ts`: Añadidos stubs `stubHouseholdMembers`, `stubHouseholdEmpty`, `stubCreateInvitation`, `stubAcceptInvitation`. Se usa `route.fallback()` (no `route.continue()`) para que los stubs se encadenen correctamente en Playwright.
- `frontend/e2e/household.spec.ts` (nuevo): 15 tests × 2 dispositivos = 30 tests. Cubre:
  - Sección "Unidad familiar" en el menú de usuario (miembros, badge "(tú)")
  - Crear enlace de invitación, botón copiar, error al fallar
  - Abandonar unidad: cierra el menú / muestra error
  - Flujo `/?invite=TOKEN`: modal visible cuando autenticado, no visible sin sesión
  - Aceptar invitación: llama al endpoint correcto, muestra error si expirada

**Truco para test "sin sesión":** Se usa un segundo `page.addInitScript` en el `test.describe` interior que elimina la clave de auth de localStorage, contrarrestando el `loginViaStorage` del `beforeEach` exterior.

#### basket-cost-urv — Auto-traducción de nombres catalanes via API

**Archivos creados/modificados:**
- `internal/enricher/translator.go` (nuevo):
  - Interfaz `Translator` con método `Translate(ctx, text) (string, error)`
  - `MyMemoryTranslator`: llama a `api.mymemory.translated.net` (ca→es, free, sin API key), caché en `sync.Map`, `baseURL` configurable para tests
  - `NoopTranslator`: devuelve el texto sin cambios (tests/fallback)
- `internal/enricher/enricher.go`:
  - Campo `translator Translator` en el struct `Enricher`
  - `New(s)` usa `NewMyMemoryTranslator()` por defecto
  - `newEnricher(s, t)` constructor interno para inyectar mock en tests
  - Nuevo método `productKeywords(ctx, name)`: intenta traducir con el API; si falla, hace fallback al diccionario `catalanToSpanish`
  - `Run()` usa `productKeywords` en lugar de `translateCatalan` directamente
- `internal/enricher/translator_test.go` (nuevo): 11 tests — NoopTranslator, MyMemoryTranslator (éxito, caché, error HTTP, traducción vacía), `productKeywords` (usa translator, fallback al dict, non-catalan preservado)

**Diseño:** El diccionario manual (`catalan_dict.go`) se mantiene como fallback. Si la API de MyMemory no está disponible o falla, el enricher sigue funcionando sin degradación. La traducción ocurre a nivel de nombre completo (no token a token), lo que da mejor contexto al API de traducción.

### Tests finales

```
Backend: ok  basket-cost/internal/enricher  (todos pasan, incluidos 11 nuevos)
Frontend unit: 161 tests — 14 archivos — todos pasan
E2E: 122 passed, 2 skipped (columnas en móvil, esperado)
```
