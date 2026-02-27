# AGENTS.md

Guidelines for AI coding agents working on the `basket-cost` project.

---

## Project Overview

Two-tier SPA for tracking grocery prices in the Spanish market.

- **Backend:** Go 1.21, pure stdlib, no external dependencies. JSON REST API on `:8080`.
- **Frontend:** React 18 + TypeScript + Vite, on `:5173`. Proxies `/api` to the backend.
- **Task runner:** `task` (go-task). Use `Taskfile.yml` at the repo root.

---

## Commands

### Run the project

```bash
task dev          # backend + frontend in parallel (full dev stack)
task backend      # backend only → http://localhost:8080
```

### Frontend only

```bash
cd frontend
npm run dev       # dev server → http://localhost:5173
npm run build     # type-check (tsc) + production bundle (vite)
npm run preview   # serve the production build
```

### Backend only

```bash
cd backend
go run ./cmd/server/main.go
```

### Type-check frontend (no emit)

```bash
cd frontend && npx tsc --noEmit
```

### Run Go tests

```bash
cd backend && go test ./...                  # all packages
cd backend && go test ./internal/handlers/   # single package
cd backend && go test -run TestFunctionName ./internal/handlers/  # single test
```

### Run frontend tests

```bash
cd frontend && npm test              # run all tests once (Vitest)
cd frontend && npm run test:watch    # watch mode
cd frontend && npm run test:coverage # coverage report
```

---

## Project Structure

```
basket-cost/
├── Taskfile.yml
├── backend/
│   ├── go.mod                        # module: basket-cost, go 1.21, no external deps
│   ├── cmd/server/main.go            # entry point: routing, CORS middleware, ListenAndServe
│   └── internal/
│       ├── handlers/
│       │   ├── handlers.go           # HTTP handlers
│       │   └── handlers_test.go      # handler tests
│       ├── models/models.go          # domain types (pure structs)
│       └── mockdata/
│           ├── data.go               # in-memory data + query/lookup logic
│           └── data_test.go          # mockdata tests
└── frontend/
    ├── vite.config.ts
    ├── tsconfig.json
    └── src/
        ├── main.tsx / App.tsx
        ├── App.test.tsx
        ├── test/setup.ts             # Vitest global setup (@testing-library/jest-dom)
        ├── api/
        │   ├── products.ts           # fetch-based API client
        │   └── products.test.ts
        ├── components/               # React components (each with a co-located *.test.tsx)
        └── types/index.ts            # shared TypeScript interfaces
```

---

## Testing Policy

**Tests are mandatory.** Every new piece of code — backend or frontend — must have corresponding unit tests written alongside it. This is non-negotiable.

- **Backend (Go):** new functions in `mockdata` or `handlers` require a `_test.go` file in the same package. Use `net/http/httptest` for handler tests.
- **Frontend (TypeScript/React):** every new component must have a co-located `*.test.tsx` file; every new API function must have a corresponding test in `*.test.ts`. Use Vitest + `@testing-library/react`.
- Third-party libraries that do not render in jsdom (e.g. `recharts`) must be mocked with `vi.mock(...)`.
- When querying DOM elements, prefer semantic queries (`getByRole`, `getByLabelText`). Fall back to `document.querySelector` with a CSS class selector only when the same text appears in multiple nodes.
- After implementing any feature or fix, run both test suites and confirm they pass before considering the task done:

```bash
cd backend && go test ./...
cd frontend && npm test
```

---

## Go Style Guidelines

### Structure

- Follow strict `cmd/` + `internal/` layout. Never place business logic in `cmd/`.
- `cmd/server/main.go` is wiring only: routing, middleware, `http.ListenAndServe`.
- `internal/models` contains pure data structs — no logic, no methods.
- `internal/mockdata` acts as the data/repository layer for now.
- `internal/handlers` is the HTTP layer only — delegate logic to other packages.

### Naming

- Exported identifiers: `PascalCase`. Unexported: `camelCase`.
- Package names: lowercase single words (`handlers`, `models`, `mockdata`).
- Acronyms follow Go convention: `GetProductByID`, not `GetProductById`.
- No Hungarian notation or type suffixes on variables.

### Imports

- Single grouped `import (...)` block. No blank-line separation between stdlib and internal packages.
- Prefer stdlib `errors` and `context` — do not use `github.com/pkg/errors` or `golang.org/x/net/context`.
- Keep imports sorted: stdlib first, then internal (goimports order).

### Error Handling

- Return errors immediately — guard clause style, early returns.
- In HTTP handlers, use `http.Error(w, message, statusCode)` and `return`.
- In `main()`, use `log.Fatal` for startup failures.
- No custom error types or sentinel errors unless the package genuinely needs them.
- Do not swallow errors silently.

```go
// Correct pattern
if r.Method != http.MethodGet {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
}
```

### HTTP Responses

- Always set `Content-Type: application/json` before writing the body.
- Use `json.NewEncoder(w).Encode(v)` for JSON responses (idiomatic, streams output).
- CORS is handled by a hand-rolled middleware — do not add external CORS libraries.

### Struct Tags

- All exported struct fields must have `json:"..."` tags.
- Use `omitempty` on optional/nullable fields.

---

## TypeScript / React Style Guidelines

### Components

- Function components only — no class components.
- Export pattern: `export default function ComponentName(props: ComponentNameProps)`.
- Define props with a local `interface ComponentNameProps { ... }` in the same file.
- Shared domain types live in `src/types/index.ts` as `interface` (not `type` aliases).

### Imports

- Use `import type { ... }` for type-only imports (`isolatedModules: true` is enforced).
- Use plain `import { ... }` for runtime values.
- Use relative paths — no path aliases are configured.
- Group: React/external libs first, then internal modules.

```ts
import { useState, useEffect } from 'react';
import { searchProducts } from '../api/products';
import type { SearchResult } from '../types';
```

### Async / Data Fetching

- API calls use raw `fetch` — no axios or other HTTP libraries.
- The API client (`src/api/products.ts`) exports plain `async` functions. It throws `Error` on non-OK responses.
- In components, use a cancellation flag inside `useEffect` to avoid state updates on unmounted components:

```ts
useEffect(() => {
  let cancelled = false;
  getProduct(id).then((data) => { if (!cancelled) setData(data); });
  return () => { cancelled = true; };
}, [id]);
```

- Debounce via `setTimeout`/`clearTimeout` inside `useEffect` — no external debounce library.

### State Management

- Local `useState` only. No Redux, Zustand, Context API, or any external state library.

### TypeScript Strictness

The following compiler flags are enabled — all code must satisfy them:

- `strict: true` (all strict checks)
- `noUnusedLocals: true`
- `noUnusedParameters: true`
- `noFallthroughCasesInSwitch: true`
- `noUncheckedSideEffectImports: true`

Do not disable these flags or use `@ts-ignore` / `@ts-expect-error` without a clear comment explaining why.

### Formatting Helpers

Inline arrow functions within components are acceptable for simple formatters (`formatPrice`, `formatDate`). Extract to a shared `src/utils/` module only when used in more than one component.

---

## Language

- UI text, labels, and user-facing strings must be in **Spanish** — the app targets the Spanish grocery market.
- Code identifiers, comments, and commit messages may be in English or Spanish; be consistent within a file.

---

## WSL / Windows Note

`node_modules` is installed in a Linux-native path (`/home/carlos/.npm-workspaces/basket-cost/node_modules`) and symlinked into `frontend/node_modules` to avoid NTFS `chmod` errors. This is handled automatically by `task dev:deps`. Do not delete or recreate `frontend/node_modules` as a real directory — keep it as a symlink.
