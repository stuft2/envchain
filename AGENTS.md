# AGENTS.md

### Do
- use **Go** for all backend logic
- use the **standard library first** (net/http, html/template, context, database/sql)
- use **HTMX** for interactivity and progressive enhancement (server-rendered HTML first)
- use **Pico.css** for styling; extend via custom CSS/SCSS only when needed
- keep handlers small and focused (one responsibility per handler)
- return **HTML fragments** for HTMX requests and full pages for normal requests
- default to **small components/templates**
- default to **small diffs**

### Don't
- do not build a SPA or client-side state machine
- do not introduce frontend frameworks (React, Vue, etc.)
- do not hard-code inline styles when a semantic Pico class or token exists
- do not fetch data directly in templates
- do not add heavy Go dependencies without approval

---

### Commands
All commands are run via **go-task** (Taskfile.yml). Use **only** the tasks defined there.

# preferred fast checks
task fmt            # Rewrite files with gofmt -s -w
task lint           # Format check + vet
task test           # go test (race + short)

# coverage (after running tests with -cover)
task test -- -cover
task coverage       # Generate coverage reports from .coverage/coverage.out

# full build / dependency verification
task build          # Build all binaries
task ci             # Verify and install dependencies

# auth / tooling
task vault:login    # Login to BYU OIT's Hashicorp Vault

---

### Safety and permissions

Allowed without prompt:
- read files, list files
- task fmt, lint, test
- go test runs only when invoked via `task test`
- task coverage (after tests generate .coverage/coverage.out)

Ask first:
- adding Go modules / new dependencies
- deleting files or changing permissions
- database migrations
- running `task build` or `task ci` unless explicitly requested
- running `task vault:login`

---

### Project structure
- entrypoint: `cmd/app/main.go`
- HTTP routing: `internal/web/routes.go`
- handlers: `internal/web/handlers/*`
- templates: `internal/web/templates`
  - pages: full-page templates
  - fragments: HTMX partials
- static assets: `web/static`
- SCSS / CSS sources: `web/scss`
- database layer: `internal/db`
- migrations: `migrations/`

---

### Good and bad examples
- avoid large, multi-purpose handlers like `admin.go`
- prefer small handlers like `projects_list.go`, `projects_update.go`
- templates:
  - pages: `templates/pages/projects.html`
  - fragments: `templates/fragments/project_row.html`
- HTMX:
  - return HTML fragments, not JSON, unless explicitly required
  - use `HX-Request` header to branch behavior
- data access:
  - query data in handlers or repositories
  - never access the database directly from templates

---

### HTMX conventions
- detect HTMX via `HX-Request: true`
- POST/PUT/DELETE:
  - validate input
  - enforce CSRF
  - return fragments or `204 No Content`
- redirects:
  - use `HX-Redirect` header for HTMX flows
- avoid client-side state; let the server be the source of truth

---

### Styling system
- Pico.css provides the base design system
- use semantic HTML elements first (`main`, `section`, `article`, `nav`)
- extend Pico with:
  - CSS variables
  - small, well-scoped custom classes
- no inline styles unless unavoidable
- SCSS should compile to `web/static/css`

---

### API usage
- server-rendered HTML is the default interface
- JSON APIs are allowed only when necessary
- keep API logic in handlers or service layers
- document non-HTML endpoints in `./api/docs/*.md`

---

### Change workflow (required)

For every feature, bug fix, or doc change:

1. sync `main` before starting work
2. create a separate worktree and branch off `main` (`codex/<short-topic>`)
3. define 1-3 acceptance criteria
4. write/update tests first (for bugs, start with a failing repro test)
5. implement the smallest change that satisfies acceptance criteria
6. run quality gates:
   - `task fmt`
   - `task lint`
   - `task test`
7. verify staged diff (`git diff --staged`) and confirm no secrets or unrelated files are included
8. commit using Conventional Commits (imperative mood), keeping commits small and logical
9. open a PR with:
   - summary of change
   - test evidence (commands run + results)
   - risk/rollback notes for behavior changes

---

### PR checklist
- fmt: `task fmt`
- lint: `task lint`
- tests: `task test` (add coverage for new paths)
- HTMX responses verified manually
- diff is small and focused with a clear summary

---

### When stuck
- ask one clarifying question
- propose a short plan before implementing
- prefer a small draft PR over a large speculative change

---

### Test-first mode
- write or update tests for new behavior
- make tests pass with the smallest possible change

---

### Philosophy
- server-rendered first
- progressive enhancement, not JS-heavy UIs
- boring, readable Go over clever abstractions
- simple tools, explicit behavior
