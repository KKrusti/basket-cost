# basket-cost

A toy grocery price tracker for the Spanish market — built as a playground for experimenting with [OpenCode](https://opencode.ai), the AI coding agent.

The idea is simple: you search for a product, and the app shows you its current price plus a history chart so you can see how it has evolved over time. The data is all mocked (no real scraping), so the focus is entirely on the code rather than the content.

---

## What this actually is

This repo is a **sandbox for vibe-coding with OpenCode**. Every feature, refactor, test, and config tweak in this codebase was driven through natural-language prompts to the agent. The project is intentionally simple so the interesting part is watching the agent navigate a real two-tier codebase — not the app itself.

If you want to play with OpenCode on a project that has real structure (routing, API layer, React components, tests, a task runner, WSL quirks…) without building something from scratch, clone this and start prompting.

---

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.21, pure stdlib, no external deps |
| Frontend | React 18 + TypeScript + Vite |
| Tests | Go testing + Vitest + Testing Library |
| Task runner | [go-task](https://taskfile.dev) |

---

## Project structure

```
basket-cost/
├── Taskfile.yml
├── backend/
│   ├── cmd/server/main.go          # entry point
│   └── internal/
│       ├── handlers/               # HTTP handlers + tests
│       ├── mockdata/               # in-memory data + tests
│       └── models/                 # domain types
└── frontend/
    └── src/
        ├── api/                    # fetch-based API client + tests
        ├── components/             # React components + tests
        └── types/                  # shared TypeScript interfaces
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
