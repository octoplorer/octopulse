# Octopulse implementation plan

Objective: implement the complete platform specified in docs/v1-spec.md, docs/monitor-options.md, docs/status-page-policy.md and docs/database-design.md. Do not narrow completion to scaffolding or a passing subset.

Authorization: user requested implementation, a new branch, and incremental Conventional Commits. Branch: feat/uptime-platform. Do not push.

## Phases

1. [complete] Reproducible mise/aube toolchain, Go/Vue project, dependency contracts and build tasks; managed tools and locked dependencies installed and used for builds.
2. [complete] SQLite/PostgreSQL persistence, accounts/sessions/roles/secrets/audit and server bootstrap are committed and verified by the full SQLite and real PostgreSQL18.6 race suites, including actual backup tools and session/lifecycle regressions.
3. [complete] HTTP/TCP/DNS/certificate probes, heartbeat, scheduler, budgets/retries/thresholds/pause/maintenance and crash gaps plus accepted-check/scheduling fixes are verified in both full race suites and both real-time capacity reruns.
4. [complete] Time-weighted availability/coverage, retention and aggregates plus durable Shoutrrr delivery rules are committed and verified on both databases, including final API history and maintenance regressions.
5. [complete] Public pages, independent customization/draft/preview/publication, incidents/maintenance and domain/path routing are committed and verified by both full race suites and browser domain/path/publication workflows.
6. [complete] Independent Beszel adapter is committed; normal readonly authentication, refresh, summaries/history/containers, units and explicit stale/auth/version failures are verified against fixtures and released Hub 0.20.0.
7. [complete] CSR admin/public UI, generated HeyAPI/Colada contracts, final types/format/8 unit tests/production build and desktop/mobile browser workflows passed, including readonly/domain/Beszel/statistics/publication and canonical visit-link optional-domain boundaries.
8. [complete] Full real PostgreSQL/file SQLite race suites, both normal/race 100-monitor/30-second workloads, vet, sqlc/OpenAPI/HeyAPI zero drift, final arm64 image, both fresh Compose startups and both restart/encrypted-secret/PNG/complete backup-restore workflows passed. Requirement audit and delivery docs complete; local architecture evidence is arm64 only. No remote CI run or push.

## Delivery status

Implementation, local acceptance and the authorized incremental delivery/documentation commits are complete. All changes are on feat/uptime-platform; no remote push or CI execution occurred. Local architecture evidence covers Darwin arm64 and Linux arm64 containers.

## Committed phases

- b2fdc17 feat(core): initialize managed toolchain and secure application foundation
- cf0bdcd feat(storage): add transactional SQLite and PostgreSQL persistence
- c64f32f feat(monitoring): implement probes and transactional monitoring engine
- b9d5153 fix(storage): serialize monitor updates and preserve aggregate outcomes
- 6c54c5e feat(notifications): add durable Shoutrrr delivery and recovery ordering
- 0fa0d82 feat(integrations): display independent Beszel server metrics
- 5adfb79 feat(statistics): calculate weighted availability and retain history aggregates
- 057232d fix(monitoring): preserve accepted checks and serialize scheduling completion
- c3164c4 feat(api): expose secured administration and published status pages
- 3bee56d test(capacity): verify 100 monitors against both databases
- 18e44c0 feat(web): add monitoring console and customizable status pages
- 7d088aa feat(delivery): package binary and Docker deployments with verification CI
- Final documentation commit: docs: record completed implementation and acceptance evidence

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

Tool discovery initially found no aube or Docker executable. aube was installed via mise; PostgreSQL and Docker/Colima validation used isolated test resources. Actual arm64 image and both database delivery/restore workflows passed. Cleanup targets only this task's test containers, volumes, PostgreSQL cluster and named Colima profile.

- A patch adding certificate notification fields did not match a gofmt-aligned line. Verified it made no partial edits, then reapplied against the current line.
- A combined API patch was rejected for two operations targeting server.go; split the operations into distinct verified patches.
