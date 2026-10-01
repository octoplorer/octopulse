# Octopulse implementation plan

Objective: implement the complete platform specified in docs/v1-spec.md, docs/monitor-options.md, docs/status-page-policy.md and docs/database-design.md. Do not narrow completion to scaffolding or a passing subset.

Authorization: user requested implementation, a new branch, and incremental Conventional Commits. Branch: feat/uptime-platform. Do not push.

## Phases

1. [in_progress] Reproducible mise/aube toolchain, Go/Vue project, dependency contracts and build tasks.
2. [pending] SQLite/PostgreSQL migrations, typed persistence, accounts/sessions/roles/secrets/audit and server bootstrap.
3. [pending] HTTP/TCP/DNS/certificate probes, heartbeat, scheduler, budgets/retries/thresholds/pause/maintenance and crash gaps.
4. [pending] Time-weighted availability/coverage, retention and aggregates; durable Shoutrrr deliveries and notification rules.
5. [pending] Public status pages, customization/draft/preview/publication, incidents/maintenance and domain/path routing.
6. [pending] Independent Beszel adapter, summaries/history/containers and stale/failed/auth handling.
7. [pending] Complete CSR admin/public UI, HeyAPI/Pinia Colada, Ark UI, UnoCSS attributify, zh-CN/en, timezones and responsive light/dark.
8. [pending] Cross-database tests, browser verification, 100-monitor/30-second load, Docker/binary delivery, backup/restore docs and requirement-by-requirement completion audit.

## Next Step

Establish package/domain interfaces and delegate independent probe, persistence and frontend work; finish and verify the foundation before its first implementation commit.

## Completion gates

- Every explicit first-release option has a working API and applicable UI; no fake metrics or placeholder workflows.
- Same business tests run against real PostgreSQL and file SQLite.
- Round/state/outbox atomically commit; secrets never read back; roles and CSRF enforced server-side.
- Availability uses confirmed intervals, excludes unknown/pause/maintenance, preserves coverage and crash gaps.
- Public projections expose only published whitelisted fields and both path/host routes work.
- Beszel remains independent from uptime; no GitHub themes.
- Frontend typecheck/build and browser interactions pass; Go tests/race checks pass as applicable.
- Deployment and measured capacity verified, with evidence recorded.

## Errors

Tool discovery found no aube or Docker executable; aube installed via mise and real PostgreSQL18.6 isolated test cluster created. Docker delivery still requires verification.

- A patch adding certificate notification fields did not match a gofmt-aligned line. Verified it made no partial edits, then reapplied against the current line.
