# Code organization

Organized by domain, following DDD (domain-driven design). Names come from [the glossary](../glossary.md). Rules without a marker are the author's; **(P)** marks a Claude Code proposal awaiting the author's approval.

## Bounded contexts

| Context | Owns | Milestone |
|---|---|---|
| `ledger` | Accounts, ledger accounts, transactions, postings, instruments, holdings, lots, balance assertions | M0 |
| `valuation` | Prices, index levels, FX rates, bond series and statutory valuation, valuations | M0 |
| `ingest` | Institutions, connectors' ingest paths, import batches, mapping broker data to the ledger, and the account settings connectors read, such as the bond payout setting (D-135) | M0 |
| `views` | The read API and view plugins (built-in UI, Ghostfolio) | M0 |
| `plugins` | The operator: Plugin, Grant, manifests, egress allow-lists, image verification | M0 |
| `identity` | Owner, passkeys, script tokens, plugin identity | M0 |
| `attention` | Attention items: their kinds, subjects, states and history, the checks that raise and clear them, and the outbox of their changes (D-130, D-131); schema `attention` | M0 |
| `portfolio` | Sleeves, the model portfolio and its targets, backtest and max drawdown, its drift from the real portfolio and the maintainer's runs, notes and proposals (D-204) | M0 (model portfolio optional), the maintainer M3 |
| `jurisdiction/pl` | Polish rules: bond terms, tax wrappers; later PIT-38 | M0 (bonds) |
| `modelrouter` | The one switch for agents and models, each role's primary model and fallbacks, the data tiers and the models allowed per tier, the model-provider plugins' contract, and the audit entry of every model call (D-205, D-206, D-211, D-212); schema `modelrouter` | M1 |
| `fiefs`, `execution`, `orchestration`, `audit` | Leases, intents/orders/policy, agents and wakes, audit log | M1+ |

**Context map.** `ingest` writes to `ledger` and `valuation` through their app APIs, and reads `plugins` for what a connector's signed manifest declares. `valuation` reads `ledger`. `portfolio` owns sleeves and which account is in which sleeve, and reads `ledger` and `valuation`; `ledger` knows nothing of sleeves. The maintainer in `portfolio/app` reaches the model router, the holdings the views read, the attention items, the audit log and the job queue only through its own ports, which `cmd/core` wires, so `portfolio` imports none of `modelrouter`, `views`, `attention` and `identity`; its River job (`portfolio/jobs`) rides on the revaluation client in `valuation/postgres`, which `cmd/core` configures. `views` reads `ledger`, `valuation`, `portfolio`, `attention`, `ingest` (accounts' connectors and statement uploads), `plugins` (the installed plugins) and `identity` (the authorizer, D-163). `attention` reads `ledger`, `valuation`, `ingest` and the jurisdiction pack through ports in `attention/app` for its checks; a context that raises items itself, such as `valuation` for an instrument to classify (D-132), declares its own port, which `cmd` wires to `attention`'s items. `jurisdiction/pl` is pure rules, used by `valuation`, `ingest`, `views` and `attention`. `plugins`, `identity` and `modelrouter` depend on no other context: `modelrouter` reaches model-provider plugins through its `Providers` port and writes its audit entries through its `Auditor` port, which `cmd/core` wires to the plugins' Services and to `identity`'s log. Dependencies never point the other way. A tiny shared kernel `internal/kernel` holds money, currency, quantity and date types. An architecture test enforces it ([Architecture tests](#architecture-tests)).

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
- A context's Kubernetes API types (`internal/plugins/operator/v1alpha1`, its CRDs) are a contract, like `proto/`: the context's adapters and `cmd` may import them.
- `app` imports domain packages.
- Adapters import `app` and domain packages.
- Only `cmd` imports adapters.
- A context never imports another context's adapters or reads its tables: each context owns its schema **(P)**.
- Enforced by lint (`.golangci.yml`) and by the architecture tests.

**Plugins.** A plugin, first-party ones included, imports only `sdk/go`, never `internal/`. What a connector needs to build a delivery lives in the SDK, so our own plugins prove the SDK is enough. An architecture test enforces it.

**Interfaces.** Declare an interface where it's consumed (in `app`), keep it small, and return concrete types.

**Anti-corruption.** Broker and vendor formats are parsed and translated in their adapter or plugin. Domain code never sees an IBKR, XTB or Ghostfolio shape.

**Support packages.** `internal/demo` holds the synthetic demo dataset and is no context: it imports domain packages only, and only `cmd` and the benches import it. `internal/kernel` is the shared kernel above. A *bench* is a development tool under `tools/` that drives the product's own code and the demo dataset to grade it, never shipped: the Ask box bench (`tools/askbench`, [Models](../concepts/models.md)). Like `cmd` it may import adapters and the datasets, and nothing imports it; it is the only package that imports goeval (D-218), and only the packages of goeval that need nothing but the standard library.

**Banned packages.** No `util`, `common`, `helpers`, `models` or `types`. Find the context the code belongs to.

**Generated code** lives in `gen/` directories and is never edited by hand.

**Secrets and data.**
- Never log amounts, holdings, account numbers or tokens. Log internal IDs only (GR-12, UC-0).
- Fixtures use synthetic data only.

**Requirements.** Every behavioral function is owned by a kibi requirement ([kibi.md](kibi.md)).

**New context.** Add it to the table above and its terms to the glossary, and get the author's OK. Contexts are design decisions, not refactors.


## Architecture tests

Lint checks what one file does; architecture tests check how the code fits together, from the whole repository's packages, types, SQL, contracts and tests. They live in [tools/archtest](../../tools/archtest/doc.go) and run in `task check` with the other Go tests (`go test ./...`), and in the pre-commit hook whenever a change can affect them. Each rule they enforce is stated above or in [testing.md](testing.md), and a failure names the rule and this section:

- **Context map.** No context imports another that the map doesn't name, and `kernel` imports no context. Every package under `internal/` belongs to a context in the table, or is the demo dataset (`internal/demo`, [demo.md](demo.md)), which is not a context: it reads only domain packages and only `cmd` and the benches import it **(P)**.
- **Layering.** A domain package imports only the kernel, its own context's domain and the domains the map allows; an `app` package imports domain and other contexts' `app` packages, never an adapter; no context imports another context's adapters; only `cmd` and the benches import adapters (and only `cmd` imports `cmd`). An adapter is everything below `internal/<context>/` other than `app`; its own subpackages, such as `cosign/cosigntest`, count as the adapter.
- **Plugins.** Nothing under `plugins/` or `sdk/` imports `internal/` or `cmd/`.
- **Schemas.** A context's SQL, in its migrations and in its packages' query strings, names only its own schema; `valuation` also owns `river`, River's job queue (D-36). SQL comments don't count.
- **Money.** No `float32` or `float64` in domain, `app` and kernel packages, and no `float` or `double` field in the contracts under `proto/`. Code that only draws, such as chart coordinates, is listed by package.
- **Pure domain.** No `go` statements and no package-level variables other than errors, compiled regular expressions (and `_`) in domain and kernel packages, beyond what depguard and forbidigo already ban.
- **Privacy (GR-12).** No logging call (`log/slog`, `log`, logr, logrus) takes a value of a money, quantity, decimal, price, FX rate, ISIN or script token type as an argument, or an error, whose text may quote one: an error is logged through a `LogSafe` function, such as `kernel.LogSafe`. Attribute constructors, `fmt.Sprint*`, `.String()` and `.Error()` are looked through. A string holding an account number or a token has no type to tell it apart, so review still owns those.
- **No network in transactions.** No PostgreSQL adapter, a package named `postgres` below `internal/`, which is the code that runs in a change's transaction, imports `net/http`, a Connect or gRPC client (SpiceDB's included) or the Kubernetes client. A change that needs a connector, a model provider, SpiceDB or the cluster asks before its transaction opens, in a use case, or defers the call until it has committed ([Identity](../concepts/identity.md#no-network-call-inside-a-change)); the clients `cmd/core` builds refuse a call made inside one at run time, which covers what an import can't see.
- **Banned packages.** No package, test-only ones included, named `util`, `common`, `helpers`, `models` or `types`.
- **Stable tests.** No `time.Sleep` and no `Skip` in tests, outside a listed few that wait on something outside the process or need a tool a contributor may lack, each with its reason.
- **Ask box bench.** Only a bench (`tools/askbench`) imports goeval (D-218), tests included, and only `goeval/api`, `goeval/heuristic` and `goeval/openai`, which need nothing but the standard library: the root package, `gemini`, `llmjudge` and `embedding` bring the Gemini and Google Cloud client libraries into every build of the module.
- **Catalog-independent tests.** The embedded plugin catalog (`internal/plugins/catalog`, D-43) grows with every release (D-88), so no test pins it: a test never compares the number of `catalog.Load()` entries with a literal, never picks an entry by a literal index, and never spells out a digest the catalog pins; a test that inspects what a function of its package returns, when that function reads the catalog, derives its expectation from `catalog.Load()` too, directly or through a helper. Checking only the error is fine **(P)**.

A rule that the code breaks today starts with those breaks listed in a checked-in baseline, [tools/archtest/baseline.txt](../../tools/archtest/baseline.txt), each with its reason; the few breaks a rule allows for good (the listed tests, drawing packages) are in [exceptions.txt](../../tools/archtest/exceptions.txt). A break is named by a stable key (an import edge, a package, file or function), never a line number. The test fails on any break neither list names and on any listed break that no longer happens, so the baseline only shrinks, both lists stay exact, and neither is ever edited to get green: fix the code, and delete the line in the same change.
