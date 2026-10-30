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
| Opening cash | ledger | `ledger.NewOpeningCash` | The cash an account holds at go-live, seeded as a deposit, or a withdrawal for a negative balance, against the equity ledger account, keyed `opening-cash:<currency>:<date>`. | opening balance |
| Seeding | ledger | `app.Ledger.Seed` | Bringing an account into the ledger at go-live from its institution's current state: its opening lots and opening cash (D-23). An account whose current transactions include either is seeded, and seeding it again is refused unless nothing would change. | bootstrap, initial import |
| Balance assertion | ledger | `ledger.BalanceAssertion` | A broker-reported balance on a date that the derived balance must match; a mismatch is an error. | snapshot |
| Assertion check | ledger | `app.AssertionCheck` | A balance assertion with the result of its latest check: matched, or flagged with the derived balance when they differ. A flagged check is an error for the owner to resolve; the import that found it is kept. | reconciliation |
| Model portfolio | portfolio | `portfolio.Model` | An optional weighted basket of targets for the whole portfolio, which the real portfolio is compared with. Entered by the owner in M0; GR-4's maintainer agent adjusts it from M2. | strategy, benchmark |
| Target | portfolio | `portfolio.Target` | One instrument or index in the model portfolio, with its weight. The weights sum to 100%. | allocation |
| Backtest | portfolio | `portfolio.Backtest` | The model portfolio's value series computed from end-of-day history, rebalanced to the target weights once a year. | simulation |
| Max drawdown | portfolio | `portfolio.MaxDrawdown` | The worst fall from a peak to a later trough in a value series, with both dates. | max loss |
| Sleeve | portfolio | `portfolio.Sleeve` | A part of the portfolio with one sleeve kind and one role, which defaults to the kind's. Defensive: not available to agents. Offensive: the part fiefs are leased from. A role other than the kind's, either way, needs the owner's explicit choice, the role override (D-71). | bucket |
| Sleeve kind | portfolio | `portfolio.Kind` | A named category of sleeve with a default role, such as core, bonds, retirement or agent-pool. Built-in kinds come with OpenFief; the owner can add and rename kinds, but a kind's default role never changes. | sleeve type |
| Jurisdiction pack | jurisdiction | `jurisdiction.Pack` | The pluggable rules for one country (GR-13), pure functions behind one interface (D-70): its tax wrappers, the three values of every holding in each, bond terms, and the default reporting currency and number format. Poland first (`pl.Pack`). | locale, country plugin |
| Tax wrapper | jurisdiction | `jurisdiction.Wrapper`; `pl.TaxWrapper` in the Polish pack | An account's tax regime, one of its jurisdiction pack's, which the owner sets when creating the account (D-76): for Poland a standard taxable account (`taxable`), IKE, IKZE or OKI. It sets the taxes and fees in the account's net and liquidation values. | account type |
| Gross value | jurisdiction | `jurisdiction.Values.Gross` | A holding's value before any tax or fee: its market value, or the accrued value of a bond or deposit (D-77). What broker statements show. | |
| Net value | jurisdiction | `jurisdiction.Values.Net` | What the owner keeps at the planned exit (maturity, or a withdrawal from the wrapper once its conditions are met), after the taxes due then and without early-exit fees (D-77). | after-tax value |
| Liquidation value | jurisdiction | `jurisdiction.Values.Liquidation` | The cash in hand if everything were sold or redeemed and withdrawn today, after every tax and fee due today: early-redemption fees, early-withdrawal taxes and sale commissions (D-77). | net-if-redeemed (the older term) |
| Fee rule | jurisdiction | `jurisdiction.FeeRule` | How an account's broker charges a sale, set by the owner per account (D-78): a percentage of the value with a minimum, plus a currency-conversion percentage for an instrument in a foreign currency. Zero until set. | commission (for the rule) |
| Marginal rate | jurisdiction | `jurisdiction.Settings.MarginalRate` | The owner's marginal personal income tax (PIT) rate, which taxes a withdrawal from IKZE before 65 (D-78). A setting the owner enters once; until then an IKZE's liquidation value is missing. | tax bracket |
| Number format | jurisdiction | `jurisdiction.NumberFormat` | How the UI writes amounts: the separator between thousands and the decimal separator, a setting next to the reporting currency with the pack's default, `5 000,00 PLN` for Poland (D-113). It changes only separators, never digits or the currency. | locale |
| Rules version | jurisdiction | `jurisdiction.Pack.Version`, `pl.RulesVersion` | The version of a jurisdiction pack's rules. The core records it, and when a release changes it, every account is revalued. | |
| Import batch | ingest | `ingest.Batch` | One run of a connector's ingest path. | sync |
| Statement | ingest | `app.StatementFetcher` | What an institution reports about an account, as a file or a pulled document, such as an IBKR Flex Query or an xStation export. A connector maps it. | report, export (alone) |
| Ingest mode | ingest | `ingest.Mode` | What an import batch takes from a statement: `opening` seeds an account at go-live with opening lots and its opening cash (D-23); `transactions` imports transactions from go-live on. Both import balance assertions. | |
| Mapping | ingest | `ingest.Mapping` | What a connector makes of one statement: transactions, balance assertions, ingest problems and the institution's own prices as cross-check prices. | |
| Ingest problem | ingest | `ingest.Problem` | A statement row the mapping left out of the ledger, with its reason: `unsupported` (can't be booked yet, such as an option trade) or `skipped` (left out on purpose, such as a summary row). Not an error. | warning |
| Import summary | ingest | `ingest.ImportSummary` | What the ledger did with a batch's mapping: new transactions, new versions, unchanged ones, and failed balance assertions. | |
| Connector contract | ingest | `openfief.connector.v1.ConnectorService` | The protobuf service over Connect RPC that a connector plugin serves and the core calls (D-11, D-49): Sync pulls what the institution reported since a watermark, ImportFile maps a statement file, ImportBondLots maps bond lots the owner entered. Each call is one import batch, with a batch ID the core generates and reuses on a retry. Not the public read API. | ingest API |
| Delivery | ingest | `app.Delivery` | What a connector returns for one import batch: its mapping, plus any prices, FX rates, index readings, index levels, bond series terms and cross-check prices, and for a sync the next watermark. The core scopes its transaction keys to the connector instance. | push, upload (for plugins) |
| Watermark | ingest | `app.Delivery.Watermark` | Where a connector's next sync continues from. The connector returns it, the core stores it after importing the delivery and hands it back on the next sync; only the connector understands it. | cursor, checkpoint |
| Connector binding | ingest | `app.ConnectorBindings` | Permission for one connector instance to write to one account. The owner adds it in the built-in UI with the account, or with `core connector bind`; the core refuses, whole, a delivery that names an account its instance is not bound to. | account mapping |
| Broker account | ingest | `ingest.BrokerAccount` | One of the owner's accounts as the institution names it, such as an IBKR Flex account ID, reported by a broker connector (D-72). The owner maps it to an OpenFief account in the built-in UI instead of typing it; nothing of an unmapped one is booked. | external account, account number |
| Bond lot | ingest | `ingest.BondLot` | A lot of treasury bonds the owner enters by hand (D-9): its series code, quantity and purchase date, plus the series terms valuing it needs. The core hands it to the connector bound to its account (D-49), which books it as a deposit and a buy on the purchase date. | bond entry, manual lot |
| Sync policy | ingest | `ingest.SyncPolicy` | When the core syncs each connector instance (D-49): every interval after a success, with jitter, and after a failure again with a doubling backoff. | cron, schedule (alone) |
| Index reading | ingest | `ingest.IndexReading` | One published value of an index that sets treasury bond interest, as a yearly fraction: CPI year on year for a month (GUS), or the NBP reference rate for the month whose interest periods it sets, dated on that month's first day. Period indices are picked from them. | rate (alone) |
| Price | valuation | `valuation.Price` | An end-of-day price of an instrument from a market-data source. | quote |
| Cross-check price | valuation | `valuation.CrossCheckPrice`; `ingest.CrossCheckPrice` as a connector maps it | An institution's own end-of-day price of an instrument, such as the mark price of a position in an IBKR Flex statement, kept only to cross-check the market-data connector's prices (D-22). No holding is valued from it. `cross_check_prices` on the connector contract. | reference price, second price |
| Source ranking | valuation | `valuation.Ranking`, `valuation.RankSources` | The owner's priority order of the sources of one kind of market data (prices, FX rates or index readings), a source being the connector instance that delivered the values. For each key, such as an instrument and date, the highest-ranked source with a value wins; the others are kept as cross-checks (D-94). A source the owner hasn't placed follows in its default place: the first-party connector of its kind first, then by name. | source priority |
| Price check | valuation | `valuation.PriceCheck` | A price compared with a cross-check price of the same instrument, date and currency: the cross-check price's relative difference from the price, and matched, or flagged when the difference is beyond the price check threshold. A flagged check is for the owner to look at; it changes no price or valuation. | price reconciliation |
| Index | valuation | `valuation.Index` | A market index, such as WIG20 or MSCI World, with an end-of-day level series from the market-data connector. It can be a target, but it can't be held. | benchmark |
| Index level | valuation | `valuation.IndexLevel` | One end-of-day closing level of an index, named as the market-data connector's config names the index; `openfief.connector.v1.IndexLevel` on the contract. Not an index reading. | index price, quote |
| FX rate | valuation | `valuation.FXRate` | The value in PLN of one unit of a currency on a date, from a source; NBP table A in M0. The rate for a date is the latest one dated on or before it. | |
| Reporting currency | valuation | `valuation.ReportingCurrency` | The currency totals, charts and the model comparison are shown in. A setting whose default comes from the jurisdiction pack: PLN for the Polish pack. | base currency, home currency |
| Valuation | valuation | `valuation.Valuation` | The value of a holding on a date, with its basis and source. | |
| Valuation basis | valuation | `valuation.Basis` | How a holding's gross value was found: `market` (quantity times a price) or `gross_accrued` (a bond's nominal plus accrued interest). | |
| Holding valuation | valuation | `valuation.HoldingValuation` | The stored valuation of one holding on one day, keyed by account, instrument, date and valuation basis: its gross value in the instrument's own currency and in PLN, its net and liquidation values in PLN, with the source and dates of its inputs, or the missing input that stopped it; a net or liquidation value it can't have names its own reason. | |
| Revaluation | valuation | `valuation.Revaluation`, `app.Revaluer` | Rebuilding an account's holding valuations from a date to today, from the transaction log, prices, FX rates and index readings. Any change to those inputs enqueues one in the same database transaction, and pending ones coalesce per account from the earliest date. | recalculation, refresh |
| Missing input | valuation | `valuation.MissingReason` | A price, FX rate or index reading that a holding valuation needed and did not have. The holding valuation records the gap instead of failing the revaluation. | |
| Valued-through date | valuation | `app.ValuationStatus.ValuedThrough` | The last day up to which every holding of an account has a valuation: the day before its first missing input, or today. | |
| Index reading | valuation | `valuation.IndexReading` | One published reading of a period index for one month: CPI year on year for the month it measures, or the NBP reference rate for the month whose interest periods it sets. | |
| Bond series | jurisdiction/pl | `pl.Series` | A Polish retail treasury bond series code, such as EDO0536 (EDO maturing in May 2036). Valuation uses it through jurisdiction/pl's rules. | bond (ambiguous) |
| Bond type | jurisdiction/pl | `pl.BondType` | The product a bond series belongs to, named by the code's first three letters: OTS, ROR, DOR, TOS, COI, EDO, ROS, ROD. Fixes the term, interest period, indexation and payout. | series (for the type) |
| Series terms | jurisdiction/pl | `pl.Terms` | What one series' issue letter fixes: first-period rate, margin and early-redemption fee. Data supplied per series, never hard-coded. | |
| Interest period | jurisdiction/pl | `pl.PeriodStart` | One period of a bond's interest, counted from its own purchase date: a month (ROR, DOR), three months (OTS) or a year. | coupon period (for capitalized types) |
| Period index | jurisdiction/pl | `pl.Indices` | The published reading that sets an interest period's rate after the first: CPI year on year for inflation-indexed types, the NBP reference rate for ROR and DOR. | |
| Business day | jurisdiction/pl | `pl.IsBusinessDay` | A day that is neither a Saturday nor a statutory day off work in Poland (Sundays and the holidays the act on days off work lists), as the issue letters count them. The NBP reference rate for a ROR or DOR interest period is the one in force on the tenth business day before the month it starts in (`pl.ReferenceRateDay`). | working day |
| Margin | jurisdiction/pl | `pl.Terms.Margin` | The fixed rate a series adds to its period index. | spread |
| Early-redemption fee | jurisdiction/pl | `pl.Terms.EarlyRedemptionFee` | The per-bond fee for redeeming before maturity, capped at the bond's accrued interest except for coupon types after their first interest period, and not charged on a withdrawal from IKE or IKZE. OTS has none: redeemed early, it earns no interest. | penalty |
| Exchange price | jurisdiction/pl | `pl.Purchase.Price` | The price per bond of a series bought by exchanging (zamiana) maturing bonds, set in its issue letter: usually 99.90 PLN, 100 PLN for OTS. The discount below the nominal is taxed with the interest. | rollover price |
| Lot value | jurisdiction/pl | `pl.LotValue` | A bond lot's statutory value on a date: its gross accrued value, the interest in it, its purchase discount and the early-redemption fee redeeming it that day would charge in its wrapper. The pack turns them into net and liquidation values. | |
| OKI act | jurisdiction/pl | `pl.OKIAct` | What the act on OKI fixes for its yearly tax: the day it is in force from, the balance limit and the rate on the part of the average balance above it (D-76). Data, not in force until set. | |
| Connector | ingest | `ingest.Connector` | A plugin that brings data in through an ingest path. | integration |
| View | views | `views.View` | A plugin that presents the canonical record elsewhere and never writes back. | sync, export |
| Owned account | views | `ghostfolio.AccountUUID` | A Ghostfolio account the Ghostfolio view created, one per account, known by the UUID derived from the view instance's name and the account's ID. The view writes only into owned accounts. Inside the Ghostfolio adapter, activity is Ghostfolio's word for what the record calls a transaction. | managed account |
| Read API | views | `views` + `connectapi` | The public, read-only API: the `openfief.views.v1.ReadService` protobuf service, served as Connect and transcoded to REST under `/v1` (D-50). | |
| Allocation | views | `views.Allocation` | The portfolio's value in the reporting currency on a date, split into groups by asset class, currency, institution or sleeve (D-30). | breakdown |
| Allocation item | views | `views.AllocationItem` | One valued holding or cash balance of an account on a date: what an allocation groups. | position |
| Asset class | views | `views.AssetClass` | The kind of asset an allocation item is: `cash`, `treasury_bond` (a bond series) or `listed` (any other instrument: a stock or ETF in M0). | asset type |
| Date total | views | `views.DateTotal` | The sum of the stored valuations dated on one date, in the reporting currency and per original currency. | |
| Time-weighted return | views | `views.Performance` | The portfolio's return chained day by day from its value, holdings plus cash, across external cash flows, so money put in or taken out is not performance (D-61). | performance (alone), profit |
| Performance index | views | `views.PerformanceDay.Index` | The time-weighted return since a first day as an index that starts at 100 on it: 105 means 5% up. | |
| External cash flow | views | `views.CashFlow` | Money entering or leaving the portfolio from outside it: a deposit, a withdrawal or opening cash; an opening lot is one too, at its valuation. A transfer is not one. | contribution |
| Transfer | views | `views.Transfer` | A withdrawal from one of the owner's accounts matched with a deposit of the same amount and currency into another, at most five days apart: money moving inside the portfolio. | internal flow |
| Plugin | plugins | `plugins.Plugin` | The custom resource for one installed, named instance of a plugin image: the image and its manifest (the vendor's) and the instance's name and config (the owner's). Its status says whether the image is verified. | extension, addon |
| Grant | plugins | `plugins.Grant` | The custom resource holding what the owner approved for a plugin. | permission set |
| Catalog | plugins | `plugins.Catalog` | The file in the OpenFief repository that lists first-party and verified plugins, with pinned image digests and signer identities. | marketplace, registry |
| Verified plugin | plugins | `plugins.Verified` | A plugin whose catalog entry pins a version the author reviewed. Anything installed by image reference alone is unverified. | official, trusted |
| Trusted signer | plugins | `plugins.Signer` | An identity whose image signatures the operator accepts: OpenFief's CI, a catalog publisher, or one the owner added. | |
| Signer identity | plugins | `plugins.SignerIdentity` | Who made a signature: for a keyless signature, the OIDC issuer and subject in its Fulcio certificate (such as a GitHub Actions workflow at a ref); otherwise a public key. | key ID |
| Trust policy | plugins | `plugins.TrustPolicy` | Every trusted signer plus the catalog. It decides whether a signed image is a verified plugin, unverified but signed by a trusted signer, or refused. Nothing overrides a refusal. | allow-list (that is egress) |
| Manifest | plugins | `plugins.Manifest` | A plugin's declared capabilities, egress hosts, secrets, data tiers, config schema and, for a view plugin that serves browsers, its browser path. | |
| Secret role | plugins | `plugins.SecretMount.Role` | The name a manifest gives one Secret a plugin needs, such as `flex-token`. Each instance mounts its own Secret for it, named `<instance>-<role>`, so two instances of one image never share one (D-89). | secret name (that is `<instance>-<role>`) |
| Config host | plugins | `plugins.ConfigHost` | An egress host taken from a plugin instance's config instead of its manifest: the host of the https URL in a config field the manifest names, approved by the grant and reached through the egress proxy (D-75). | dynamic host |
| In-cluster service | plugins | `plugins.ClusterService` | The one Service in another namespace of the cluster that a plugin may reach directly, outside the egress proxy, at its ClusterIP: its manifest asks for one, the owner picks it at install and the grant approves it (D-74). | internal host |
| Egress allow-list | plugins | `plugins.EgressAllowList` | The hostnames a plugin may reach through the egress proxy. | whitelist |
| Capability | plugins | `plugins.Capability` | What a plugin may act as: `connector` or `view` in M0. Declared in its manifest and approved in its grant. | kind, role |
| Egress host | plugins | `plugins.EgressHost` | One hostname and port in a manifest, grant or egress allow-list. Never an IP address or a wildcard. | domain, URL |
| Browser path | plugins | `plugins.Manifest.BrowserPath` | The URL path prefix a view plugin serves to browsers, such as `/` for the built-in UI. Its manifest declares it and its grant approves it; the core, the only entry point for browsers, proxies it to the plugin (D-81). | route, mount point |
| Egress decision | plugins | `app.EgressDecision` | The egress proxy's answer to one plugin's HTTP CONNECT: the plugin, the destination, whether its egress allow-list allows it, and why. Resolved addresses are checked separately, before dialing. | ACL verdict |
| Data tier | plugins | `plugins.DataTier` | A named class of data a plugin asks to see. Only a name in M0; what each tier holds is undecided (P-33). | |
| Image reference | plugins | `plugins.ImageRef` | A plugin's container image, pinned by digest (`repository@sha256:...`), never by tag. | image tag |
| Effective permissions | plugins | `plugins.EffectivePermissions` | What OpenFief enforces for a plugin: only what its Plugin requests and its Grant approves. | |
| Workload | plugins | `plugins.Workload` | What the operator creates for one plugin, compiled from its effective permissions: identity, token, Secret mounts, environment, NetworkPolicy and egress allow-list. | deployment (the Kubernetes object is one part) |
| Rollout | plugins | `plugins.Rollout` | Which version of a plugin runs under its grant: the current one, the previously deployed one while an upgrade waits for approval, or none. | deployment (the Kubernetes object) |
| Pending approval | plugins | `plugins.Rollout.Pending` | What a plugin's current version requests that its grant does not approve. The owner approves it by adding it to the grant. | diff |
| Plugin identity | identity | `plugins.ServiceAccountName` | The ServiceAccount `plugin-<name>` in the `openfief` namespace, proven to the core by its projected token with audience `openfief-core`, and to the egress proxy by a second one with audience `openfief-egress`. | |
| Signing keys | plugins | `plugins.SigningKeys` | The cluster's ServiceAccount token issuer and the public keys that verify its tokens. The operator publishes them in the ConfigMap `openfief-sa-jwks`, which every plugin mounts, so a plugin can check the core's token without reaching the Kubernetes API (D-59). | JWKS (that is their file format) |
| Owner | identity | `identity.Owner` | The person who runs an installation: one individual investor. | user, customer |
| Script token | identity | `identity.ScriptToken` | A scoped, expiring, read-only personal API token. | API key |
| Script token secret | identity | `identity.FormatSecret` | What a script presents as its bearer token: `ofs_`, the script token's ID, `_` and a random key. Shown once when the token is issued; only the SHA-256 hash of the key is stored. | API key, password |
| Authenticator | identity | `identity.Authenticator` | A way for the owner to prove who they are before the core issues a session: a passkey (owner scope), or the demo authenticator (read scope, demo databases only). Every authenticator resolves to the one owner. | login provider, IdP |
| Passkey | identity | `identity.Passkey` | One of the owner's WebAuthn credentials. It works only on the host it was created on, its relying party ID. | security key, password |
| Session | identity | `identity.Session` | The owner's login on one host: an opaque, revocable bearer token (`ofsess_`, an ID, `_` and a random key) that the core issues after an authenticator succeeds. It expires when idle or old, carries a read or owner scope, and only the SHA-256 hash of its key is stored. | JWT, cookie (the view plugin's container for it) |
| Session scope | identity | `identity.SessionScope` | What a session may do: `read` (the read API, WhoAmI and logout) or `owner` (also passkeys, script tokens and enrollment links). | role, permission |
| Enrollment code | identity | `identity.EnrollmentCode` | A one-time, expiring permission to create a passkey on one origin, carried in the fragment of an enrollment link (`/enroll#code=ofenroll_...`). The CLI issues the first one; an owner session issues more. | invite, magic link |
| Origin | identity | `identity.Origin` | The scheme, host and port the owner's browser uses, such as `https://dev.example.com`. The core serves the owner only on origins on its allow-list; the host is the WebAuthn relying party ID. | URL, domain |
| Currency | kernel | `kernel.Currency` | An ISO 4217 alphabetic code, such as PLN or USD. | |
| Money | kernel | `kernel.Money` | An exact decimal amount in one currency. It keeps its original currency; arithmetic across currencies is an error, and conversion is an explicit FX step in valuation. | |
| Quantity | kernel | `kernel.Quantity` | An exact, possibly fractional number of units of an instrument. | |
| Date | kernel | `kernel.Date` | A calendar date with no time of day or time zone, written YYYY-MM-DD. | timestamp |
| Demo dataset | demo (not a context: tooling) | `demo.Dataset` | A synthetic portfolio with about a year of history and its valuation inputs, generated from a seed and a date, for showing OpenFief on a demo database. Its instruments have synthetic ISINs; only public data is real: bond series terms, GUS CPI and NBP rates. Loaded by `cmd/demo-seed`, which refuses any database without its demo marker that already holds data. | sample data, test portfolio |
| Demo marker | demo (not a context: tooling) | `demo.DatabaseState` | The row `demo = true` in the `demo` schema that marks a database as a demo database. Only the demo loader writes it, and only a marked database may be reset. | |

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
| Gateway API / HTTPRoute | The Kubernetes networking API that succeeds Ingress. A cluster's operator runs a Gateway, which listens for traffic and terminates TLS; an application attaches an HTTPRoute to it for its hostname and paths. |
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
| JWT / JWKS | A JSON Web Token (RFC 7519) is a signed set of claims, such as issuer, subject, audience and expiry; Kubernetes ServiceAccount tokens are JWTs. A JSON Web Key Set (RFC 7517) is a JSON list of the public keys that verify them. |
| kind / k3d / k3s | kind and k3d run Kubernetes clusters in containers for testing; k3s is a lightweight Kubernetes distribution. |
| Kustomize component | Kustomize builds Kubernetes manifests from a base plus overlays, with no templating; a component is an optional piece an installation opts into by listing it. |
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
| OKI | Osobiste konto inwestycyjne, a Polish personal investment account: no Belka tax inside it, and a yearly tax on the part of its average balance above a statutory limit (D-76). |
| OMS | Order management system: the component that validates, routes and tracks orders. |
| OPA / Rego | Open Policy Agent, a CNCF policy engine; Rego is its Datalog-inspired policy language. |
| Paper trading | Trading against a simulated account with fake money. |
| OTS / ROR / DOR / TOS | Polish retail treasury bond types: OTS 3-month fixed, ROR 1-year and DOR 2-year with monthly coupons at the NBP reference rate plus a margin, TOS 3-year fixed and capitalized yearly. |
| PIT | Podatek dochodowy od osób fizycznych, Poland's personal income tax; also the act that levies it. |
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
| SDK | Software development kit: the libraries plugin authors build on. OpenFief's are `sdk/go` (a separate Go module) and, later, `sdk/ts`, Apache-2.0 like the contracts (D-5). |
| Smokescreen | Stripe's open-source egress proxy, written in Go, which checks each client's HTTP CONNECT destinations against a per-client allow-list. |
| TIF | Time in force: how long an order stays working (for example day, or good-till-cancelled). |
| TokenRequest / TokenReview | Kubernetes APIs: TokenRequest issues a short-lived ServiceAccount token for a chosen audience; TokenReview checks whether a token is valid. |
| TWR / MWR | Time-weighted return chains the returns between cash flows, so deposits and withdrawals don't count as performance; it is the standard for comparing with a benchmark. Money-weighted return (an IRR, such as XIRR) weights by the money invested, so it reflects the investor's timing. |
| Vanguard | A Go library by Buf that serves one Connect service as REST, gRPC and Connect, translating between them (transcoding). |
| VWAP | Volume-weighted average price. |
| WebAuthn / passkey | The web standard for public-key authentication; a passkey is a WebAuthn credential stored on a device. |
| xAPI | XTB's former public trading API, shut down in 2025. |
| XTB | X-Trade Brokers, a Polish broker. |
