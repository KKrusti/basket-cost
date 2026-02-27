# basket-cost

A grocery price tracker for the Spanish market — built as a playground for experimenting with [OpenCode](https://opencode.ai), the AI coding agent.

Upload Mercadona PDF receipts to populate the database, then search for any product to see its current price and a historical evolution chart.

---

## What this actually is

This repo is a **sandbox for vibe-coding with OpenCode**. Every feature, refactor, test, and config tweak in this codebase was driven through natural-language prompts to the agent. The project is intentionally contained so the interesting part is watching the agent navigate a real two-tier codebase — routing, a SQLite-backed API, React components, PDF parsing, tests, a task runner, WSL quirks — not the app itself.

If you want to play with OpenCode on a project that has real structure without building something from scratch, clone this and start prompting.

---

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.24 · SQLite (`modernc.org/sqlite`, CGO-free) · PDF parsing (`ledongthuc/pdf`) |
| Frontend | React 18 · TypeScript · Vite |
| Tests | Go `testing` + `net/http/httptest` · Vitest + Testing Library |
| Task runner | [go-task](https://taskfile.dev) |

---

## Features

- **Upload tickets** — drag the PDF button in the header to import one or several Mercadona receipts at once. Each file is processed independently; partial failures are reported per-file without aborting the batch.
- **Search products** — live search with 300 ms debounce across everything in the database.
- **Browse catalogue** — grid view of all products with configurable page size and column count.
- **Price history** — interactive line chart (Recharts) plus a full price table for any selected product.
- **Product images** — enriched from the public Mercadona catalogue API; falls back to a category emoji when unavailable.

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
│   │   ├── server/main.go            # entry point: routing, CORS, ListenAndServe
│   │   ├── seed/main.go              # CLI: bulk-import PDF receipts into the DB
│   │   └── enrich/main.go            # CLI: download product images from Mercadona API
│   └── internal/
│       ├── database/db.go            # SQLite connection, WAL pragmas, schema migrations
│       ├── models/models.go          # domain types: Product, PriceRecord, SearchResult
│       ├── store/                    # Store interface + SQLiteStore implementation
│       ├── handlers/                 # HTTP handlers (Search, Product, Ticket) + tests
│       ├── enricher/                 # image-URL enrichment from Mercadona public API
│       └── ticket/                   # PDF import pipeline: extract → parse → persist
└── frontend/
    └── src/
        ├── App.tsx                   # app shell: header with TicketUploader + main view
        ├── index.css                 # design system: CSS variables, all component styles
        ├── types/index.ts            # shared TypeScript interfaces
        ├── api/products.ts           # fetch-based API client
        ├── components/
        │   ├── TicketUploader.tsx    # PDF upload button, batch support, result toast
        │   ├── SearchBar.tsx         # search input + result list
        │   ├── ProductBrowser.tsx    # full catalogue grid with pagination
        │   ├── ProductDetail.tsx     # price history chart + table
        │   ├── ProductImage.tsx      # image with emoji fallback
        │   └── ...                   # co-located *.test.tsx for every component
        └── utils/productImages.ts    # static image URL map + category emoji fallbacks
```

---

## Running it

```bash
# Install frontend deps (handles the WSL/NTFS symlink)
task dev:deps

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
| `GET` | `/api/products?q=<query>` | Search products; empty `q` returns all |
| `GET` | `/api/products/<id>` | Full product detail with price history |
| `POST` | `/api/tickets` | Upload a Mercadona PDF receipt (`multipart/form-data`, field `file`, max 10 MB) |

The frontend uploads multiple files by calling `POST /api/tickets` once per file in parallel via `Promise.all`. There is no dedicated batch endpoint — concurrency is handled entirely on the client.

---

## Tests

```bash
task test:backend   # Go tests
task test:frontend  # Vitest
task test           # both
```

---

## Notes

- Product names, categories and store names are in Spanish (the app targets the Spanish market).
- All other code, comments, and identifiers are in English.
- `frontend/node_modules` is a symlink to a Linux-native path to avoid NTFS `chmod` errors on WSL. Don't delete it.
- See `AGENTS.md` for the full coding guidelines used to prompt the agent.
