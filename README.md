# basket-cost

A grocery price tracker for the Spanish market — built as a playground for experimenting with AI coding agents ([OpenCode](https://opencode.ai) and [Claude Code](https://claude.ai/code)).

Upload Mercadona PDF receipts to populate the database, then search for any product to see its current price and a historical evolution chart.

---

## What this actually is

This repo is a **sandbox for vibe-coding with AI agents**. Every feature, refactor, test, and config tweak in this codebase was driven through natural-language prompts to the agent. The project is intentionally contained so the interesting part is watching the agent navigate a real two-tier codebase — routing, a SQLite-backed API, React components, PDF parsing, JWT auth, tests, a task runner — not the app itself.

If you want to play with an AI coding agent on a project that has real structure without building something from scratch, clone this and start prompting.

---

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.24 · SQLite (`modernc.org/sqlite`, CGO-free) · PDF parsing (`ledongthuc/pdf`) · JWT auth (`golang-jwt/jwt`) |
| Frontend | React 18 · TypeScript · Vite · Recharts |
| Tests | Go `testing` + `net/http/httptest` · Vitest + Testing Library · Playwright (E2E) |
| Task runner | [go-task](https://taskfile.dev) |

---

## Features

- **Upload tickets** — import one or several Mercadona PDF receipts at once. Each file is processed independently; partial failures are reported per-file without aborting the batch.
- **Search products** — live search with 300 ms debounce across your catalogue.
- **Browse catalogue** — grid view of all products with configurable page size and column count.
- **Price history** — interactive line chart plus a full price table for any selected product, with a badge showing overall price change since first purchase.
- **Analytics** — top products by purchase frequency and biggest price increases over time.
- **Product images** — enriched from the public Mercadona catalogue API; falls back to a category emoji when unavailable. Supports manual image URL override.
- **User accounts** — register and log in to keep your data private. Anonymous mode is also supported (data shared under a global namespace).

---

## Project structure

```
basket-cost/
├── Taskfile.yml
├── backend/
│   ├── go.mod                        # module: basket-cost, go 1.24
│   ├── basket-cost.db                # SQLite database (created on first run)
│   ├── seed/                         # sample Mercadona PDF receipts
│   ├── cmd/
│   │   ├── server/main.go            # entry point: routing, middleware chain, ListenAndServe
│   │   ├── seed/main.go              # CLI: bulk-import PDF receipts into the DB
│   │   └── enrich/main.go            # CLI: download product images from Mercadona API
│   └── internal/
│       ├── auth/                     # bcrypt password hashing + HS256 JWT (72 h TTL)
│       ├── database/db.go            # SQLite connection, WAL pragmas, schema migrations
│       ├── models/models.go          # domain types: User, Product, PriceRecord, SearchResult…
│       ├── store/                    # Store interface + SQLiteStore (multi-tenant, user_id scoped)
│       ├── handlers/                 # HTTP handlers (Auth, Search, Product, Ticket, Analytics) + tests
│       ├── enricher/                 # image-URL enrichment from Mercadona public API
│       └── ticket/                   # PDF import pipeline: extract → parse → persist
└── frontend/
    └── src/
        ├── App.tsx                   # app shell: header, tabs (Productos / Analítica), auth state
        ├── index.css                 # design system: CSS variables, all component styles
        ├── types/index.ts            # shared TypeScript interfaces
        ├── api/products.ts           # fetch-based API client (auth headers, timeouts)
        ├── components/
        │   ├── LoginModal.tsx        # register / login modal
        │   ├── TicketUploader.tsx    # PDF upload button, batch support, progress bar, result toast
        │   ├── SearchBar.tsx         # search input with 300 ms debounce + result list
        │   ├── ProductBrowser.tsx    # full catalogue grid with pagination
        │   ├── ProductDetail.tsx     # price history chart + table + PriceChangeBadge
        │   ├── ProductImage.tsx      # image with emoji fallback
        │   ├── Analytics.tsx         # top purchased + biggest price increases
        │   └── ...                   # co-located *.test.tsx for every component
        └── utils/productImages.ts    # static image URL map + category emoji fallbacks
```

---

## Running it

```bash
# Start everything
task dev
# → backend:  http://localhost:8080
# → frontend: http://localhost:5173
```

To seed the database with the bundled sample receipts:

```bash
cd backend && go run ./cmd/seed/main.go -dir ./seed
```

---

## API

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/auth/register` | Create a new user account |
| `POST` | `/api/auth/login` | Authenticate and receive a JWT |
| `GET` | `/api/products?q=<query>` | Search products (scoped to authenticated user); empty `q` returns all |
| `GET` | `/api/products/<id>` | Full product detail with price history |
| `PATCH` | `/api/products/<id>/image` | Set a manual image URL for a product |
| `POST` | `/api/tickets` | Upload a Mercadona PDF receipt (`multipart/form-data`, field `file`, max 10 MB) |
| `GET` | `/api/analytics` | Top purchased products and biggest price increases for the authenticated user |

All endpoints accept an optional `Authorization: Bearer <token>` header. Requests without a valid token are served in anonymous mode (data shared under a `user_id = NULL` namespace).

The frontend uploads multiple files by calling `POST /api/tickets` once per file in parallel via `Promise.all`. There is no dedicated batch endpoint.

---

## Tests

```bash
task test:backend   # Go tests
task test:frontend  # Vitest
task test           # both
cd frontend && npm run test:e2e  # Playwright E2E
```

---

## Notes

- Product names, categories and store names are in Spanish (the app targets the Spanish market).
- All other code, comments, and identifiers are in English.
- See `AGENTS.md` for the full coding guidelines used to prompt the agent.
