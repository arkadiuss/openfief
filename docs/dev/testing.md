# Testing

Rules without a marker are the author's; **(P)** marks a Claude Code proposal awaiting the author's approval. Every test must be stable and runnable locally.

## Levels

| Level | What | How |
|---|---|---|
| Unit | Domain logic | Pure functions, small table-driven tests |
| Integration | Adapters against real protocols | HTTP with [hypert](https://github.com/areknoster/hypert) record and replay; Postgres in a disposable container **(P)** |
| End-to-end | The full flow on Kubernetes | A local k3s (native or k3d) |

## Unit tests

- Domain code is pure functions ([code-organization.md](code-organization.md)), so unit tests need no mocks.
- One behavior per table, with small named cases in the [glossary's](../glossary.md) terms, run in parallel.
- Compare with go-cmp; no assertion framework **(P)**.
- Ledger invariants (postings sum to zero per currency; FIFO consumes the oldest lot first) also get property-based tests **(P)**.
- Parsers, such as the XTB export, use golden files built from synthetic exports.

## Integration tests

- Committed HTTP tests always replay; recording is never left on in committed code.
- Recordings are committed with their test, and replay checks each request against its recording.
- Recording is a deliberate, manual act: follow the `hypert-recording` skill.
- hypert sanitizes requests only, so responses must be scrubbed too, or recorded only against paper, demo or public endpoints.
- A recording never contains an account number, holding, amount or token from a real account (GR-12).
- Each Postgres test gets a fresh database with all migrations applied **(P)**.

## End-to-end tests

- They run the real system on a local k3s: the bootstrap install, the operator, the plugins and a real Ghostfolio.
- External institutions are fixture servers in the cluster; end-to-end tests never call a real broker.
- `task e2e` refuses any kubeconfig context that isn't local **(P)**.
- Each end-to-end test verifies a kibi scenario.
- End-to-end tests log in with a real passkey ceremony: they drive the browser at `http://localhost` through a port-forward and use the browser's virtual authenticator. No flag weakens the product's origin check for tests.

## Stability

- No sleeps: wait on conditions, with a deadline.
- No wall clock: inject the date.
- No unseeded randomness, no network in unit or replay tests, no dependence on test order.
- A flaky test is a bug: fix the cause. Never skip, disable, retry-loop or quarantine a test to get green.
- Test behavior through public APIs; don't test private helpers or mock what you own.

## Traceability

Tests that verify a kibi scenario are linked to it ([kibi.md](kibi.md)). Coverage percentage is not a target; requirement coverage is.
