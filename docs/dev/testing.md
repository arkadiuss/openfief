# Testing

Rules without a marker are the author's; **(P)** marks a Claude Code proposal pending approval. Every test must be stable and runnable locally.

## Levels

| Level | What | How | Runs in |
|---|---|---|---|
| Unit | Domain logic | Pure functions, table-driven, small cases | `task test` (pre-commit, CI) |
| Integration | Adapters against real protocols | HTTP: [hypert](https://github.com/areknoster/hypert) record/replay. Postgres: a disposable instance via testcontainers-go **(P)** | `task test` (replay only) |
| End-to-end | The full flow on Kubernetes | Local k3s (native, or k3d) | `task e2e` (locally; CI on k3s) |

## Unit tests

- Domain code is pure functions ([code-organization.md](code-organization.md)), so unit tests need no mocks.
- Table-driven: a `[]struct{ name string; ... }` table plus `t.Run(tc.name, ...)` and `t.Parallel()`.
- One behavior per table. Keep cases small and named after the rule they check, in the ubiquitous language.
- Compare with `go-cmp` (`cmp.Diff`) and print `got`/`want`. Stdlib `testing` only, no assertion framework **(P)**.
- Ledger invariants (postings sum to zero per currency; FIFO consumes the oldest lot first) also get property-based tests **(P)**.
- Parsers (e.g. the XTB export) use golden files built from synthetic exports.

## Integration tests: hypert

- Build HTTP clients with `hypert.TestClient(t, record, opts...)`. Committed tests always run with `record = false` (replay), so they're offline and deterministic.
- Recordings live in `testdata/<TestName>/` as `*.req.http` and `*.resp.http` and are committed.
- Replay validates each request against the recording; keep `DefaultRequestValidator()`.
- Re-recording is a deliberate, manual act: follow the `hypert-recording` skill (`.claude/skills/hypert-recording/SKILL.md`).
- hypert's sanitizers clean **requests only**. Responses must be scrubbed with `WithResponseTransform` in record mode, or recorded only against paper, demo or public endpoints.
- No real account data, ever: a recording must not contain an account number, holding, amount or token from a real account (GR-12).
- Postgres adapter tests run against a fresh database with all migrations applied **(P)**. Never share state between tests.

## End-to-end tests

- These run the real system on a **local** k3s: the bootstrap install, the operator, plugins, and a real Ghostfolio in the cluster.
- External institutions are fixture servers in the cluster; e2e never calls a real broker.
- `task e2e` refuses any kubeconfig context that isn't local **(P)**.
- Each e2e test maps to an M0 acceptance criterion (SPEC.md §7) or a kibi scenario.

## Stability

- No sleeps: wait on conditions with a deadline.
- No wall-clock: inject the date.
- No unseeded randomness.
- No network in unit or replay tests. No test-order dependence.
- A flaky test is a bug. Fix the cause. Never skip, disable, retry-loop or quarantine a test to get green.
- Tests verify behavior through public APIs. Don't test private helpers or mock what you own.

## Traceability

Tests that verify a kibi scenario link to it ([kibi.md](kibi.md)). Coverage percentage is not a target; requirement coverage in `kibi report` is.
