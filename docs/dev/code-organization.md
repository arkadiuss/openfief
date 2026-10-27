# Code organization

Organized by domain, following DDD (domain-driven design). Names come from [ubiquitous-language.md](ubiquitous-language.md). Rules without a marker are the author's; **(P)** marks a Claude Code proposal pending approval.

## Bounded contexts

| Context | Owns | Milestone |
|---|---|---|
| `ledger` | Accounts, ledger accounts, transactions, postings, instruments, holdings, lots, balance assertions | M0 |
| `valuation` | Prices, FX rates, bond series and statutory valuation, valuations | M0 |
| `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger | M0 |
| `views` | The read API and view plugins (built-in UI, Ghostfolio) | M0 |
| `plugins` | The operator: Plugin, Grant, manifests, egress allow-lists, image verification | M0 |
| `identity` | Owner, passkeys, script tokens, plugin identity | M0 |
| `portfolio` | Sleeves; later the model portfolio and targets | M0 (sleeves only) |
| `jurisdiction/pl` | Polish rules: bond terms, tax wrappers; later PIT-38 | M0 (bonds) |
| `fiefs`, `execution`, `orchestration`, `models`, `audit` | Leases, intents/orders/policy, agents and wakes, model routing, audit log | M1+ |

**Context map (P).** `ingest` writes to `ledger` and `valuation` through their app APIs. `valuation` reads `ledger`. `views` reads `ledger`, `valuation` and `portfolio`. `jurisdiction/pl` is pure rules used by `valuation`. `plugins` and `identity` are independent. Dependencies never point the other way. A tiny shared kernel `internal/kernel` holds money, currency, quantity and date types.

## Layout (P)

```
cmd/<binary>/            wiring only: config, adapters, start
internal/<context>/      domain: types and pure functions (package <context>)
internal/<context>/app/  use cases; ports (interfaces) declared here, by the consumer
internal/<context>/<adapter>/   postgres, connectapi, nbp, ibkrflex, ...
internal/kernel/         shared kernel: money, currency, quantity, date. Pure.
proto/openfief/<context>/v1/    contracts (Apache-2.0, D-5)
sdk/go/, sdk/ts/         plugin SDKs (Apache-2.0, separate modules)
plugins/<name>/          first-party plugins, built on the SDK
deploy/                  CRDs, bootstrap chart
test/e2e/                end-to-end suites (local k3s)
```

## Rules

**Domain packages (`internal/<context>`):**
- Pure functions over values. No I/O, network, database, logging, globals, goroutines or `context.Context`.
- No `time.Now()` or randomness: take the date or seed as a parameter.
- Business failures are returned errors with typed values, never panics.

**Money (P).** Never `float64` for amounts, prices, rates or quantities. Use the kernel's decimal-backed types, and every amount carries its currency.

**Layering:**
- `app` imports domain packages.
- Adapters import `app` and domain packages.
- Only `cmd` imports adapters.
- A context never imports another context's adapters or reads its tables: each context owns its schema **(P)**.
- Enforced by lint ([lint.md](lint.md)).

**Interfaces.** Declare an interface where it's consumed (in `app`), keep it small, and return concrete types.

**Anti-corruption.** Broker and vendor formats are parsed and translated in their adapter or plugin. Domain code never sees an IBKR, XTB or Ghostfolio shape.

**Banned packages.** No `util`, `common`, `helpers`, `models` or `types`. Find the context the code belongs to.

**Generated code** lives in `gen/` directories and is never edited by hand.

**Secrets and data.**
- Never log amounts, holdings, account numbers or tokens. Log internal IDs only (GR-12, UC-0).
- Fixtures use synthetic data only.

**Requirements.** Every behavioral function is owned by a requirement through a kibi `symbol` entity with an `implements` link, not a code comment ([kibi.md](kibi.md)). The pre-commit gate blocks unowned behavioral symbols.

**New context.** Add it to the table above and to the ubiquitous language, and get the author's OK. Contexts are design decisions, not refactors.
