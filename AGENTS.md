# Repository Guidelines

## Project Structure & Module Organization

Octopulse is a single-organization uptime platform: one Go service serves APIs and a Vue SPA.

- `cmd/octopulse/`: CLI entrypoint; `internal/`: probes, scheduling, storage, notifications, and HTTP handlers.
- `db/{sqlite,postgres}/`: migrations and SQL queries; `api/openapi.json`: generated API contract.
- `web/src/`: `pages/` routes, `components/` UI, `composables/` reactive state, `lib/` stateless helpers, and `locales/` translations. Styles live in `style.css`; built assets go to `web/dist/`.
- Tests sit beside source files. Consult `docs/development.md`, `docs/adr/`, and `GLOSSARY.md` for conventions and architecture.

## Build, Test, and Development Commands

Run from the repository root; use tool versions pinned in `mise.toml` and platform downloads locked in `mise.lock`.

- Setup: `mise trust`, `mise install`, then `mise run install` (Go modules and locked workspace dependencies in parallel).
- `mise run dev`: start API and frontend; open `http://127.0.0.1:5173/app`.
- `mise run check`: Go tests and vet, parallel frontend lint/type checking/Vitest, and hk configuration validation.
- `mise run test`: Go and frontend tests.
- `mise run build`: produce `bin/octopulse` and `web/dist/`.
- `mise install`: install pinned tools and hk Git hooks via the `postinstall` hook.
- `mise exec -- hk check --all` / `mise exec -- hk fix --all`: check or format files with hk; pre-commit fixes staged Go and frontend files.
- `mise run vet:go`: Go static analysis.
- `mise run lint:fix:web`: fix frontend formatting and lint issues.
- `mise run test:race:go`: Go race tests with serial package execution, no test cache, and a ten-minute package timeout.

Root Node scripts operate on frontend tooling; mise orchestrates the whole project. Frontend tasks prepare dependencies through `install:web` and run scripts with `--no-install`. Shared `install:web` dependencies run once per task graph; aube's frozen install skips dependencies that are already up to date. No tasks use mise source/output freshness caching, so checks always run. When updating tools, update `mise.toml` and regenerate `mise.lock` for `linux-x64,linux-arm64,macos-arm64`; hk updates also require both versioned Pkl imports in `hk.pkl` to match.

## Coding Style & Naming Conventions

Format Go with `gofmt` (tabs). Follow existing TypeScript/Vue style: two-space indentation, single quotes, no semicolons, PascalCase component filenames, and camelCase functions. ESLint with `@antfu/eslint-config` handles formatting and requires zero warnings. Keep `en.json` and `zh-CN.json` translation keys aligned.

## Testing Guidelines

Use Go's `testing` package in `*_test.go` and Vitest in `*.test.ts` or `*.spec.ts`. Add regression tests for changed behavior; no numeric coverage threshold is configured. Run `mise run test:race:go` for races. Database changes should pass SQLite and dedicated PostgreSQL suites; setup is documented in `docs/development.md`.

## Generated Code

Update source contracts, then run `mise run generate:db` or `mise run generate:web`; `mise run generate` runs both and waits for database generation before exporting OpenAPI. Standalone `generate:api` does not trigger database generation. Include regenerated outputs; CI checks drift. Keep SQLite and PostgreSQL changes aligned.

## Commit & Pull Request Guidelines

Follow history's Conventional Commits, e.g. `feat(web): add monitor filters`. Use the same style for PR titles. Describe behavior, link relevant issues, report validation, and include screenshots for UI changes. Commit or push only when explicitly requested.
