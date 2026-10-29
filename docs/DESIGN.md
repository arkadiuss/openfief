# OpenFief design

The current design: what is decided, as of now. Git holds the history.

- Proposals for later milestones, one file per area: [proposals/](proposals/)
- External research the design draws on: [research.md](research.md)
- Domain terms and abbreviations: [glossary.md](glossary.md)
- Requirements: kibi (`.kb/`), which cites the GR and D IDs below. People read them through its generated report.

Authority, highest first: ground rules (GR), then decisions (D), then kibi requirements. Tensions (T) and open questions (OQ) are unresolved. IDs are stable and never reused.

## 1. Ground rules

Written by the author and never edited. A conflict with a ground rule is recorded as a tension; the rule is never reworded to fit.

- **[GR-1]** Fully open source and self-hosted. It can eventually also be hosted for a subscription.
- **[GR-2]** Target users are people like me: technical individual investors. It is a niche, and that is fine.
- **[GR-3]** It starts as a typical portfolio summarizer: plugins for XTB, Interactive Brokers, Polish retail treasury bonds (obligacje skarbowe), etc., giving typical API access to what you hold, plus a couple of charts.
- **[GR-4]** There is a model strategy (model portfolio) to follow and adjust, with an agentic model that summarizes and maintains it. The model is configurable; I would use only my locally hosted model for this.
- **[GR-5]** For a thesis (for example a single company) you can spawn agents. Mainly long-term, but short-term is allowed. Agents do fundamental analysis, watch news, etc.
- **[GR-6]** Each thesis agent has a budget leased from the offensive part of the portfolio, with strong guarantees that the agent can never get more of the portfolio than its lease.
- **[GR-7]** Agents wake up on a schedule plus on events, and have tools to make transactions.
- **[GR-8]** Frontier models for thesis agents can be cloud-hosted.
- **[GR-9]** An orchestration layer runs the thesis agents. That orchestration layer is exactly what this project is; agents are separate products.
- **[GR-10]** Agents are plugins, so any agent can be used. The default is an agent I build or pfranczyk's per-thesis agents (github.com/Looted). Someone who trusts Interactive Brokers' agents and does not host models locally can choose those instead.
- **[GR-11]** All models are configurable per role, so the user chooses which model sees what.
- **[GR-12]** The privacy and security model comes first, designed per use case.
- **[GR-13]** Regulatory logic is pluggable too: jurisdiction plugins, Poland first.
- **[GR-14]** The project name is OpenFief.

## 2. Scope

- **[D-1]** The first iteration is M0 only: tracker and Polish imports, with no AI models and no agents.
- **[D-31]** M0 has no time-box. Its scope is everything decided in this document.
- **[D-5]** The core is licensed AGPL-3.0, the plugin contracts and SDK Apache-2.0, and contributions carry a DCO sign-off.

| Milestone | Scope | State |
|---|---|---|
| M0 | Tracker and Polish imports, optional model portfolio | Decided: this document |
| M1 | Control plane on paper trading | Proposed: [proposals/control-plane.md](proposals/control-plane.md) |
| M2 | Agents | Proposed: [proposals/agents-and-models.md](proposals/agents-and-models.md) |
| M3 | First small live fief | Proposed: [proposals/control-plane.md](proposals/control-plane.md) |

## 3. M0 architecture

Arrows show data flow, not which side starts a call.

```mermaid
flowchart LR
  you(["You: browser via kubectl port-forward<br/>or your own Ingress"])
  cli(["You: openfief CLI<br/>plugin install"])

  subgraph ns["namespace openfief"]
    op["Operator<br/>Plugin + PluginGrant CRDs<br/>digest + cosign checks"]
    core["Core (Go)<br/>double-entry log, lots, valuations<br/>read API: REST /v1 + Connect"]
    pg[("PostgreSQL<br/>CloudNativePG")]
    egress["Egress proxy (Smokescreen)<br/>per-plugin hostname allow-list"]
    subgraph plugins["Plugins: one container each, own NetworkPolicy and Secrets"]
      ui["View: built-in UI<br/>passkey login, 2 charts"]
      gfp["View: Ghostfolio<br/>reconcile loop"]
      ibkr["Connector: IBKR Flex<br/>opening lots + daily pull"]
      xtb["Connector: XTB export parser<br/>no egress"]
      bonds["Connector: treasury bonds<br/>manual lots, CPI + NBP ref. rate"]
      md["Connector: market data<br/>EODHD, your API key"]
      nbp["Connector: NBP FX"]
    end
  end

  gf["Ghostfolio<br/>(your separate install)"]
  ext["External: IBKR Flex service,<br/>NBP, GUS, market-data vendor"]

  cli -- "Plugin + PluginGrant<br/>(your kubeconfig)" --> op
  op -. "creates and constrains" .-> plugins
  core --> pg
  you --> ui
  core -- "read API" --> ui
  core -- "read API" --> gfp
  gfp -- "import + reconcile" --> egress
  egress --> gf
  ibkr -- "transactions, lots, assertions" --> core
  xtb -- "parsed transactions" --> core
  bonds -- "lots + valuations" --> core
  md -- "EOD prices" --> core
  nbp -- "EOD FX" --> core
  ext --> egress
  egress -- "granted hosts only" --> ibkr
  egress --> bonds
  egress --> md
  egress --> nbp
```

### 3.1 System of record and views

Existing trackers can't hold what OpenFief has to get exactly right: FIFO tax lots priced in PLN, statutory bond values, and sleeve attribution that the later lease ledger reconciles against. So OpenFief owns the record, and trackers only display it.

- **[D-2]** OpenFief's own core is the system of record for transactions, tax lots and valuations. Trackers are plugins that show a one-way *view* of it.
- **[D-3]** M0 ships two view plugins: a very simple built-in UI as the default, and Ghostfolio auto-push.
- **[D-4]** The Ghostfolio plugin runs a reconcile loop. It owns only the Ghostfolio accounts it created, puts OpenFief's transaction ID in each activity's comment, deletes stale activities and imports missing ones. Manual edits inside its accounts are reverted on the next sync.
- **[D-64]** How the Ghostfolio plugin syncs, given how Ghostfolio behaves:
  - It creates its Ghostfolio accounts with UUIDs it generates and stores; that's how it knows which accounts it owns.
  - Each sync imports missing activities first, then deletes stale ones, because Ghostfolio deletes an asset and its price history along with its last activity. A bulk delete always carries an account filter; an unfiltered one deletes every activity the user has.
  - A bond is a `MANUAL` asset whose symbol is a UUID derived from its series code, with the code as its name. Every sync sends its full price history.
  - Edits are found by comparing fields, since a manually edited activity keeps its comment.
  - Imports aren't atomic, so every step is safe to repeat, and each import is first checked with `dryRun=true`.
  - It refuses to sync while Ghostfolio's "restricted view" setting is on: that setting blanks comments, so every activity would look stale.
  - Ghostfolio renames these internal endpoints within `/v1`, so the plugin supports a pinned range of Ghostfolio versions, checked by contract tests on recorded fixtures.
- **[D-6]** OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it. MCP access waits for M2, so that model access to holdings goes through the model router.
- **[D-50]** The read API (D-6) is defined in protobuf with HTTP annotations, like the plugin contracts (D-11). The core serves it both as Connect, for generated Go and TypeScript clients, and as REST under `/v1` through transcoding (Buf's Vanguard). The OpenAPI spec is generated from the same definitions.
- **[D-30]** The built-in UI draws two charts in M0: allocation (by asset class, currency, broker or sleeve) and value over time, which starts at go-live (D-23).

### 3.2 Getting data in

XTB shut down its public API in 2025, and neither the treasury bond registry nor the bank that sells the bonds offers one, so both arrive as files or manual entry. Interactive Brokers' Flex Queries are expected to report positions per lot, with open date and cost (OQ-22 verifies this).

- **[D-7]** Reading of GR-3: connectors get data from each institution by whatever works (a broker API where one exists, files or other means otherwise; not necessarily REST), and OpenFief exposes its own read API (D-6). Data enters only through each connector's ingest path (upload, CLI, watched folder), never through the public API.
- **[D-8]** The core project ships only connectors that comply with the institution's terms. Third parties may build others; OpenFief does not ship or endorse them.
- **[D-9]** M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).
- **[D-22]** Prices come from a market-data connector plugin where the user brings the API key (D-51); IBKR Flex prices serve as a cross-check.
- **[D-51]** The first market-data connector uses EODHD. EODHD covers the Warsaw Stock Exchange, LSE, Xetra and Euronext, finds instruments by ISIN, includes end-of-day index history, and its terms allow storing the data for private use. Its closes are aggregated rather than each exchange's official close, which is one more reason for D-22's cross-check.
- **[D-23]** Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live. Full history backfill arrives with the tax pack.
- **[D-24]** XTB in M0 covers real stocks and ETFs only. Closed CFD results are imported as cash profit or loss so cash reconciles; open CFD positions are flagged as unsupported.
- **[D-72]** The owner creates accounts in the built-in UI, in an owner session (D-66): institution, name, sleeve, and the connector instance that feeds the account. A broker connector reports the broker accounts it saw on its last sync, and the UI offers them to pick from, so the owner never types a broker account number. Transactions for a broker account that isn't mapped to an OpenFief account are not booked, and the UI shows the broker account as unmapped.

### 3.3 Ledger and valuation

Double-entry keeps cash, withholding tax and currency conversion consistent by construction, and it gives a later lease a natural home as a sub-account.

- **[D-18]** The source of truth is an append-only, versioned transaction log. Holdings and tax lots are derived from it. Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.
- **[D-19]** Double-entry postings across cash, position, fee, withholding-tax and FX accounts.
- **[D-20]** Every amount keeps its original currency. In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface with NBP as the default FX plugin. The Polish jurisdiction pack always uses NBP rates for tax, whatever the UI uses.
- **[D-62]** Totals, charts and the model comparison are shown in the reporting currency: a setting whose default comes from the jurisdiction pack, PLN for the Polish pack. Apart from the Polish pack's tax rules, no code assumes PLN.
- **[D-70]** Reading of GR-13 for M0: a jurisdiction pack is a set of pure functions inside the core, behind one interface, so another jurisdiction is another implementation of it. It is not a plugin container and not a WebAssembly (WASM) module.
- **[D-21]** Prices are end of day in M0.
- **[D-25]** Which bond value counts in totals is configurable; the default is net-if-redeemed-today (after the early-redemption fee and Belka tax, no tax inside IKE).
- **[D-33]** Besides D-19's accounts, each account has an equity ledger account per currency: the counterpart for deposits, withdrawals and opening lots.
- **[D-71]** Every sleeve has a sleeve kind, which gives it a default role: defensive (never available to agents) or offensive (fiefs are leased from it, GR-6). The built-in kinds are agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive. The owner can add kinds and rename any kind, but a kind's default role never changes, and a sleeve whose role differs from its kind's needs the owner's explicit choice. An install starts with three sleeves: bonds, core and satellite.
- **[D-34]** Selling part of a lot rounds the consumed cost to two decimal places, half-even, and the lot keeps the exact remainder, so consumed plus remaining cost always equals the original cost. In M0 fees go to the fee ledger account and are not added to lot cost.
- **[D-35]** An instrument needs an ISIN with a valid check digit plus a venue symbol. Within one date, acquisitions apply before disposals, because broker rows carry no reliable time of day.
- **[D-36]** Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time. Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River). Pending revaluations coalesce per account from the earliest affected date. The read API reports, per account, the date it is valued through and whether a revaluation is pending or running, and the UI shows it. Configurable: worker concurrency, the end-of-day schedule and the bond basis (D-25); changing the basis revalues everything. What triggers a revaluation is not configurable.
- **[D-37]** Where a treasury bond's issue letter and the issuer's published daily interest tables or worked examples disagree, OpenFief follows the tables and examples: the TOS yearly base is not rounded, and Belka tax is rounded half up to the grosz.

```mermaid
erDiagram
  SLEEVE ||--|{ ACCOUNT : "backed by"
  ACCOUNT ||--o{ TRANSACTION : records
  TRANSACTION ||--|{ POSTING : "balanced by"
  ACCOUNT ||--o{ BALANCE_ASSERTION : "checked by"
  ACCOUNT ||--o{ HOLDING : "derived into"
  INSTRUMENT ||--o{ HOLDING : "held as"
  HOLDING ||--|{ LOT : "made of"
```

The terms are defined in [glossary.md](glossary.md).

### 3.4 Model portfolio

GR-4's model portfolio starts in M0 as something the owner enters by hand. From M2, its maintainer agent adjusts the same basket.

- **[D-52]** The model portfolio is optional and off by default; while it is off, it appears nowhere in the UI or the read API. It is a weighted basket for the whole portfolio: each target is an instrument (such as an index ETF or a stock) or an index, with a weight, and the weights sum to 100%. The owner enters it in the built-in UI, which hands it to the core like manual bond lots (D-49). When it is on, the read API serves it with its backtest and drawdowns.
- **[D-53]** The model's performance is a backtest: its value is computed from end-of-day prices as far back as every target has data, rebalanced to the target weights once a year. The value-over-time chart (D-30) shows it as a second line from go-live, with both lines starting at 100. The portfolio's line is measured as D-61 says.
- **[D-54]** M0 shows the historical max drawdown of the model and of each target: the worst fall from a peak to a later trough over its whole history, with both dates. M0 has no stress tests and no statistical risk estimates.
- **[D-61]** Next to the model, the portfolio's line is its time-weighted return: daily returns from valuations, chained across deposits and withdrawals, so cash flows don't count as performance. Transfers between the owner's own accounts are not cash flows. Both lines are in the reporting currency (D-62). The model's backtest also needs exchange rates into that currency, so it starts no earlier than they do; NBP's rates begin on 2002-01-02.

### 3.5 Runtime and plugins

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. Standard NetworkPolicy matches IP addresses, not hostnames, so outbound traffic goes through a proxy that checks hostnames.

- **[D-10]** OpenFief is Kubernetes-oriented: it is built for, and deployed on, Kubernetes.
- **[D-11]** The core is written in Go. Each plugin runs in its own container and speaks protobuf contracts over Connect RPC. Each plugin ships a manifest that OpenFief compiles into NetworkPolicy, Secret mounts, RBAC and egress allow-list entries (D-32). SDKs come in TypeScript and Go.
- **[D-12]** Any conformant Kubernetes cluster; no reference platform. Tested on kind and k3d, with documented requirements: a CNI that enforces NetworkPolicy, and a storage class.
- **[D-13]** An operator manages plugins through a `Plugin` custom resource, from M0.
- **[D-55]** The operator manages plugin workloads only and never touches money. Fiefs, intents, approvals and the ledger live in PostgreSQL behind the core, never in custom resources or etcd: etcd is not a transactional ledger, and GR-6's guarantee must not depend on who has Kubernetes API access. The only custom resources are `Plugin` and `PluginGrant`.
- **[D-14]** PostgreSQL via CloudNativePG, and no NATS in M0.
- **[D-15]** The operator is namespace-scoped: a Role in the `openfief` namespace only. The CRDs are installed once by a cluster admin. Namespace-per-agent is revisited at M2.
- **[D-16]** Plugin permissions are split into the `Plugin` resource (what the vendor requests) and a `PluginGrant` (what the user approved). The operator enforces only what is in both. An upgrade that requests more waits for approval.
- **[D-17]** Plugin images must be pinned by digest, and their signatures are verified with cosign.
- **[D-32]** Plugin egress goes through an egress proxy built on Stripe's Smokescreen, embedded as a Go library. The operator compiles each `PluginGrant`'s hostnames into that plugin's allow-list. Plugins identify themselves with their projected ServiceAccount token, through a custom role function. NetworkPolicy lets plugins reach only the core and the proxy, with no DNS, and the proxy refuses private and internal address ranges. Plugins reach external hosts through `HTTPS_PROXY`.
- **[D-63]** The egress proxy embeds Smokescreen through `BuildProxy` inside OpenFief's own HTTP server; `StartWithConfig` would exit the process on errors and take over signals. Smokescreen allows every destination when no allow-list is set, so the proxy never runs without one and refuses all destinations for a plugin that has none. Compiled allow-lists are swapped atomically. A plugin sends its projected token as the password in the proxy URL, which Go's and Node's HTTP clients send as a `Proxy-Authorization` header; the role function checks it with Kubernetes' TokenReview API (audience `openfief-egress`) and caches the result. Projected tokens rotate, so the SDK builds the proxy URL (D-32's `HTTPS_PROXY`) from the token file on each request, using the proxy's ClusterIP because plugins have no DNS. Smokescreen is pinned to a commit after v0.1.0 until a release carries its later hardening fixes.

### 3.6 Plugin distribution and installation

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

- **[D-42]** A plugin's manifest is embedded in its image, as a label on the image config. One signature covers both the code and the permissions it asks for, and the manifest can be read from the registry without running the image.
- **[D-43]** One catalog, a file in this repository released with the product, lists the first-party and verified plugins. Any other plugin is installed directly by image reference and shown as unverified. There are no subscribable third-party catalogs.
- **[D-44]** A verified plugin is one whose version the author reviewed: its catalog entry pins the publisher's signed image digest and the publisher's signer identity. Each new version needs a new catalog entry.
- **[D-45]** The operator trusts three kinds of signer: OpenFief's CI, the publishers pinned in the catalog, and signers the owner adds. Owners sign their own plugins with their own CI identity or a local key, and those plugins show as unverified.
- **[D-46]** Plugins are installed with the `openfief` CLI. It reads the catalog or the image, verifies the signature, shows the requested permissions, and on the owner's confirmation applies the `Plugin` and `PluginGrant` with the owner's own kubeconfig, or prints them for GitOps. The core has no right to create or change grants; in M0 the UI only shows installed plugins.
- **[D-47]** Plugin upgrades are always manual, even when the permissions are unchanged.
- **[D-48]** Each install is a separately named instance with its own config, secrets, grant and egress allow-list. The same image can be installed more than once, such as one instance per IBKR account.
- **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.
- **[D-56]** The operator verifies every plugin image's signature itself before deploying it, so a `Plugin` applied through GitOps or plain `kubectl apply` is checked too. It uses a bundled trust root, so verification works offline. It records in the `Plugin`'s status whether the image is verified (D-44) or only signed by a trusted signer, and the UI shows that status. The CLI's check (D-46) is a preview of the same check.
- **[D-57]** An image without a valid signature from a trusted signer is never deployed, and there is no override. A developer testing a plugin locally signs it with their own key, added as a trusted signer (D-45).
- **[D-58]** Uninstalling an instance deletes its workload, grant, secrets and egress allow-list. The data it brought in stays in the ledger (D-18), marked as coming from a removed instance; reinstalling under the same name continues from it.

### 3.7 Security: tracking and imports

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files.

- **Data seen:** every holding, transaction, statement file and account identifier, plus read-only broker credentials such as a Flex Query token.
- **Models:** none.
- **Threats:** theft of the data from the host or its backups; leaked read-only credentials, which still expose the full account history; malicious statement files; wrong valuations leading to wrong decisions.
- **Controls:** per-plugin isolation (D-11, D-16, D-17, D-32), where the XTB parser has no egress at all, and:
  - **[D-26]** People log in with passkeys (WebAuthn), built in. No external identity provider.
  - **[D-65]** OpenFief has a single user, the owner: no other users and no tenants. Authenticators only prove that a caller is the owner, and they are separate from the session that follows. In M0 the authenticators are passkeys (D-26) and, on a demo, D-68's demo authenticator.
  - **[D-66]** After authentication the core issues an opaque, revocable session token, not a JSON Web Token (JWT). It stores only the token's hash, and the session expires after a period of inactivity and after an absolute lifetime. Each session has a scope: `read` or `owner`.
  - **[D-69]** The built-in UI is a backend-for-frontend: the browser talks only to the UI view plugin, and the plugin calls the core's authentication service (`openfief.identity.v1`) for login and sessions. That service is offered to view plugins only, never on the public read-only API (D-6).
  - **[D-27]** Plugins authenticate to the core with projected ServiceAccount tokens (audience `openfief-core`). Scripts get separate, scoped, expiring, read-only personal tokens.
  - **[D-67]** Script tokens (D-27) are created only from an owner session in the UI, and no token is installed in advance as a Secret. `core token issue`, run in the core's pod, stays as a break-glass for an administrator.
  - **[D-68]** A demo authenticator issues `read` sessions without credentials, only when the database carries the marker of the synthetic demo dataset and the core runs with `--demo-sessions`. On a public demo anyone can browse read-only, and the owner logs in with a passkey for an `owner` session.
  - **[D-59]** Only the core may call a plugin. A plugin's NetworkPolicy admits ingress from the core alone, and the core presents a short-lived ServiceAccount token whose audience is that plugin instance, requested through Kubernetes' TokenRequest API as instances come and go. The SDK verifies it. Plugins can't reach the Kubernetes API, so the operator publishes the cluster's token-signing public keys in a ConfigMap the SDK reads.
  - **[D-28]** OpenFief ships no Ingress and no tunnel. The core and UI are ClusterIP services; the default way in is `kubectl port-forward`; anything else is the user's own Ingress, and the docs explain what it exposes.
  - **[D-60]** OpenFief installs with no hostname. A passkey works only on the hostname it was created on, so the CLI prints a one-time link with which the owner creates the first passkey on `localhost`, through `kubectl port-forward`. When the owner adds their own Ingress, a logged-in session issues a one-time link for creating a passkey on the new hostname. Passkeys are kept per hostname.
  - **[D-29]** Credentials are Kubernetes Secrets, fed by whatever the user already runs (External Secrets Operator, Vault, SOPS and so on). The docs require etcd encryption at rest, encrypted disks and encrypted backups. No home-grown cryptography in M0.

## 4. Open for M0

Questions and tensions for later milestones live in their proposal files.

- **[T-12]** Reporting currency vs exchange rates. D-62 makes the reporting currency a setting, but D-20 converts amounts with NBP table A, which quotes every currency only against PLN. A reporting currency other than PLN needs cross rates through PLN or another FX plugin. Proposed: P-49.
- **[OQ-8]** Domain and trademark check for "OpenFief" (TMview, classes 9/36/42).
- **[OQ-22]** Verify the assumptions behind D-9, D-23 and D-24 against real accounts: IBKR Flex open positions report lots with open date and cost basis; the author's IBKR lot-matching method is FIFO; how far back one Flex request reaches; the xStation export's columns, and whether it separates CFDs from stocks; whether obligacjeskarbowe.pl offers any CSV export.
