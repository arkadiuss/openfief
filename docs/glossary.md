# Glossary

## Domain terms

The ubiquitous language: one term, one meaning, one code name. Use these names in code, tests, kibi requirements, commits and docs. Go names and "don't say" entries are proposals **(P)**.

### M0

| Term | Context | Go name (P) | Meaning | Don't say |
|---|---|---|---|---|
| Institution | ingest | `ingest.Institution` | A broker, bank or registry that holds assets: IBKR, XTB, the treasury bond registry. | broker (for the registry) |
| Account | ledger | `ledger.Account` | One account at an institution, such as one IBKR linked account. Belongs to one sleeve. | portfolio |
| Ledger account | ledger | `ledger.Book` | A double-entry account inside an account: cash per currency, a holding per instrument, fees, withholding tax, FX, and equity: the counterpart of what enters or leaves the account from outside (deposits, withdrawals, opening lots). The Go name avoids the `ledger.LedgerAccount` stutter that lint rejects. | "account" alone |
| Transaction | ledger | `ledger.Transaction` | One economic event from a source. Append-only and versioned; keyed by source plus the institution's own ID. | trade, activity, operation |
| Posting | ledger | `ledger.Posting` | One leg of a transaction. A transaction's postings sum to zero per currency. | entry, leg |
| Instrument | ledger | `ledger.Instrument` | Something that can be held, identified by ISIN plus a venue symbol. A bond series is an instrument. | security, ticker, asset |
| Holding | ledger | `ledger.Holding` | The quantity of one instrument in one account on a date, derived from postings. | position (allowed only inside connector adapters, as the broker's word) |
| Lot | ledger | `ledger.Lot` | A quantity acquired on a date at a cost; consumed FIFO. An opening lot is seeded from a broker's current positions instead of full history. | |
| Balance assertion | ledger | `ledger.BalanceAssertion` | A broker-reported balance on a date that the derived balance must match; a mismatch is an error. | snapshot |
| Model portfolio | portfolio | `portfolio.Model` | An optional weighted basket of targets for the whole portfolio, which the real portfolio is compared with. Entered by the owner in M0; GR-4's maintainer agent adjusts it from M2. | strategy, benchmark |
| Target | portfolio | `portfolio.Target` | One instrument or index in the model portfolio, with its weight. The weights sum to 100%. | allocation |
| Backtest | portfolio | `portfolio.Backtest` | The model portfolio's value series computed from end-of-day history, rebalanced to the target weights once a year. | simulation |
| Max drawdown | portfolio | `portfolio.MaxDrawdown` | The worst fall from a peak to a later trough in a value series, with both dates. | max loss |
| Sleeve | portfolio | `portfolio.Sleeve` | A part of the portfolio with one role. Defensive: not available to agents. Offensive: the part fiefs are leased from. | bucket |
| Jurisdiction pack | jurisdiction | `jurisdiction.Pack` | The pluggable rules for one country (GR-13): bond terms, tax rules, and the default reporting currency. Poland first. | locale, country plugin |
| Import batch | ingest | `ingest.Batch` | One run of a connector's ingest path. | sync |
| Price | valuation | `valuation.Price` | An end-of-day price of an instrument from a market-data source. | quote |
| Index | valuation | `valuation.Index` | A market index, such as WIG20 or MSCI World, with an end-of-day level series from the market-data connector. It can be a target, but it can't be held. | benchmark |
| FX rate | valuation | `valuation.FXRate` | The value in PLN of one unit of a currency on a date, from a source; NBP table A in M0. The rate for a date is the latest one dated on or before it. | |
| Reporting currency | valuation | `valuation.ReportingCurrency` | The currency totals, charts and the model comparison are shown in. A setting whose default comes from the jurisdiction pack: PLN for the Polish pack. | base currency, home currency |
| Valuation | valuation | `valuation.Valuation` | The value of a holding on a date, with its basis and source. | |
| Valuation basis | valuation | `valuation.Basis` | `market`, `gross_accrued` or `net_if_redeemed`. | |
| Bond series | jurisdiction/pl | `pl.Series` | A Polish retail treasury bond series code, such as EDO0536 (EDO maturing in May 2036). Valuation uses it through jurisdiction/pl's rules. | bond (ambiguous) |
| Bond type | jurisdiction/pl | `pl.BondType` | The product a bond series belongs to, named by the code's first three letters: OTS, ROR, DOR, TOS, COI, EDO, ROS, ROD. Fixes the term, interest period, indexation and payout. | series (for the type) |
| Series terms | jurisdiction/pl | `pl.Terms` | What one series' issue letter fixes: first-period rate, margin and early-redemption fee. Data supplied per series, never hard-coded. | |
| Interest period | jurisdiction/pl | `pl.PeriodStart` | One period of a bond's interest, counted from its own purchase date: a month (ROR, DOR), three months (OTS) or a year. | coupon period (for capitalized types) |
| Period index | jurisdiction/pl | `pl.Indices` | The published reading that sets an interest period's rate after the first: CPI year on year for inflation-indexed types, the NBP reference rate for ROR and DOR. | |
| Margin | jurisdiction/pl | `pl.Terms.Margin` | The fixed rate a series adds to its period index. | spread |
| Early-redemption fee | jurisdiction/pl | `pl.Terms.EarlyRedemptionFee` | The per-bond fee for redeeming before maturity, capped at the bond's accrued interest except for coupon types after their first interest period, and not charged on a withdrawal from IKE or IKZE. OTS has none: redeemed early, it earns no interest. | penalty |
| Exchange price | jurisdiction/pl | `pl.Purchase.Price` | The price per bond of a series bought by exchanging (zamiana) maturing bonds, set in its issue letter: usually 99.90 PLN, 100 PLN for OTS. The discount below the nominal is taxed with the interest. | rollover price |
| Tax wrapper | jurisdiction/pl | `pl.TaxWrapper` | The tax regime of the account a lot is held in: taxable, IKE or IKZE. | account type |
| Lot value | jurisdiction/pl | `pl.LotValue` | A bond lot's statutory value on a date: gross accrued value, and net-if-redeemed value after the early-redemption fee and Belka tax on interest minus the fee plus any purchase discount. | |
| Connector | ingest | `ingest.Connector` | A plugin that brings data in through an ingest path. | integration |
| View | views | `views.View` | A plugin that presents the canonical record elsewhere and never writes back. | sync, export |
| Read API | views | `views` + `connectapi` | The public, read-only `/v1` API. | |
| Plugin | plugins | `plugins.Plugin` | The custom resource for one installed, named instance of a plugin image: the image and its manifest (the vendor's) and the instance's name and config (the owner's). Its status says whether the image is verified. | extension, addon |
| Grant | plugins | `plugins.Grant` | The custom resource holding what the owner approved for a plugin. | permission set |
| Catalog | plugins | `plugins.Catalog` | The file in the OpenFief repository that lists first-party and verified plugins, with pinned image digests and signer identities. | marketplace, registry |
| Verified plugin | plugins | `plugins.Verified` | A plugin whose catalog entry pins a version the author reviewed. Anything installed by image reference alone is unverified. | official, trusted |
| Trusted signer | plugins | `plugins.Signer` | An identity whose image signatures the operator accepts: OpenFief's CI, a catalog publisher, or one the owner added. | |
| Manifest | plugins | `plugins.Manifest` | A plugin's declared capabilities, egress hosts, secrets, data tiers and config schema. | |
| Egress allow-list | plugins | `plugins.EgressAllowList` | The hostnames a plugin may reach through the egress proxy. | whitelist |
| Capability | plugins | `plugins.Capability` | What a plugin may act as: `connector` or `view` in M0. Declared in its manifest and approved in its grant. | kind, role |
| Egress host | plugins | `plugins.EgressHost` | One hostname and port in a manifest, grant or egress allow-list. Never an IP address or a wildcard. | domain, URL |
| Data tier | plugins | `plugins.DataTier` | A named class of data a plugin asks to see. Only a name in M0; what each tier holds is undecided (P-33). | |
| Image reference | plugins | `plugins.ImageRef` | A plugin's container image, pinned by digest (`repository@sha256:...`), never by tag. | image tag |
| Effective permissions | plugins | `plugins.EffectivePermissions` | What OpenFief enforces for a plugin: only what its Plugin requests and its Grant approves. | |
| Workload | plugins | `plugins.Workload` | What the operator creates for one plugin, compiled from its effective permissions: identity, token, Secret mounts, environment, NetworkPolicy and egress allow-list. | deployment (the Kubernetes object is one part) |
| Plugin identity | identity | `plugins.ServiceAccountName` | The ServiceAccount `plugin-<name>` in the `openfief` namespace, proven by its projected token with audience `openfief-core`. | |
| Owner | identity | `identity.Owner` | The person who runs an installation: one individual investor. | user, customer |
| Script token | identity | `identity.ScriptToken` | A scoped, expiring, read-only personal API token. | API key |
| Currency | kernel | `kernel.Currency` | An ISO 4217 alphabetic code, such as PLN or USD. | |
| Money | kernel | `kernel.Money` | An exact decimal amount in one currency. It keeps its original currency; arithmetic across currencies is an error, and conversion is an explicit FX step in valuation. | |
| Quantity | kernel | `kernel.Quantity` | An exact, possibly fractional number of units of an instrument. | |
| Date | kernel | `kernel.Date` | A calendar date with no time of day or time zone, written YYYY-MM-DD. | timestamp |

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
| COI / EDO / ROS / ROD | Polish retail treasury bond types indexed to inflation after the first year: COI 4-year with yearly coupons, EDO 10-year, and the family bonds ROS 6-year and ROD 12-year, all three capitalized yearly. |
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
| EODHD | EOD Historical Data, a market-data vendor; the first market-data connector uses it. |
| ESMA | European Securities and Markets Authority, the EU securities regulator. |
| etcd | The key-value store that holds a Kubernetes cluster's state, including Secrets. |
| FIFO | First in, first out: the oldest lots are treated as sold first when computing gains. |
| Flex Query | IBKR's configurable reporting export (trades, positions, cash), retrievable through a web service with a token. |
| Frontier model | The most capable models from the large labs, typically cloud-hosted. |
| GPW | Giełda Papierów Wartościowych w Warszawie, the Warsaw Stock Exchange. |
| GUS | Główny Urząd Statystyczny, Statistics Poland. |
| HITL | Human in the loop: a person must approve before the system acts. |
| htmx | A small JavaScript library that lets server-rendered HTML pages update parts of themselves without a front-end framework. |
| HTTP CONNECT | The HTTP method a client uses to ask a proxy to open a tunnel to a named host and port; HTTPS runs inside the tunnel, so the proxy sees the hostname but not the content. |
| IB Gateway / TWS | Trader Workstation (TWS) is IBKR's desktop trading platform; IB Gateway is its headless variant. Both expose the TWS API. |
| IBKR | Interactive Brokers. |
| Issue letter (list emisyjny) | The Minister of Finance's document that sets one bond series' terms and how its interest and value are computed. |
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
| NBP reference rate | The National Bank of Poland's main policy interest rate, which sets ROR and DOR interest. |
| NBP table A | The National Bank of Poland's daily table of average exchange rates, used to convert foreign income for tax. |
| NetworkPolicy | A Kubernetes resource restricting which pods a pod may talk to and be reached from. |
| Notional | The money value of an order or position (quantity times price). |
| OMS | Order management system: the component that validates, routes and tracks orders. |
| OPA / Rego | Open Policy Agent, a CNCF policy engine; Rego is its Datalog-inspired policy language. |
| Paper trading | Trading against a simulated account with fake money. |
| OTS / ROR / DOR / TOS | Polish retail treasury bond types: OTS 3-month fixed, ROR 1-year and DOR 2-year with monthly coupons at the NBP reference rate plus a margin, TOS 3-year fixed and capitalized yearly. |
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
| TokenRequest / TokenReview | Kubernetes APIs: TokenRequest issues a short-lived ServiceAccount token for a chosen audience; TokenReview checks whether a token is valid. |
| TWR / MWR | Time-weighted return chains the returns between cash flows, so deposits and withdrawals don't count as performance; it is the standard for comparing with a benchmark. Money-weighted return (an IRR, such as XIRR) weights by the money invested, so it reflects the investor's timing. |
| Vanguard | A Go library by Buf that serves one Connect service as REST, gRPC and Connect, translating between them (transcoding). |
| VWAP | Volume-weighted average price. |
| WebAuthn / passkey | The web standard for public-key authentication; a passkey is a WebAuthn credential stored on a device. |
| xAPI | XTB's former public trading API, shut down in 2025. |
| XTB | X-Trade Brokers, a Polish broker. |
