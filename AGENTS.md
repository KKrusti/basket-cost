# AGENTS.md

Guidelines for AI coding agents working on the `basket-cost` project.

---

## Communication

Always communicate with the user in **Spanish**, regardless of the language used in code, comments, or commit messages.

---

## Skills

- Before modifying any **backend** code, load the `golang-pro` skill.
- Before modifying any **frontend** code, load **both** the `vercel-react-best-practices` skill and the `ui-ux-pro-max` skill. These two skills are complementary and must always be used together for frontend/UI work: `vercel-react-best-practices` governs React/Next.js performance patterns, while `ui-ux-pro-max` governs visual design, component styling, and UX decisions.
- Every time a skill is loaded, announce it in the response with the tag `[skill: <name>]` before writing any code.

---

## Project Overview

Two-tier SPA for tracking grocery prices in the Spanish market.

- **Backend:** Go 1.24, SQLite via `modernc.org/sqlite` (CGO-free), PDF parsing via `ledongthuc/pdf`. JSON REST API on `:8080`.
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
│   ├── go.mod                        # module: basket-cost, go 1.24, modernc.org/sqlite + ledongthuc/pdf
│   ├── basket-cost.db                # SQLite database (created on first run)
│   ├── seed/                         # sample Mercadona PDF receipts for seeding
│   ├── cmd/
│   │   ├── server/main.go            # entry point: routing, CORS middleware, ListenAndServe
│   │   ├── seed/main.go              # CLI: bulk-import PDF receipts into the DB
│   │   └── enrich/main.go            # CLI: download product images from the Mercadona API
│   └── internal/
│       ├── database/db.go            # SQLite connection, WAL pragmas, schema migrations
│       ├── models/models.go          # domain types: Product, PriceRecord, SearchResult
│       ├── store/
│       │   ├── store.go              # Store interface + SQLiteStore implementation
│       │   └── store_test.go
│       ├── handlers/
│       │   ├── handlers.go           # HTTP handlers: Search, Product, Ticket
│       │   └── handlers_test.go
│       ├── enricher/
│       │   ├── enricher.go           # orchestrates image-URL enrichment
│       │   ├── mercadona_client.go   # HTTP client for the public Mercadona catalogue API
│       │   └── enricher_test.go
│       └── ticket/
│           ├── model.go              # Ticket and TicketLine types
│           ├── extractor.go          # PDFExtractor interface + ledongthuc implementation
│           ├── parser.go             # MercadonaParser: PDF text → Ticket struct
│           ├── importer.go           # Importer: extract → parse → store.UpsertPriceRecord
│           ├── parser_test.go
│           └── importer_test.go
└── frontend/
    ├── vite.config.ts                # Vite + proxy /api→:8080, manual chunks, Vitest config
    ├── tsconfig.json                 # strict mode, noUnusedLocals, noUnusedParameters
    ├── package.json
    └── src/
        ├── main.tsx                  # ReactDOM.createRoot entry point
        ├── App.tsx                   # app shell: header with TicketUploader + SearchBar/ProductDetail
        ├── App.test.tsx
        ├── index.css                 # design system: CSS variables, all component styles
        ├── test/setup.ts             # Vitest global setup (@testing-library/jest-dom)
        ├── types/index.ts            # shared TypeScript interfaces
        ├── api/
        │   ├── products.ts           # fetch-based API client (search, getProduct, uploadTicket, uploadTickets)
        │   └── products.test.ts
        ├── components/
        │   ├── SearchBar.tsx         # search input with 300 ms debounce + result list
        │   ├── SearchBar.test.tsx
        │   ├── ProductBrowser.tsx    # full catalogue grid with page-size and column-count controls
        │   ├── ProductBrowser.test.tsx
        │   ├── ProductDetail.tsx     # product detail: Recharts line chart + price history table
        │   ├── ProductDetail.test.tsx
        │   ├── ProductImage.tsx      # product image with emoji fallback
        │   ├── ProductImage.test.tsx
        │   ├── TicketUploader.tsx    # PDF upload button (single or batch), uploading state, result toast
        │   └── TicketUploader.test.tsx
        └── utils/
            ├── productImages.ts      # static product-ID → image URL map + category emoji fallbacks
            └── productImages.test.ts
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/products?q=<query>` | Search products; empty `q` returns all |
| `GET` | `/api/products/<id>` | Full product detail with price history |
| `POST` | `/api/tickets` | Upload a single Mercadona PDF receipt (`multipart/form-data`, field `file`) |

### POST /api/tickets — request / response

**Request:** `multipart/form-data` with a single field named `file` containing a PDF (max 10 MB).

**Response 201:**
```json
{ "invoiceNumber": "4144-017-284404", "linesImported": 23 }
```

**Error responses:** `400` (missing field or bad form), `422` (PDF cannot be parsed), `500` (internal error).

---

## Frontend API Client (`src/api/products.ts`)

| Function | Description |
|----------|-------------|
| `searchProducts(query)` | `GET /api/products?q=<query>` |
| `getAllProducts()` | alias for `searchProducts('')` |
| `getProduct(id)` | `GET /api/products/<id>` |
| `uploadTicket(file)` | `POST /api/tickets` — single PDF |
| `uploadTickets(files)` | Runs `uploadTicket` for every file concurrently via `Promise.all`. Individual failures are captured; the batch never aborts. Returns a `TicketUploadSummary`. |

### TicketUploadSummary shape

```ts
interface TicketUploadSummary {
  total: number;
  succeeded: number;
  failed: number;
  items: TicketUploadItem[];   // discriminated union: ok/error per file
}
```

---

## Testing Policy

**Tests are mandatory.** Every new piece of code — backend or frontend — must have corresponding unit tests written alongside it. This is non-negotiable.

- **Backend (Go):** new functions in `store` or `handlers` require a `_test.go` file in the same package. Use `net/http/httptest` for handler tests.
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
- `internal/store` is the data/repository layer (`Store` interface + `SQLiteStore` implementation).
- `internal/handlers` is the HTTP layer only — delegate logic to other packages.
- `internal/ticket` owns the full PDF import pipeline: extract → parse → persist.
- `internal/enricher` handles product image enrichment from the Mercadona public API.

### Naming

- Exported identifiers: `PascalCase`. Unexported: `camelCase`.
- Package names: lowercase single words (`handlers`, `models`, `store`, `ticket`, `enricher`).
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
- CORS is handled by a hand-rolled middleware in `cmd/server/main.go` — do not add external CORS libraries.

### Struct Tags

- All exported struct fields must have `json:"..."` tags.
- Use `omitempty` on optional/nullable fields.

### Database

- Driver: `modernc.org/sqlite` (CGO-free, pure Go).
- `SetMaxOpenConns(1)` to avoid WAL write contention.
- PRAGMAs set at open time: `foreign_keys=ON`, `journal_mode=WAL`, `synchronous=NORMAL`, `cache_size=-16000`.
- Schema changes go through the migration table in `internal/database/db.go` — never ALTER TABLE outside a migration.
- Product IDs are derived slugs (`slugify(name)`) generated in `store.UpsertPriceRecord`; do not generate IDs in handlers.

---

## TypeScript / React Style Guidelines

### Components

- Function components only — no class components.
- Export pattern: `export default function ComponentName(props: ComponentNameProps)`.
- Define props with a local `interface ComponentNameProps { ... }` in the same file.
- Shared domain types live in `src/types/index.ts` as `interface` (not `type` aliases), except for discriminated unions which use `type`.
- SVG icons are defined as small inline function components within the same file — do not use emoji as UI icons.

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
- For concurrent independent requests (e.g. batch uploads), use `Promise.all` — do not chain sequentially.

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
