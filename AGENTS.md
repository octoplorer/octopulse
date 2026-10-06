# Repository Guidelines

## Project Structure & Module Organization

Octopulse is a single-organization uptime platform: one Go service serves APIs and a Vue SPA.

- `cmd/octopulse/`: CLI entrypoint; `internal/`: probes, scheduling, storage, notifications, and HTTP handlers.
- `db/{sqlite,postgres}/`: migrations and SQL queries; `api/openapi.json`: generated API contract.
- `web/src/`: `pages/` routes, `components/` UI, `composables/` reactive state, `lib/` stateless helpers, and `locales/` translations. Styles live in `style.css`; built assets go to `web/dist/`.
- Tests sit beside source files. Consult `docs/development.md`, `docs/adr/`, and `GLOSSARY.md` for conventions and architecture.

## Build, Test, and Development Commands

Run from the repository root; use tools pinned in `mise.toml`.

- Setup: `mise trust`, `mise install`, `mise exec -- go mod download`, then `mise run install` (locked workspace dependencies).
- `mise run dev`: start API and frontend; open `http://127.0.0.1:5173/app`.
- `mise run check`: Go tests, frontend lint, type checking, and Vitest.
- `mise run test`: Go and frontend tests.
- `mise run build`: produce `bin/octopulse` and `web/dist/`.
- `mise exec -- go vet ./...`: Go static analysis.
- `mise exec -- aube run lint:fix`: fix frontend formatting and lint issues.

Root Node scripts operate on frontend tooling; mise orchestrates the whole project.

## Coding Style & Naming Conventions

Format Go with `gofmt` (tabs). Follow existing TypeScript/Vue style: two-space indentation, single quotes, no semicolons, PascalCase component filenames, and camelCase functions. ESLint with `@antfu/eslint-config` handles formatting and requires zero warnings. Keep `en.json` and `zh-CN.json` translation keys aligned.

## Testing Guidelines

Use Go's `testing` package in `*_test.go` and Vitest in `*.test.ts` or `*.spec.ts`. Add regression tests for changed behavior; no numeric coverage threshold is configured. Run `mise exec -- go test -race -p 1 ./...` for races. Database changes should pass SQLite and dedicated PostgreSQL suites; setup is documented in `docs/development.md`.

## Generated Code

Update source contracts, then run `mise run generate:db` or `mise run generate:web`. Include regenerated outputs; CI checks drift. Keep SQLite and PostgreSQL changes aligned.

## Commit & Pull Request Guidelines

Follow history's Conventional Commits, e.g. `feat(web): add monitor filters`. Use the same style for PR titles. Describe behavior, link relevant issues, report validation, and include screenshots for UI changes. Commit or push only when explicitly requested.
