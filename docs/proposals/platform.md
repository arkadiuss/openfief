# Proposals: platform (M0 details)

Not decided. When the author decides an item, it moves into [DESIGN.md](../DESIGN.md) as a decision and leaves this file.

- **[P-3] Any Postgres.** Besides CloudNativePG (D-14), accept any PostgreSQL through a connection string, so D-12's "any conformant cluster" holds for managed databases too.
- **[P-6] Bootstrap install.** One small Helm chart or plain manifest installs the CRDs, the operator and the egress proxy. Everything after that is a `Plugin` and its `PluginGrant`.
- **[P-7] Built-in UI stack.** Server-rendered Go templates with htmx, a small chart library served locally (no CDN), a strict CSP, and no JavaScript build chain.
- **[P-45] Embedding Smokescreen (R-14).** Build the proxy with `BuildProxy` inside OpenFief's own HTTP server, not `StartWithConfig`, which exits the process on errors and takes over signals. Never leave the egress ACL unset: Smokescreen then allows everything. Swap compiled allow-lists atomically. Plugins send their token as the password in the proxy URL, which Go's and Node's HTTP clients turn into a `Proxy-Authorization` header. The role function checks it with Kubernetes' TokenReview API (audience `openfief-egress`) and caches the result. Projected tokens rotate, so the SDK rebuilds the proxy URL from the token file on each request instead of a fixed `HTTPS_PROXY`, and it points at the proxy's ClusterIP, since plugins have no DNS. Pin a commit after v0.1.0 until a release carries its later hardening fixes.

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
  | Ghostfolio | Reconcile loop per D-4. Bonds become manually priced assets. Sync mechanics: P-46. |

- **[P-46] Ghostfolio sync mechanics (R-15), within D-4.**
  - The plugin creates its Ghostfolio accounts with UUIDs it generates and stores; that's how it knows which accounts it owns.
  - Each sync imports missing activities first, then deletes stale ones. Ghostfolio deletes an asset and its price history along with its last activity, so deleting first would wipe a bond's prices. A bulk delete is never sent without an account filter, since an unfiltered one deletes every activity the user has.
  - A bond is a `MANUAL` asset whose symbol is a UUID derived from its series code, with the code as its name. Every sync sends its full price history.
  - A manually edited activity keeps its comment, so catching edits means comparing fields, not only IDs.
  - Imports aren't atomic, so every step is safe to repeat, and each import is first checked with `dryRun=true`.
  - Ghostfolio blanks comments when its "restricted view" setting is on, so the plugin refuses to sync then; otherwise every activity would look stale.
  - These are Ghostfolio's internal endpoints, which get renamed within `/v1`. The plugin supports a pinned range of Ghostfolio versions, checked by contract tests on recorded fixtures.

## Model portfolio

- **[P-47] Time-weighted return for the portfolio's line.** The portfolio's line on the value-over-time chart is its time-weighted return (TWR): daily returns from valuations, chained across external cash flows such as deposits and withdrawals. Transfers between the owner's own accounts are not flows. That makes it comparable with the model, which has no flows. A money-weighted return (an IRR, such as XIRR) can come later as a personal figure. Answers OQ-28 with P-48.
- **[P-48] Currency of the comparison.** The model and the portfolio are both measured in PLN at NBP table A rates (D-20), so currency moves count as they do for the owner. NBP's API serves table A from 2002-01-02 (R-17), so the model's backtest in PLN starts no earlier. Each target's own max drawdown (D-54) is in the target's own currency, as drawdowns are usually quoted, so it can use the target's whole history.

## M0 acceptance

- **[P-10] M0 acceptance criteria.** Each becomes a kibi scenario on the requirement it checks when work on that area starts, and leaves this list.
  1. **Consolidation.** The author's real IBKR, XTB (stocks and ETFs) and treasury bond holdings appear in one view. Every broker-reported position and cash balance is stored as a balance assertion and matches the derived value to 0.01 of the account currency. Open XTB CFDs appear only as flagged, unsupported items.
  2. **Double-entry.** Every transaction's postings sum to zero per currency, and a check over the whole log passes.
  3. **Idempotent, versioned imports.** Re-importing the same file adds nothing. A changed row creates a new version, and the old one remains, marked superseded.
  4. **Opening lots.** IBKR and XTB accounts are seeded from current lot-level positions, each lot with open date and cost, and PLN cost at the NBP table A rate from the day before the open date.
  5. **Bonds.** For every held series, the gross accrued value matches an open-source reference valuer on the same date to within 0.01 PLN, and net-if-redeemed matches a hand calculation for at least one series of each type held. Totals use the configured basis, net by default.
  6. **Prices and FX.** Every PLN figure uses NBP table A. End-of-day prices come from the market-data plugin; where IBKR Flex reports a price for the same instrument and date, a difference beyond a configured tolerance is flagged.
  7. **Read API.** `/v1` serves accounts, transactions, tax lots, holdings and valuations as its OpenAPI spec describes, rejects unauthenticated calls, and has no write endpoint. The Connect clients generated from the same definitions return the same data.
  8. **Built-in UI.** Passkey login works on `localhost`, and a passkey created for a second hostname through a one-time link works there. The allocation and value-over-time charts are drawn using only the read API.
  9. **Ghostfolio.** After a sync, Ghostfolio's OpenFief-owned accounts hold exactly the projected activities. A second sync changes nothing. A manually edited activity in an owned account is reverted, and accounts OpenFief did not create are untouched.
  10. **Plugin isolation.** Each plugin runs in its own pod. The XTB parser has no egress at all. A test plugin's request through the proxy for a host outside its grant is refused, and attempts to bypass the proxy (a direct connection, a DNS lookup) fail.
  11. **Operator.** A `Plugin` without a matching `PluginGrant` is not deployed. An upgrade requesting a new egress host stays on the old version until granted. An image referenced by tag, or without a valid cosign signature from a trusted signer, is rejected, including when applied with plain `kubectl`. The `Plugin`'s status shows verified only for an image the catalog pins.
  12. **Operator scope.** The operator's ServiceAccount cannot create or change anything outside the `openfief` namespace.
  13. **Exposure.** The installation creates no Ingress, no LoadBalancer or NodePort Service and no tunnel; the UI is reachable through `kubectl port-forward`.
  14. **Secrets.** Credentials exist only as Kubernetes Secrets, mounted only into plugins granted them.
  15. **Portability.** The whole acceptance suite passes on both kind and k3d.
  16. **Egress.** Outbound traffic goes only to configured destinations (IBKR's Flex service, NBP, GUS, the market-data vendor, the user's Ghostfolio), verified from the egress proxy's logs.
  17. **Core-to-plugin.** A call to a connector from any pod other than the core fails, and a call from the core without a valid token for that instance is refused.
  18. **Uninstall.** Uninstalling a connector instance removes its pod, grant, secrets and allow-list; its imported transactions remain, marked as coming from a removed instance.
  19. **Model portfolio.** While off, the UI and the read API show nothing of it. When on, the backtest of a two-target basket over a synthetic price series matches a hand calculation across one yearly rebalance, and each max drawdown matches the series' known peak and trough dates.
