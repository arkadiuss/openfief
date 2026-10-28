# Glossary

## Domain terms

The ubiquitous language: one term, one meaning, one code name. Use these names in code, tests, kibi requirements, commits and docs. Go names and "don't say" entries are proposals **(P)**.

### M0

| Term | Context | Go name (P) | Meaning | Don't say |
|---|---|---|---|---|
| Institution | ingest | `ingest.Institution` | A broker, bank or registry that holds assets: IBKR, XTB, the treasury bond registry. | broker (for the registry) |
| Account | ledger | `ledger.Account` | One account at an institution, such as one IBKR linked account. Belongs to one sleeve. | portfolio |
| Ledger account | ledger | `ledger.LedgerAccount` | A double-entry account inside an account: cash per currency, a holding per instrument, fees, withholding tax, FX. | "account" alone |
| Transaction | ledger | `ledger.Transaction` | One economic event from a source. Append-only and versioned; keyed by source plus the institution's own ID. | trade, activity, operation |
| Posting | ledger | `ledger.Posting` | One leg of a transaction. A transaction's postings sum to zero per currency. | entry, leg |
| Instrument | ledger | `ledger.Instrument` | Something that can be held, identified by ISIN plus a venue symbol. A bond series is an instrument. | security, ticker, asset |
| Holding | ledger | `ledger.Holding` | The quantity of one instrument in one account on a date, derived from postings. | position (allowed only inside connector adapters, as the broker's word) |
| Lot | ledger | `ledger.Lot` | A quantity acquired on a date at a cost; consumed FIFO. An opening lot is seeded from a broker's current positions instead of full history. | |
| Balance assertion | ledger | `ledger.BalanceAssertion` | A broker-reported balance on a date that the derived balance must match; a mismatch is an error. | snapshot |
| Sleeve | portfolio | `portfolio.Sleeve` | A part of the portfolio with one role. Defensive: not available to agents. Offensive: the part fiefs are leased from. | bucket |
| Import batch | ingest | `ingest.Batch` | One run of a connector's ingest path. | sync |
| Price | valuation | `valuation.Price` | An end-of-day price of an instrument from a market-data source. | quote |
| FX rate | valuation | `valuation.FXRate` | A rate to PLN; NBP table A in M0. | |
| Valuation | valuation | `valuation.Valuation` | The value of a holding on a date, with its basis and source. | |
| Valuation basis | valuation | `valuation.Basis` | `market`, `gross_accrued` or `net_if_redeemed`. | |
| Bond series | valuation | `valuation.Series` | A Polish treasury bond series code, such as EDO0536, with its terms. | bond (ambiguous) |
| Connector | ingest | `ingest.Connector` | A plugin that brings data in through an ingest path. | integration |
| View | views | `views.View` | A plugin that presents the canonical record elsewhere and never writes back. | sync, export |
| Read API | views | `views` + `connectapi` | The public, read-only `/v1` API. | |
| Plugin | plugins | `plugins.Plugin` | The custom resource holding what a plugin requests: image, manifest, config. | extension, addon |
| Grant | plugins | `plugins.Grant` | The custom resource holding what the owner approved for a plugin. | permission set |
| Manifest | plugins | `plugins.Manifest` | A plugin's declared capabilities, egress hosts, secrets, data tiers and config schema. | |
| Egress allow-list | plugins | `plugins.EgressAllowList` | The hostnames a plugin may reach through the egress proxy. | whitelist |
| Owner | identity | `identity.Owner` | The person who runs an installation: one individual investor. | user, customer |
| Script token | identity | `identity.ScriptToken` | A scoped, expiring, read-only personal API token. | API key |

### Later milestones

Don't create code for these before their milestone.

| Term | Meaning |
|---|---|
| Fief (lease) | A revocable capital lease from the offensive sleeve to one agent for one thesis, with hard limits. In feudal Europe, a fief was land a lord granted to a vassal in return for service, revocable on breach. |
| Thesis | An investment idea with a horizon and invalidation conditions that one or more agents work on. |
| Agent | One running instance of an agent plugin, spawned for one thesis. |
| Intent | An agent's structured request to trade (instrument, side, size, limit price, thesis, invalidation conditions), sent to the gateway instead of a broker order. |
| Order | The broker order the gateway creates from an approved intent. |
| Fill | One execution against an order. |
| Policy decision | The result of evaluating an intent: allow, deny or needs approval, with the rules matched. |
| Audit event | One append-only record of a state change. |
| Model portfolio | A target allocation the real portfolio follows and adjusts towards. |
| Target | One weight in the model portfolio. |
| Invalidation condition | A condition stated up front that would prove a thesis wrong, such as a price level or a fundamental change. |

### Rules

- A new domain term goes here first, in the same change as the code that introduces it.
- Broker vocabulary (position, trade, conid, symbol quirks) stays inside connector adapters. Translate at the boundary; that's the anti-corruption layer.
- One name per concept across contexts. If two contexts need different meanings, name them differently; never overload a term.

## Other terms and abbreviations

| Term | Meaning |
|---|---|
| A2A | Agent2Agent protocol: an open protocol for agents to talk to each other and to orchestrators. Introduced by Google in 2025, now hosted by the Linux Foundation. MCP connects an agent to tools; A2A connects agents to agents. |
| ADV | Average daily volume: mean shares (or value) traded per day. A minimum ADV filter keeps agents out of instruments a small order could move. |
| AGPL-3.0 | GNU Affero General Public License v3: copyleft that also triggers on network use, so running modified code as a service obliges you to publish the source. |
| API | Application programming interface. |
| Belka tax | Poland's flat 19% tax on capital income, named after Marek Belka, the finance minister under whom it was introduced. |
| Capability token | A token that *is* the permission: holding it lets you do exactly what it encodes. Macaroons and Biscuit let a holder attenuate a token, narrowing it but never widening it. The idea goes back to capability-based security (Dennis and Van Horn, 1966). |
| CAS | Compare-and-set: write only if the stored value or revision is still what you read. Originally a CPU instruction (IBM System/370). |
| CFD | Contract for difference: a leveraged derivative that pays the price change of an instrument without ownership. XTB's main product. |
| Circuit breaker (drawdown) | An automatic stop that suspends trading once losses from a peak (the drawdown) pass a limit. |
| Client Portal Web API | IBKR's REST-style web API. |
| CloudNativePG (CNPG) | A Kubernetes operator that runs PostgreSQL, including backups to object storage. |
| ClusterIP / port-forward | A ClusterIP Service is reachable only inside the Kubernetes cluster; `kubectl port-forward` tunnels it to your own machine. |
| CNI | Container Network Interface: the cluster's network plugin. NetworkPolicy works only if the CNI enforces it. |
| COI / EDO | Polish retail treasury bond series: COI 4-year and EDO 10-year, both inflation-indexed after the first year. |
| Connect RPC | A protocol by Buf: the same protobuf service can be called as gRPC or as plain HTTP with JSON. |
| cosign / Sigstore | Sigstore is an open project for signing software; cosign is its tool for signing and verifying container images. Keyless signing ties a signature to a CI identity instead of a long-lived key. |
| CPI | Consumer price index; the inflation measure published by GUS. |
| CRD / operator | A CustomResourceDefinition adds a new resource type to Kubernetes; an operator is a controller that turns those resources into running workloads. |
| CSP | Content Security Policy: a browser header that restricts where a page may load scripts and data from. |
| DCO | Developer Certificate of Origin: contributors certify they may submit their change with a `Signed-off-by` line, instead of signing a contributor license agreement. |
| Digest pinning | Referring to a container image by the hash of its content (`image@sha256:...`) rather than a mutable tag such as `:latest`. |
| Double-entry bookkeeping | Every transaction is recorded as postings that sum to zero across accounts, so money never appears or vanishes; documented by Luca Pacioli in 1494. |
| Egress proxy | A proxy that all outbound traffic from a set of pods must pass through, so destinations can be allowed by hostname rather than IP. |
| End of day (EOD) | Prices or rates as of each trading day's close. |
| ESMA | European Securities and Markets Authority, the EU securities regulator. |
| etcd | The key-value store that holds a Kubernetes cluster's state, including Secrets. |
| FIFO | First in, first out: the oldest lots are treated as sold first when computing gains. |
| Flex Query | IBKR's configurable reporting export (trades, positions, cash), retrievable through a web service with a token. |
| Frontier model | The most capable models from the large labs, typically cloud-hosted. |
| GUS | Główny Urząd Statystyczny, Statistics Poland. |
| HITL | Human in the loop: a person must approve before the system acts. |
| htmx | A small JavaScript library that lets server-rendered HTML pages update parts of themselves without a front-end framework. |
| HTTP CONNECT | The HTTP method a client uses to ask a proxy to open a tunnel to a named host and port; HTTPS runs inside the tunnel, so the proxy sees the hostname but not the content. |
| IB Gateway / TWS | Trader Workstation (TWS) is IBKR's desktop trading platform; IB Gateway is its headless variant. Both expose the TWS API. |
| IBKR | Interactive Brokers. |
| IKE / IKZE | Polish tax-advantaged retirement accounts with annual contribution limits. IKE: gains tax-exempt if withdrawal conditions are met. IKZE: contributions deductible, withdrawals taxed at a reduced flat rate. |
| Ingress | The Kubernetes resource that exposes an HTTP service outside the cluster. |
| kind / k3d / k3s | kind and k3d run Kubernetes clusters in containers for testing; k3s is a lightweight Kubernetes distribution. |
| KNF | Komisja Nadzoru Finansowego, the Polish Financial Supervision Authority. |
| KV | Key-value store. |
| Limit order | An order that executes only at a stated price or better. |
| Linked account | An IBKR account linked under one user, with its own cash, positions and buying power. |
| LLM | Large language model. |
| MCP | Model Context Protocol: an open protocol, introduced by Anthropic in 2024, through which a model or agent calls tools and reads resources. |
| MiFID II | The EU Markets in Financial Instruments Directive (second version), which defines investment advice and portfolio management. |
| Model router | The component that decides which model serves each role and which data tiers it may receive. |
| NATS / JetStream | NATS is an open-source messaging system; JetStream is its persistence layer (streams, KV, object store). |
| NBP table A | The National Bank of Poland's daily table of average exchange rates, used to convert foreign income for tax. |
| NetworkPolicy | A Kubernetes resource restricting which pods a pod may talk to and be reached from. |
| Notional | The money value of an order or position (quantity times price). |
| OMS | Order management system: the component that validates, routes and tracks orders. |
| OPA / Rego | Open Policy Agent, a CNCF policy engine; Rego is its Datalog-inspired policy language. |
| Paper trading | Trading against a simulated account with fake money. |
| PIT-8C | The annual statement of capital income that Polish brokers issue to clients and the tax office. |
| PIT-38 | The Polish annual tax return for capital gains (securities, derivatives). |
| PIT/ZG | The PIT attachment that declares foreign income, one per country. |
| PKO BP | PKO Bank Polski, the bank that distributes Polish retail treasury bonds. |
| Pod model | How multi-manager hedge funds run: capital is allocated to independent teams (pods) with risk limits, increased on a good record and cut on drawdown. |
| Projected ServiceAccount token | A short-lived token that Kubernetes mounts into a pod, scoped to one audience and verifiable against the cluster's issuer. |
| Prompt injection | Text that a model reads as data but follows as instructions, such as a news article telling a trading agent to buy. |
| PWA | Progressive web app: a website installable on a phone like an app. |
| RBAC | Role-based access control: in Kubernetes, Roles and ClusterRoles define what an identity may do. |
| Reader / decider split | One model reads untrusted text and emits typed notes; a separate decider sees only those notes. |
| Reconcile loop | The Kubernetes controller pattern: repeatedly compare desired state with actual state and act to remove the difference. |
| RP ID | Relying-party ID: the domain a WebAuthn passkey is bound to. |
| Smokescreen | Stripe's open-source egress proxy, written in Go, which checks each client's HTTP CONNECT destinations against a per-client allow-list. |
| TIF | Time in force: how long an order stays working (for example day, or good-till-cancelled). |
| WebAuthn / passkey | The web standard for public-key authentication; a passkey is a WebAuthn credential stored on a device. |
| xAPI | XTB's former public trading API, shut down in 2025. |
| XTB | X-Trade Brokers, a Polish broker. |
