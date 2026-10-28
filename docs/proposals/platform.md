# Proposals: platform (M0 details)

Not decided. When the author decides an item, it moves into [DESIGN.md](../DESIGN.md) as a decision and leaves this file.

- **[P-1] One contract language.** Define D-6's public read API in protobuf too, with HTTP annotations, and serve it as REST with a generated OpenAPI spec through transcoding (for example Buf's Vanguard). Plugins and the public API then share one source of truth.
- **[P-2] Signature verification in the operator.** D-12 (any cluster) means no admission controller such as Kyverno or Sigstore's policy-controller can be assumed, so the operator verifies image signatures itself before creating any Deployment: `sigstore-go`, keyless signatures from the project's GitHub Actions identity, and a bundled trust root so verification also works offline.
- **[P-3] Any Postgres.** Besides CloudNativePG (D-14), accept any PostgreSQL through a connection string, so D-12's "any conformant cluster" holds for managed databases too.
- **[P-4] Canonical origin set at install.** WebAuthn binds each passkey to a domain (the RP ID). The installer requires the canonical hostname up front; `localhost` (D-28's port-forward) is used only for bootstrap, and the docs warn that passkeys do not carry over between hostnames. Answers T-11 and OQ-24 if accepted.
- **[P-5] The operator never touches money.** The operator manages plugin workloads only. Fiefs, intents, the ledger and approvals live in Postgres behind the core, never in custom resources or etcd: etcd is not a transactional ledger, and GR-6's guarantee must not depend on who holds Kubernetes API access.
- **[P-6] Bootstrap install.** One small Helm chart or plain manifest installs the CRDs, the operator and the egress proxy. Everything after that is a `Plugin` and its `PluginGrant`.
- **[P-7] Built-in UI stack.** Server-rendered Go templates with htmx, a small chart library served locally (no CDN), a strict CSP, and no JavaScript build chain.
- **[P-44] Core-to-plugin authentication.** A plugin accepts calls only from the core (D-49): its NetworkPolicy admits ingress from the core pod alone, and the core presents a projected ServiceAccount token with the plugin instance as audience, which the SDK verifies.

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
  | Ghostfolio | Reconcile loop per D-4. Bonds become manually priced assets. API calls to verify first: OQ-23. |

## M0 acceptance

- **[P-10] M0 acceptance criteria.** Each becomes a kibi scenario on the requirement it checks when work on that area starts, and leaves this list.
  1. **Consolidation.** The author's real IBKR, XTB (stocks and ETFs) and treasury bond holdings appear in one view. Every broker-reported position and cash balance is stored as a balance assertion and matches the derived value to 0.01 of the account currency. Open XTB CFDs appear only as flagged, unsupported items.
  2. **Double-entry.** Every transaction's postings sum to zero per currency, and a check over the whole log passes.
  3. **Idempotent, versioned imports.** Re-importing the same file adds nothing. A changed row creates a new version, and the old one remains, marked superseded.
  4. **Opening lots.** IBKR and XTB accounts are seeded from current lot-level positions, each lot with open date and cost, and PLN cost at the NBP table A rate from the day before the open date.
  5. **Bonds.** For every held series, the gross accrued value matches an open-source reference valuer on the same date to within 0.01 PLN, and net-if-redeemed matches a hand calculation for at least one series of each type held. Totals use the configured basis, net by default.
  6. **Prices and FX.** Every PLN figure uses NBP table A. End-of-day prices come from the market-data plugin; where IBKR Flex reports a price for the same instrument and date, a difference beyond a configured tolerance is flagged.
  7. **Read API.** `/v1` serves accounts, transactions, tax lots, holdings and valuations as its OpenAPI spec describes, rejects unauthenticated calls, and has no write endpoint.
  8. **Built-in UI.** Passkey login works. The allocation and value-over-time charts are drawn using only the read API.
  9. **Ghostfolio.** After a sync, Ghostfolio's OpenFief-owned accounts hold exactly the projected activities. A second sync changes nothing. A manually edited activity in an owned account is reverted, and accounts OpenFief did not create are untouched.
  10. **Plugin isolation.** Each plugin runs in its own pod. The XTB parser has no egress at all. A test plugin's request through the proxy for a host outside its grant is refused, and attempts to bypass the proxy (a direct connection, a DNS lookup) fail.
  11. **Operator.** A `Plugin` without a matching `PluginGrant` is not deployed. An upgrade requesting a new egress host stays on the old version until granted. An image referenced by tag, or without a valid cosign signature, is rejected.
  12. **Operator scope.** The operator's ServiceAccount cannot create or change anything outside the `openfief` namespace.
  13. **Exposure.** The installation creates no Ingress, no LoadBalancer or NodePort Service and no tunnel; the UI is reachable through `kubectl port-forward`.
  14. **Secrets.** Credentials exist only as Kubernetes Secrets, mounted only into plugins granted them.
  15. **Portability.** The whole acceptance suite passes on both kind and k3d.
  16. **Egress.** Outbound traffic goes only to configured destinations (IBKR's Flex service, NBP, GUS, the market-data vendor, the user's Ghostfolio), verified from the egress proxy's logs.
