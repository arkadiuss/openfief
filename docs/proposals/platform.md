# Proposals: platform (M0 details)

Not decided. When the author decides an item, it moves into [DESIGN.md](../DESIGN.md) as a decision and leaves this file.

- **[P-3] Any Postgres.** Besides CloudNativePG (D-14), accept any PostgreSQL through a connection string, so D-12's "any conformant cluster" holds for managed databases too.
- **[P-6] Bootstrap install.** One small Helm chart or plain manifest installs the CRDs, the operator and the egress proxy. Everything after that is a `Plugin` and its `PluginGrant`.
- **[P-7] Built-in UI stack.** Server-rendered Go templates with htmx, a small chart library served locally (no CDN), a strict CSP, and no JavaScript build chain.

## Plugin contracts for M0

Interface outlines, not code. Every plugin also follows D-11, D-13, D-16, D-17, D-27 and D-32.

- **[P-8] Connector contract.**

  | Aspect | Outline |
  |---|---|
  | Purpose | Bring broker, registry or data-source data into OpenFief. The connector serves the service and the core calls it (D-49). Execution connectors (M1) also carry orders. |
  | Manifest | Institution, version, capabilities (read holdings, read transactions, statement import with file formats, market data; later execute live and execute paper), authentication method. |
  | Read operations | List accounts. Get a holdings snapshot as of a date. Get transactions and tax lots since a date. Import a statement file into normalized transactions. Get quotes and reference data (data connectors). |
  | Normalization | Instruments by ISIN where available, plus venue symbol. Every amount with its currency. Every record tagged with its source and the institution's own ID, so re-imports are idempotent (D-18). |
  | Failure | Errors mark the data as stale, never as zero. |

- **[P-9] View contract.**

  | Aspect | Outline |
  |---|---|
  | Purpose | Present the canonical record somewhere else; never write back (D-2). |
  | Manifest | Target system, egress hosts it needs (such as the user's Ghostfolio URL), secrets (such as the Ghostfolio security token), sync interval. |
  | Reads | Only the public read API (D-6). |
  | Writes | Only to its own target system. |
  | Ghostfolio | Reconcile loop per D-4 and D-64. Bonds become manually priced assets. |


## Model portfolio and currency

- **[P-48] Each target's drawdown in its own currency.** A target's own max drawdown (D-54) is computed in the target's own currency, as drawdowns are usually quoted, so it can use the target's whole history. The model's drawdown is in the reporting currency (D-62), like its backtest.
- **[P-49] Cross rates through PLN.** When the reporting currency (D-62) isn't PLN, M0 converts with NBP table A cross rates: an amount in X becomes X→PLN divided by reporting-currency→PLN, on the same table. One FX source then serves any reporting currency until another FX plugin exists, and tax keeps using NBP directly (D-20). NBP's history starts on 2002-01-02 (R-17). Resolves T-12 if accepted.
- **[T-12]** Reporting currency vs exchange rates. D-62 makes the reporting currency a setting, but D-20 converts amounts with NBP table A, which quotes every currency only against PLN. A reporting currency other than PLN needs cross rates through PLN or another FX plugin. Proposed: P-49. Deferred past M0 (D-201).

