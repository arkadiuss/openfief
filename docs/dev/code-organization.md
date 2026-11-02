# Code organization

Organized by domain, following DDD (domain-driven design). Names come from [the glossary](../glossary.md). Rules without a marker are the author's; **(P)** marks a Claude Code proposal awaiting the author's approval.

## Bounded contexts

| Context | Owns | Milestone |
|---|---|---|
| `ledger` | Accounts, ledger accounts, transactions, postings, instruments, holdings, lots, balance assertions | M0 |
| `valuation` | Prices, index levels, FX rates, bond series and statutory valuation, valuations | M0 |
| `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger | M0 |
| `views` | The read API and view plugins (built-in UI, Ghostfolio) | M0 |
| `plugins` | The operator: Plugin, Grant, manifests, egress allow-lists, image verification | M0 |
| `identity` | Owner, passkeys, script tokens, plugin identity | M0 |
| `attention` | Attention items: their kinds, subjects, states and history, the checks that raise and clear them, and the outbox of their changes (D-130, D-131); schema `attention` | M0 |
| `portfolio` | Sleeves, the model portfolio and its targets, backtest and max drawdown | M0 (model portfolio optional) |
| `jurisdiction/pl` | Polish rules: bond terms, tax wrappers; later PIT-38 | M0 (bonds) |
| `fiefs`, `execution`, `orchestration`, `models`, `audit` | Leases, intents/orders/policy, agents and wakes, model routing, audit log | M1+ |

**Context map.** `ingest` writes to `ledger` and `valuation` through their app APIs, and reads `plugins` for what a connector's signed manifest declares. `valuation` reads `ledger`. `portfolio` owns sleeves and which account is in which sleeve, and reads `ledger` and `valuation`; `ledger` knows nothing of sleeves. `views` reads `ledger`, `valuation`, `portfolio`, `attention`, `ingest` (accounts' connectors and statement uploads), `plugins` (the installed plugins) and `identity` (the authorizer, D-163). `attention` reads `ledger`, `valuation`, `ingest` and the jurisdiction pack through ports in `attention/app` for its checks; a context that raises items itself, such as `valuation` for an instrument to classify (D-132), declares its own port, which `cmd` wires to `attention`'s items. `jurisdiction/pl` is pure rules, used by `valuation`, `ingest`, `views` and `attention`. `plugins` and `identity` depend on no other context. Dependencies never point the other way. A tiny shared kernel `internal/kernel` holds money, currency, quantity and date types. A test in `task check` builds the import graph with `go list` and fails on any dependency between contexts that this map doesn't name.

## Layout (P)

```
cmd/<binary>/            wiring only: config, adapters, start
internal/<context>/      domain: types and pure functions (package <context>)
internal/<context>/app/  use cases; ports (interfaces) declared here, by the consumer
internal/<context>/<adapter>/   postgres, connectapi, nbp, ibkrflex, ...
internal/kernel/         shared kernel: money, currency, quantity, date. Pure.
proto/openfief/<context>/v1/    contracts (Apache-2.0, own LICENSE file, D-5)
sdk/go/, sdk/ts/         plugin SDKs (Apache-2.0, own LICENSE file, separate modules)
plugins/<name>/          first-party plugins, built on the SDK
deploy/                  CRDs, bootstrap chart
test/e2e/                end-to-end suites (local k3s)
tools/<name>/            development tools, never shipped (docs generator)
site/                    the docs site: Hugo config, theme, site-only pages (docs/dev/docs.md)
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
- Enforced by lint (`.golangci.yml`) and by the context map's import test.

**Plugins.** A plugin, first-party ones included, imports only `sdk/go`, never `internal/`. What a connector needs to build a delivery lives in the SDK, so our own plugins prove the SDK is enough. The import test enforces this too.

**Interfaces.** Declare an interface where it's consumed (in `app`), keep it small, and return concrete types.

**Anti-corruption.** Broker and vendor formats are parsed and translated in their adapter or plugin. Domain code never sees an IBKR, XTB or Ghostfolio shape.

**Banned packages.** No `util`, `common`, `helpers`, `models` or `types`. Find the context the code belongs to.

**Generated code** lives in `gen/` directories and is never edited by hand.

**Secrets and data.**
- Never log amounts, holdings, account numbers or tokens. Log internal IDs only (GR-12, UC-0).
- Fixtures use synthetic data only.

**Requirements.** Every behavioral function is owned by a kibi requirement ([kibi.md](kibi.md)).

**New context.** Add it to the table above and its terms to the glossary, and get the author's OK. Contexts are design decisions, not refactors.
