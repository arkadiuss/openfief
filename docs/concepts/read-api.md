---
title: Read API
description: How OpenFief serves its record to the built-in UI, view plugins and scripts as one read-only protobuf service, called over Connect or as REST under /v1, and how it totals valuations, builds the allocation measures the portfolio's time-weighted return and serves the model portfolio while it is on.
covers:
  - internal/views/**
  - proto/openfief/views/**
  - sdk/go/gen/openfief/views/**
  - gen/openapi/**
---

The [read API](../glossary.md) is how anything outside the core reads the record: sleeves, accounts, transactions, holdings, tax lots, cash, balance assertion checks, stored valuations, [price checks](../glossary.md), the allocation, the portfolio's [time-weighted return](../glossary.md) and, while it is on, the [model portfolio](../glossary.md). The [built-in UI](ui.md) reads through it, with the owner's session as its bearer token, every other view plugin is meant to, and so can the owner's own scripts. It only reads: data enters OpenFief through connectors, never through this API. This page explains the service, the pure rules behind its totals, the allocation and the performance series, and the HTTP adapter that serves it. The core service, `cmd/core`, serves it on both its ports (D-115): Service `openfief-core` port `http` (8080), where scripts and browsers arrive, and port `connect` (8081), where view plugins call it.

## Concepts

**A separate read side.** Bertrand Meyer's command-query separation (1988) says a method should either change state or answer a question, never both. Greg Young's CQRS (around 2010) applies the same split to whole systems: the write path and the read path get separate models. OpenFief uses a light version of it. Data enters only through connectors' ingest paths into the append-only transaction log (D-7, D-18). The read API answers from the ledger's queries and from stored valuations, which are a read model: rows derived from the log, prices and rates, rebuildable at any time (D-36).

**One service, two ways to call it.** The API is one protobuf service, `openfief.views.v1.ReadService`, in [read_service.proto](../../proto/openfief/views/v1/read_service.proto). Protobuf (Google, open-sourced 2008) describes messages and services once and generates code for many languages. Each method also carries a `google.api.http` annotation, Google's convention from its own APIs for mapping an RPC onto a REST path. Clients generated from the protobuf call the service over [Connect RPC](../glossary.md), gRPC or gRPC-Web. Everyone else calls plain REST with JSON under `/v1`, which Buf's [Vanguard](https://github.com/connectrpc/vanguard-go) library transcodes from the annotations into the same handlers. The OpenAPI 3.1 document, [read_service.openapi.yaml](../../gen/openapi/openfief/views/v1/read_service.openapi.yaml), is generated from the same definitions by `protoc-gen-connect-openapi`, with its title and bearer scheme from [read_service.openapi-base.yaml](../../proto/openfief/views/v1/read_service.openapi-base.yaml). So the Go types, the Connect service, the REST paths and the OpenAPI document can't disagree.

**Read-only by construction.** Every method is unary, declares `idempotency_level = NO_SIDE_EFFECTS`, and maps to a `GET` under `/v1` with no body. Over REST, any other HTTP method gets `405 Method Not Allowed` with `Allow: GET`, before the token is even looked at. The Connect protocol still accepts its usual `POST` for unary calls, and also its `GET` form, which the no-side-effects option enables.

| Method | REST | Parameters |
|---|---|---|
| `ListSleeves` | `GET /v1/sleeves` | |
| `ListAccounts` | `GET /v1/accounts` | |
| `GetAccount` | `GET /v1/accounts/{account_id}` | |
| `ListTransactions` | `GET /v1/accounts/{account_id}/transactions` | `pageSize` (1 to 500, 0 means 100), `pageToken` |
| `ListHoldings` | `GET /v1/accounts/{account_id}/holdings` | `date` |
| `ListLots` | `GET /v1/accounts/{account_id}/lots` | `date` |
| `ListCash` | `GET /v1/accounts/{account_id}/cash` | `date` |
| `ListAssertionChecks` | `GET /v1/accounts/{account_id}/balance-assertions` | `status`, such as `ASSERTION_STATUS_FLAGGED`; none for all |
| `ListAccountValuations` | `GET /v1/accounts/{account_id}/valuations` | `from`, `to` |
| `GetValuationStatus` | `GET /v1/accounts/{account_id}/valuation-status` | |
| `ListValuationTotals` | `GET /v1/valuations` | `from`, `to` |
| `ListValuationStatuses` | `GET /v1/valuation-status` | |
| `ListPriceChecks` | `GET /v1/price-checks` | `status`, such as `PRICE_CHECK_STATUS_FLAGGED`; none for all |
| `GetAllocation` | `GET /v1/allocation` | `date`, `by`, such as `ALLOCATION_DIMENSION_SLEEVE` |
| `GetPerformance` | `GET /v1/performance` | `from`, `to`, both optional |
| `GetModelPortfolio` | `GET /v1/model-portfolio` | `from`, `to`, both optional; `404` while the model portfolio is off |
| `ListPlugins` | `GET /v1/plugins` | `503` while the core can't read the cluster |

**Money, quantities and dates on the wire.** [Money](../glossary.md) is a `Money` message: `amount` as a decimal string such as `"-1234.56"`, never a JSON number, and `currency` as an ISO 4217 code. A JSON number is a binary float in most clients, which can't hold every decimal amount exactly. Quantities are decimal strings too. Dates are `YYYY-MM-DD` strings with no time of day.

**Original and reporting currency.** Every amount keeps its original currency. Valuations, [date totals](../glossary.md) and the [allocation](../glossary.md) also carry a value in the [reporting currency](../glossary.md): fields named `value_reporting` and `total_reporting`, with a `reporting_currency` field on the valuation responses. Only PLN is accepted as the reporting currency today (see [Decisions](#decisions)).

**Valuations, totals and their gaps.** The API reads the [holding valuations](../glossary.md) the valuation context stores ([Valuation](valuation.md)); it computes no prices itself. A holding valuation that lacks an input (a price, an FX rate or an index reading) is left out of the account's valuation list and counts in no total. How far each account is complete is its [valued-through date](../glossary.md), which `GetValuationStatus` and `ListValuationStatuses` return with whether a revaluation is pending or running. An account never valued has an empty `valued_through`. Stored valuations cover holdings only, never cash, so `ListValuationTotals` doesn't include cash; the performance series' daily value does.

**Allocation.** The [allocation](../glossary.md) is the portfolio's value on one date split into groups by one dimension: [asset class](../glossary.md), currency, [institution](../glossary.md) or [sleeve](../glossary.md). Each [allocation item](../glossary.md) is one valued holding or one cash balance of an account. A holding counts at its stored valuation on that date; cash, which has no stored valuation, is converted here at the FX rate for the date. A holding with no valued holding valuation on the date is not counted but listed in `unvalued`, so a chart can say what it is missing instead of silently showing less. The asset class is `cash`, `treasury_bond` for a Polish ISIN whose venue symbol is a bond series code such as `EDO0536`, and `listed` for anything else. Groups come largest first, then by key.

**Time-weighted return.** How did the portfolio do, when the owner keeps adding and taking out money? Its value alone can't say: a deposit raises it without anyone earning anything. Peter Dietz's *Pension Funds: Measuring Investment Performance* (1966) gave a formula that takes cash flows out of a return, and the Bank Administration Institute's 1968 study for US pension funds recommended the time-weighted rate of return for comparing managers, because a manager doesn't choose when money arrives. The CFA Institute's Global Investment Performance Standards (GIPS, 1999, grown out of AIMR's 1993 standards) require it. The idea: split the time into sub-periods at every external cash flow, measure each sub-period's return on the money that was there, and multiply the returns. Where the portfolio isn't valued at each flow, the Modified Dietz method approximates a sub-period's return by weighting each flow by the part of the period it was invested; GIPS has required a valuation at every large cash flow since 2010, which makes each sub-period exact. OpenFief has a stored valuation for every day (D-36), so every day is a sub-period. Valuations are end-of-day, so a flow's time inside its day is a convention: OpenFief takes [external cash flows](../glossary.md) at the end of their day. Each day's factor is (V − F) / V′, where V is the portfolio's value at the end of the day, F the day's net external flow and V′ the value at the end of the day before, which is Modified Dietz for one day with the flow weighted zero. The [performance index](../glossary.md) starts at 100 and multiplies by each day's factor, so 105 means 5% up, whatever was deposited or withdrawn. Money moved between two of the owner's own accounts is a [transfer](../glossary.md), not a cash flow (D-61). The model portfolio's backtest is drawn next to this line (D-53); it starts at 100 on the same day.

**The model portfolio.** While the owner keeps the [model portfolio](portfolio.md) off, the default, `GetModelPortfolio` answers `NOT_FOUND` (`404` over REST) with the fixed text `no model portfolio`, and no other method mentions it (D-52). While it is on, the answer carries its targets, its [backtest](../glossary.md) as `days`, an index that is 100 on the first day of the range like the performance series, the first and last day every target has history (`backtest_start`, `backtest_end`), the [max drawdown](../glossary.md) of the model over its whole backtest and each target's history with its own max drawdown (D-54), in the reporting currency. The range resolves like `GetPerformance`'s: go-live to the latest valued-through date by default, so both chart lines cover the same days. A model whose backtest starts after go-live starts its line at 100 on its first day. The values of a drawdown are those of the series it was measured on: the backtest's index from 100 on `backtest_start` for the model, a target's value in the reporting currency per unit for a target.

**Paging transactions.** `ListTransactions` returns the current version of each transaction in date, then source, then external ID order. A transaction the owner deleted (D-100) has no current version and is left out; interest, such as a bond's coupon, has the kind `TRANSACTION_KIND_INTEREST` (D-101). A page that has more after it returns a `nextPageToken`: an opaque cursor naming the last transaction's date and key, base64url-encoded. Passing it back starts the next page right after that transaction, so a client reads every current transaction once even if a page boundary falls between transactions on the same date. This is keyset paging: an offset ("skip 200") skips or repeats a row when rows arrive between two requests, while a cursor that names the last row read doesn't. The parameter names follow Google's API improvement proposal AIP-158.

**Bearer tokens.** Every call, over any protocol, needs an `Authorization: Bearer <token>` header. The token is a [session](../glossary.md) of the owner, of any scope, such as the one the built-in UI holds for the browser, or a [script token](../glossary.md); or a view plugin's projected ServiceAccount token ([plugin identity](../glossary.md)), such as the [Ghostfolio view](ghostfolio.md)'s. The API asks a `TokenVerifier` whether the token may read. The core's verifier accepts sessions and read-scoped script tokens through the identity context's `Auth.Authenticate`, and view plugins' tokens through the same check as the `AuthService`'s callers ([Identity](identity.md#in-the-core)). Logging in, and every other change, goes through the identity context's separate `AuthService`, never through this API.

## How it works

The views context follows the layering in [code-organization.md](../dev/code-organization.md):

| Package | Layer | Contents |
|---|---|---|
| `views` ([allocation.go](../../internal/views/allocation.go), [valuation.go](../../internal/views/valuation.go), [page.go](../../internal/views/page.go), [performance.go](../../internal/views/performance.go), [flows.go](../../internal/views/flows.go), [values.go](../../internal/views/values.go), [stale.go](../../internal/views/stale.go)) | Domain, pure | `ValueKind`, `HoldingValue`, `CashValue`, `CashOnDays`, `RemovedBehindHoldings`, `RemovedBehindLots`, `Allocate`, `ValueHoldings`, `AssetClassOf`, `Dimension`, `AssetClass`; `TotalsByDate`, `ValidateRange`, `MaxRangeDays`; `PageTransactions`, `TransactionCursor`, `DefaultPageSize`, `MaxPageSize`; `Performance`, `PerformanceInput`, `PerformanceSeries`, `PerformanceDay`, `PerformanceGap`, `TWRPlaces`; `ExternalFlows`, `SuggestTransfers`, `ConfirmedTransfers`, `LinkedTransfers`, `CashFlow`, `Transfer`. The model portfolio's rules are the portfolio context's ([Portfolio](portfolio.md)) |
| `views/app` ([views.go](../../internal/views/app/views.go), [performance.go](../../internal/views/app/performance.go), [model.go](../../internal/views/app/model.go), [ports.go](../../internal/views/app/ports.go), [tokens.go](../../internal/views/app/tokens.go), [values.go](../../internal/views/app/values.go)) | Use cases and ports | `Views` (with `AccountTerms`, `Settings`, `YearlyTaxes` and the `Terms` port) and its queries, `Views.Performance` and `Views.ModelPortfolio` among them; the ports `Ledger`, `Valuations`, `ValuationStatuses`, `FXRates`, `TokenVerifier`, `ModelReports` (set with `WithModel`) and `Connections` (set with `WithConnections`, [connections.go](../../internal/views/app/connections.go)); `UnsupportedReportingCurrencyError`, `ErrModelOff` |
| `views/connectapi` ([handler.go](../../internal/views/connectapi/handler.go), [server.go](../../internal/views/connectapi/server.go), [auth.go](../../internal/views/connectapi/auth.go), [errors.go](../../internal/views/connectapi/errors.go), [convert.go](../../internal/views/connectapi/convert.go), [values.go](../../internal/views/connectapi/values.go)) | Adapter | `New`, the HTTP handler; `readService`, the generated `ReadServiceHandler`; the auth interceptor, error mapping and conversion to the generated messages |

The generated Go code for the service lives with the other contracts in `sdk/go/gen/openfief/views/v1` ([Plugins](plugins.md#the-connector-contract)). From the contract to the ports:

```mermaid
flowchart TB
  proto["proto/openfief/views/v1<br/>ReadService + HTTP annotations"] -- "task proto:gen" --> gen["sdk/go/gen/openfief/views/v1<br/>messages, handler, Go client"]
  proto -- "task proto:gen" --> oas["gen/openapi<br/>OpenAPI document"]
  subgraph views["internal/views"]
    connectapi["connectapi<br/>Vanguard + Connect handlers"]
    app["app<br/>Views use cases, ports"]
    pure["views<br/>Allocate, PageTransactions, TotalsByDate"]
  end
  gen --> connectapi
  connectapi --> app
  app --> pure
  app -- "Ledger port" --> ledger["ledger/app.Ledger"]
  app -- "Valuations, ValuationStatuses, FXRates" --> vstore["valuation/postgres.Store"]
  app -- "TokenVerifier port" --> tokens["sessions and script tokens<br/>identity/app.Auth.Authenticate"]
  bin["cmd/core"] -.-> connectapi
```

### A request, from HTTP to the ports

`connectapi.New(views, verifier, opts...)` builds one `http.Handler`. The ports are declared in `views/app`, by the consumer. Each is shaped like a method the other contexts already have, and `cmd/core` passes `*ledgerapp.Ledger` as `Ledger` and `*valuation/postgres.Store` as `Valuations`, `ValuationStatuses` and `FXRates`.

```mermaid
flowchart TD
  client(["Generated client, script or plugin"]) --> reqlog["withRequestLog<br/>request ID, X-Request-ID"]
  reqlog --> sec["withSecurityHeaders"]
  sec --> vg["vanguard.Transcoder"]
  vg -- "REST: not GET" --> m405(["405, Allow: GET"])
  vg -- "REST GET /v1/..., Connect,<br/>gRPC, gRPC-Web" --> auth["authInterceptor"]
  auth -- "no Bearer token" --> u401(["unauthenticated"])
  auth -- "VerifyToken" --> tv["app.TokenVerifier"]
  auth --> rs["readService<br/>viewsv1connect.ReadServiceHandler"]
  rs --> views["app.Views"]
  views --> ledger["Ledger<br/>*ledgerapp.Ledger"]
  views --> vals["Valuations"]
  views --> stat["ValuationStatuses"]
  views --> fx["FXRates"]
  vals & stat & fx --> store["*valuation/postgres.Store"]
  views --> pure["views: PageTransactions,<br/>TotalsByDate, ValueHoldings, Allocate"]
  rs -- "any error" --> tce["toConnectError"]
```

The handler parses each request field: dates with `kernel.ParseDate`, the page token with `views.ParseTransactionCursor`, and enums against the domain's constants. An unspecified allocation dimension or an unknown assertion status is an invalid argument. `TestEnumsMatch` checks that every domain enum and its protobuf twin share their numbers.

### Errors, logs and headers

Errors never carry their own text to the client, because it can quote an account ID, an amount or what was asked. `toConnectError` maps them to Connect codes with fixed texts, and Vanguard turns the codes into HTTP statuses for REST:

| Cause | Code | REST | Client sees |
|---|---|---|---|
| `app.ErrUnknownAccount` | not found | 404 | `no such account` |
| `app.ErrModelOff` | not found | 404 | `no model portfolio` |
| `app.ErrPluginsUnknown` | unavailable | 503 | `the installed plugins can't be read now` |
| `views.ErrInvalidRange`, `ErrInvalidCursor`, `ErrInvalidPageSize`, `ErrUnknownDimension`, a malformed field | invalid argument | 400 | a fixed sentence per cause |
| No bearer token, or one the verifier rejects | unauthenticated | 401 | a fixed sentence and a `WWW-Authenticate: Bearer` challenge |
| Anything else, including a token that can't be checked | internal | 500 | `the request failed; its request ID names it in the server log` and the request ID |

`withRequestLog` gives each request a random ID, returns it as `X-Request-ID`, and logs one line with only `request_id`, the HTTP `method` and the `status`. The path is never logged: it can name an account. A failure adds a line with `request_id` and `while` (`answering` or `verifying a token`), and not the error, for the same reason. `withSecurityHeaders` sets `Cache-Control: no-store`, `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`, `Cross-Origin-Resource-Policy: same-origin`, `Referrer-Policy: no-referrer`, `X-Content-Type-Options: nosniff` and `X-Frame-Options: DENY` on every response, errors and 405s included.

### Account queries

Sleeves, accounts, holdings, lots, cash and balance assertion checks come straight from the `Ledger` port; [Ledger](ledger.md) explains how they are derived. A query about an account the ledger doesn't store fails with `ErrUnknownAccount`. Any other store failure stays a failure, so a broken database never looks like a missing account. `AssertionChecks` filters to one status when asked. `Transactions` loads the current transactions and hands them to `PageTransactions`, which sorts a copy, starts after the cursor, and returns at most `size` of them with a cursor when more follow. It then gives each one its origin, as the next section explains.

### Sleeves, accounts and connectors

What the owner manages in the built-in UI is listed here for any read caller: each `Account` with its [tax wrapper](../glossary.md) (D-76) and [fee rule](../glossary.md) (D-78), from `Views.AccountTerms`, the pack's default wrapper when none is recorded; `ListSleeveKinds` (`GET /v1/sleeve-kinds`) gives every [sleeve kind](../glossary.md) with its default role, `ListSleeves` each sleeve's `role_override` (D-71), `ListConnectorBindings` (`GET /v1/connector-bindings`) which connector instance may write to which account, and `ListBrokerAccounts` (`GET /v1/broker-accounts`) the [broker accounts](../glossary.md) the instances reported, each with the account it is mapped to, the institution, the day it was last seen and whether the instance's last reporting delivery had it; `unmapped=true` keeps only those not mapped, whose transactions are not booked (D-72). Kinds come from the `Ledger` port; bindings and broker accounts from the `Connections` port, which `cmd/core` fills with the ingest context's `*ingest/postgres.Store` (`WithConnections`); without it both lists are empty. A broker account's ID is an account number at the institution; any read caller may list it, as it may read the record itself.

### Installed plugins and removed instances

`ListPlugins` lists the installed [plugin](../glossary.md) instances (D-46, D-48), so the built-in UI can show them and how far each image is trusted (D-56). It only reads: nothing in the API installs, removes or approves a plugin, and the core can't change a grant (D-46). Each `Plugin` carries its instance name, its image and the image's digest, the capabilities its running version may use and those its manifest requests, a `trust` level with the signer's name, the operator's reason for it, and a `phase`: the operator's `Ready` reason, such as `Running` or `AwaitingApproval`, `Removing` while it is being uninstalled, or `Pending` before the operator reconciled it.

The [operator](operator.md#the-signature-and-manifest-check) records the trust in each `Plugin`'s status: the `SignatureVerified` condition and, while that is `True`, `status.signature` with the trust, the signer's kind and its name. `views.PluginOf` turns that into one `SignatureTrust` level:

| Status | Level |
|---|---|
| `True`, `status.signature.trust` `Verified` (a [verified plugin](../glossary.md), D-44) | `SIGNATURE_TRUST_VERIFIED` |
| `True`, signed by OpenFief's CI, a catalog publisher or a signer the owner added (D-45) | `SIGNATURE_TRUST_OPENFIEF_CI`, `_PUBLISHER`, `_OWNER` |
| `True` without `status.signature`, as an older operator writes it | `SIGNATURE_TRUST_TRUSTED_SIGNER`, or `_VERIFIED` when the reason is `Verified` |
| `False`: unsigned, an untrusted signer, a manifest that doesn't match (D-57) | `SIGNATURE_TRUST_REFUSED` |
| `Unknown` or no condition yet | `SIGNATURE_TRUST_NOT_CHECKED` |

The signer is named only for a trusted signature, so a refused image never shows one.

The list comes from the `Plugins` port, set with `WithPlugins`. `internal/views/kube` implements it: `kube.Plugins` watches the `Plugin` resources of the `openfief` namespace with a client-go informer and reads the fields above from them as unstructured data (`RecordOf`), so the views context doesn't import the operator's Go types; a test converts an operator-written `v1alpha1.Plugin` to check the field names. `cmd/core serve` starts it in a cluster with the core's ServiceAccount, which already has `get`, `list` and `watch` on Plugins for connector discovery; PluginGrants aren't read, since the `Plugin`'s status holds the effective capabilities. Outside a cluster, or until the watch has synced, `Plugins` fails with `ErrPluginsUnknown`, which the API answers as `UNAVAILABLE` (`503`).

**A removed instance's transactions (D-58).** Uninstalling an instance keeps the data it brought in. A connector instance's transactions have the key source `plugin:<instance>:<source>` ([Ingest](ingest.md#calling-a-connector-plugin)), so `views.ConnectorInstance` reads the instance back from it. `Transactions` asks the `Plugins` port which instances are installed and gives each transaction an `Origin`: `connector_instance`, and `connector_instance_removed` when no `Plugin` of that name exists any more. The mark is derived each time it is read, never written: the transaction log stays append-only (D-18), and reinstalling an instance under the same name makes its transactions current again, as its syncs continue from its own keys and watermark. While the installed plugins are unknown, nothing is marked removed and the listing still answers. A source that no connector instance wrote, such as the demo dataset's, has no instance.

**Holdings and lots no longer updated (D-91).** `ListHoldings` and `ListLots` work out the same mark for what is derived from the log, when read. A holding is `no_longer_updated` when any of the account's current transactions dated on or before the date that posts to its holding ledger account came from a removed instance (`views.RemovedBehindHoldings`). A lot is when the transaction that opened it, or a disposal that consumed part of it and left it open, did (`views.RemovedBehindLots`, which replays the lots FIFO with acquisitions before disposals on a date, as `ledger.LotsOn` does, and notes which disposal cut which lot). Each carries `removed_instances`, the names in order. While the installed plugins are unknown nothing is marked.

### Valuations and totals

`AccountValuations(account, from, to)` and `ValuationTotals(from, to)` first check the range with `ValidateRange`: `to` may not be before `from`, nor more than `MaxRangeDays` (3700, about ten years) after it. Then they load stored holding valuations through the `Valuations` port and the cash of each day (`views.CashOnDays`, a running sum of the current transactions' cash postings), for one account or for every account, and pass them to `TotalsByDate`. Totals include cash (D-106). Each day's cash is converted at the latest FX rate on or before it, half up, and valued in its account's [tax wrapper](../glossary.md) by the jurisdiction pack's `ValueCash` (only an IKZE's cash is taxed: 10% net, the marginal rate if liquidated); cash in a currency with no rate yet counts in no total but `by_currency` and sets `cash_unconverted`. `TotalsByDate` sums, per date, the three values (D-77) exactly in the reporting currency, and every original-currency amount per currency, in currency code order. A holding valuation that lacks its net or liquidation value, or cash whose liquidation value needs the owner's unset marginal rate, leaves that total unset, never a partial sum. A converted value in any other currency is `ErrNotReportingCurrency`. `AccountValuations` also returns the valued holding valuations themselves, each with its basis, source, gross value in its own currency and in the reporting currency, and its net and liquidation values or the reason each is missing (`fx_rate`, `lots` or `marginal_rate`).

### Settings and yearly taxes

`GetSettings` (`GET /v1/settings`) gives any reader the reporting currency (D-62), the [number format](../glossary.md) to write amounts in (D-113), the owner's or the pack's default with `number_format_default`, and the pack's tax wrappers, the default first. `ListYearlyTaxes` (`GET /v1/yearly-taxes?year=`) gives, for a year, each wrapper the pack taxes on its balance: OKI for Poland (D-76). While its act isn't in force that year, which is always today, the tax says so (`in_force` false) and carries nothing else; under an act it is computed from the daily gross totals, cash included, of every OKI account, summed per day of the year up to the last day every one of them is valued for. Both read the pack and the `Terms` port, which `cmd/core` fills with valuation's store through `WithJurisdiction`; without it there are no wrappers, the format is space-dot, cash is worth its amount in all three values and no yearly tax is listed.

### Price checks

`ListPriceChecks` serves D-22's cross-check: every [cross-check price](../glossary.md), such as IBKR's mark price, compared with the latest stored price of its instrument on or before its date (D-111). The `PriceChecks` port, which `cmd/core` fills with valuation's `CrossChecks` through `WithPriceChecks`, compares them ([Valuation](valuation.md#cross-checking-prices)); `Views.PriceChecks` only filters by status. Each `PriceCheck` carries the cross-check price's date, the instrument, both prices with the connector instance each came from, the price's own `price_date`, the `relative_difference`, (cross-check - price) / price as a decimal string rounded half-even to six places, its `status`, `matched`, `flagged` or `currency_mismatch` (no difference, D-111), the `threshold` it was flagged at, `"0.01"` or the owner's for the instrument (D-110), and whether the owner `acknowledged` it (D-112). The response's `threshold` is the default, `"0.01"`; without the port, outside `cmd/core`, it is empty and there are no checks. Prices aren't per account, so the method takes no account. `price_date`, `threshold`, `acknowledged` and `PRICE_CHECK_STATUS_CURRENCY_MISMATCH` were added after the method was first published; the owner changes thresholds and acknowledges checks through the `PriceCheckService`, not the read API.

### The allocation

```mermaid
flowchart TD
  start["Views.Allocation(on, d, kind)"] --> accts["Ledger.Accounts"]
  accts --> each["per account: allocationItems"]
  each --> hold["Ledger.HoldingsOn(account, on)"]
  each --> sv["Valuations.Valuations(account, on, on)"]
  hold & sv --> vh["views.ValueHoldings"]
  vh -- "the kind's value" --> items["AllocationItems"]
  vh -- "no valuation, or it lacks the kind's value" --> unv["Unvalued, with what it lacks"]
  each --> cash["Ledger.CashOn(account, on)"]
  cash --> isrep{"cash in the<br/>reporting currency?"}
  isrep -- "yes" --> conv["valuation.ConvertToPLN,<br/>ValueRounding"]
  isrep -- "no" --> rates["FXRates.FXRates(currency, on)"]
  rates --> conv
  conv -- "ErrNoRate" --> fail(["the call fails: internal"])
  conv --> cv["Pack.ValueCash in the<br/>account's wrapper, the kind's value"]
  cv -- "unknown: no marginal rate" --> unv
  cv --> citem["AllocationItem<br/>AssetCash"]
  citem --> items
  items --> alloc["views.Allocate(items, d, reporting)"]
  alloc --> out["AllocationOn<br/>Date, Allocation, Unvalued"]
  unv --> out
```

`GetAllocation` takes `value`, `VALUE_KIND_GROSS` (also when unspecified, D-77's default), `_NET` or `_LIQUIDATION`, and answers with the kind it split. `ValueHoldings` pairs each holding with the holding valuation of the same account, instrument and date, takes that kind's value in the reporting currency (`views.HoldingValue`; grouping by currency still uses the gross value's original currency), and its asset class from `AssetClassOf`; a holding without a valuation, or whose valuation lacks that value, is `Unvalued` with its reason, as is cash whose value of that kind needs the unset marginal rate. Cash is converted with the latest stored rate on or before the date, rounded to grosze like holding values (`valuation.ValueRounding`, half up), so holdings and cash add up alike. A cash currency with no rate fails the whole call rather than leaving that cash out. `Allocate` groups the items by the dimension's key: the asset class name, the original currency code, the account's institution or its sleeve ID. It sums each group and the total exactly, and sorts the groups.

So on the same date the allocation's total and `ListValuationTotals` both include cash; the allocation leaves out unvalued holdings that a later revaluation would add.

### The performance series

`Views.Performance(from, to)` loads every account's current transactions and valuation status, the stored valuations of the range and the FX rates of every currency the cash moves in, and hands them to the pure `views.Performance`. An empty `from` is go-live, the first transaction's date, and a `from` before go-live starts there, as the value-over-time chart does (D-30). An empty `to` is the latest valued-through date of any account. An account with transactions that was never valued counts as valued through the day before its first transaction. `revaluation_pending` says whether any account's revaluation is pending or running, so the series may still change.

```mermaid
flowchart TD
  start["Views.Performance(from, to)"] --> accts["Ledger.Accounts"]
  accts --> each["per account with transactions:<br/>CurrentTransactions, ValuationStatuses.Status"]
  each --> range["from: go-live at the earliest;<br/>to: latest valued-through date"]
  range --> vals["Valuations.Valuations(account, from, to)"]
  range --> rates["FXRates.FXRates(currency, to)<br/>per cash currency"]
  vals & rates --> perf["views.Performance"]
  perf --> flows["ExternalFlows: deposits, withdrawals,<br/>opening cash"]
  flows --> match["ConfirmedTransfers<br/>(the owner's links)"]
  match -- "transfers" --> transit["in transit between accounts"]
  match -- "the rest" --> ext["external cash flows"]
  perf --> days["per day: holdings + cash + in transit = value;<br/>flows + opening lots = flow"]
  transit & ext --> days
  days -- "a missing input" --> gap(["stop: GapUnvalued or GapNoFXRate"])
  days --> chain["factor (value - flow) / previous value,<br/>chained into the index"]
```

For each day from `from` to `to`, `views.Performance` values the portfolio at the end of the day, in the reporting currency:

- **Holdings** at their stored valuations. The transactions up to the day give each account's holdings; each must have exactly one valued holding valuation of that quantity on the day.
- **Cash**: each account's balance per currency, summed from its cash postings, converted with the latest stored rate on or before the day and rounded to grosze half up like holding values (`valuation.ValueRounding`). This is how the series includes the cash that `ListValuationTotals` leaves out.
- **In transit**: the money of each transfer from the withdrawal's date to the day before the deposit's, at the day's rate. When the deposit is booked first, it counts negative until the withdrawal, so the money counts once.
- **Flow**: the day's external cash flows at the day's rate, plus each opening lot at its share of its holding's valuation that day. An account seeded after go-live brings its opening lots and opening cash in as flows, so they are not returns.

Only a [transfer](../glossary.md) the owner confirmed is left out of the cash flows (D-105): `ConfirmedTransfers` takes the owner's current transfer links from the ledger and finds both legs among the day's deposits and withdrawals. A link counts only while its legs are still a withdrawal and a deposit of the opposite amount in different accounts; a correction that breaks that leaves the link out, and both legs are cash flows again. A pair that is only suggested, or that the owner rejected, is two cash flows. The response lists the confirmed transfers with a leg in the range.

Each day's factor is (value − flow) / the previous day's value, rounded to `TWRPlaces` (12) decimal places half-even, and the running product is rounded the same way. A day that starts with a value of zero or less has no return: its factor is 1. The index is 100 times the product, exact. Half-even rounding doesn't lean one way, so a synthetic ten-year series stays within a few billionths of an index point of the exact value.

The series stops before the first day that lacks an input and says why in `gap`: `PERFORMANCE_GAP_UNVALUED` for a day after some account's valued-through date, or with a holding that has no valuation, an incomplete one, or one of a quantity the log no longer gives (a revaluation not yet run); `PERFORMANCE_GAP_NO_FX_RATE` for cash, a flow or money in transit in a currency with no rate yet. A chart then ends at the last day it can draw truthfully instead of drawing a drop.

### Transfers

`ListTransfers` (`GET /v1/transfers`) is where the owner's confirmations start. `Views.Transfers` ([transfers.go](../../internal/views/app/transfers.go)) reads every account's current deposits and withdrawals and the ledger's current transfer links, and answers three lists:

- **Suggested** (`SuggestTransfers`): each withdrawal that is not a leg of a confirmed transfer, in date order, paired with the deposit into another account of exactly the same amount and currency, dated at most `ledger.TransferWindowDays` (five) calendar days before or after it, that is nearest in date, the earliest of equally near ones. The deposit must not be a leg of a confirmed transfer or of an earlier suggestion, and the pair must not be rejected; a rejected pair's withdrawal gets its next nearest deposit. Opening cash is never a leg.
- **Confirmed** and **rejected** (`LinkedTransfers`): the pairs whose current link has that decision, with their legs.

Each `Transfer` names its legs' accounts, dates and transaction keys, so the UI can hand a pair back to the ledger's owner-only `TransferService` ([Ledger](ledger.md#concepts)), which records the decision. The read API itself never records anything (D-7).

### The model portfolio

`Views.ModelPortfolio(from, to)` resolves the range as `Performance` does, asks the `ModelReports` port (`*portfolio/app.Models`, which `cmd/core` passes with `WithModel`) for the report through `to`, and rebases the backtest with `portfolio.Since` to 100 on the first day on or after `from`. With nothing recorded, there is no go-live: the line starts on the backtest's first day and runs through all stored history. A core built without `WithModel`, or a model that is off, is `ErrModelOff`.

## Decisions

- **D-6**: the read-only public API under `/v1` with an OpenAPI spec. Built in this context; the [built-in UI](ui.md) and the [Ghostfolio view](ghostfolio.md) read through it. MCP access is not part of it.
- **D-50**: protobuf with HTTP annotations, Connect plus REST through Vanguard, OpenAPI generated from the same definitions. Built as described above.
- **D-7**: data never enters through the public API. Enforced by the service definition and by REST answering 405 to anything but `GET`.
- **D-30**: the allocation by asset class, currency, broker or sleeve, and value over time. The code groups by institution where D-30 says broker, since the treasury bond registry is an institution and not a broker (see the glossary). The built-in UI draws value over time from `GetPerformance`, as the time-weighted return index, rather than from `ListValuationTotals`; the charts themselves are the UI's.
- **D-36**: the read API reports, per account, the valued-through date and whether a revaluation is pending or running. Built as `GetValuationStatus` and `ListValuationStatuses`, which the built-in UI shows per account.
- **D-61**: the portfolio's line is its time-weighted return, chained across deposits and withdrawals, with transfers between the owner's accounts not counted as cash flows. Built as `GetPerformance`.
- **D-105**: transfers are suggested and confirmed; only confirmed ones are not cash flows. Built as `ListTransfers` and `ConfirmedTransfers`. That a link whose legs no longer match after a correction counts for nothing is this context's reading.
- **D-52**: while the model portfolio is off, `GetModelPortfolio` answers `NOT_FOUND` and nothing else mentions it. The method itself is in the contract and the OpenAPI document either way; that it answers 404 rather than an empty response is this context's reading.
- **D-53**: the model's backtest as a second line, both starting at 100 on `from`. Built as `GetModelPortfolio`'s `days`.
- **D-62** and **T-12**: totals, the allocation and the performance series are in the reporting currency, next to each original currency. `app.New` refuses any reporting currency but PLN with an `*UnsupportedReportingCurrencyError`, because stored valuations and FX rates are against PLN only (D-20) and T-12 (proposal P-49) is open.
- **D-46**: the core only reads the installed plugins, with the `get`, `list` and `watch` on Plugins it already has; `ListPlugins` is how the UI shows them. Read sessions may list them too: no decision restricts it.
- **D-56**: the trust levels above. The task's split of "only signed" into OpenFief's CI, a catalog publisher and the owner needs `status.signature`, which the operator writes next to the condition; `refused` and `not checked` cover the images D-56's two levels don't.
- **D-58**, **D-91**: transactions of a removed instance stay and are marked, at read time from the plugin list, and so are the holdings and lots they are behind. Which transactions are behind a lot (its opening one and the disposals that cut it) is this context's reading. Balance assertions carry no mark.
- **D-77**, **D-106**: every valuation and total has gross, net and liquidation values; totals include cash; the allocation splits any of the three. A total with a missing value is left unset rather than summed over what has one.
- **D-76**, **D-78**, **D-113**: accounts carry their wrapper and fee rule, `GetSettings` the reporting currency and number format, `ListYearlyTaxes` OKI's yearly tax, not in force today.
- **D-22**, **D-51**: the cross-check, served as `ListPriceChecks`, additions only to the contract (a method, `PriceCheck`, `PriceCheckStatus` and the request and response). It serves every comparison with the threshold, and filters by status like `ListAssertionChecks`; the UI asks for the flagged ones.
- **D-27**: the bearer tokens the API accepts. The core accepts the owner's sessions and script tokens, and view plugins' projected tokens with audience `openfief-core` ([Identity](identity.md)). Connector plugins' tokens are refused: nothing decided gives a connector the read API.
- Half-up rounding of cash in the allocation and the performance series follows `valuation.ValueRounding`, which no decision fixes (see [Valuation](valuation.md#decisions)).
- Flows at the end of their day and 12 decimal places half-even for the factors were this context's choices; D-107 now records them.

## Verification

| Behavior | Tests | kibi |
|---|---|---|
| Grouping by each dimension, largest first; asset classes; pairing holdings with their valuations | `TestAllocate`, `TestAllocateEdges`, `TestAllocateInAnyReportingCurrency`, `TestParseDimension`, `TestAssetClassOf`, `TestValueHoldings` in [allocation_test.go](../../internal/views/allocation_test.go) | REQ-views-allocation, SCEN-views-allocation, TEST-views-allocation |
| Exact totals per date and per currency; inputs that lack a value count nowhere; the 3700-day range | `TestTotalsByDate`, `TestTotalsByDateOtherCurrency`, `TestValidateRange` in [valuation_test.go](../../internal/views/valuation_test.go) | REQ-views-totals-with-cash, SCEN-views-valuation-totals, TEST-views-valuation |
| The three values of a holding and of cash per kind; totals with cash and a missing value left unset; cash per day; the allocation of a kind | `TestValueKindNames`, `TestHoldingValue`, `TestCashValue`, `TestTotalsByDateWithCash`, `TestCashOnDays`, `TestValueHoldingsOfAKind` in [values_test.go](../../internal/views/values_test.go) | REQ-views-three-values, REQ-views-totals-with-cash, TEST-views-values |
| Holdings and lots behind a removed instance, the lot's opening and cutting transactions, none while unknown | `TestRemovedBehindHoldings`, `TestRemovedBehindLots` in [stale_test.go](../../internal/views/stale_test.go) | REQ-views-removed-instance-holdings-lots, REQ-views-holdings-lots-mark, TEST-views-stale |
| Account terms, settings, IKZE cash in totals with and without the marginal rate, the allocation of a kind, OKI's yearly tax with and without an act, holding and lot marks over the ports | `TestAccountTerms`, `TestSettings`, `TestTotalsWithIKZECash`, `TestAllocationOfAKind`, `TestYearlyTaxes`, `TestHoldingsAndLotsMarkRemovedInstances` in [app/values_test.go](../../internal/views/app/values_test.go) | REQ-views-three-values, REQ-views-settings, REQ-views-yearly-taxes, TEST-views-app-values |
| Settings, account terms, three values with cash, yearly taxes and the allocation of a kind over REST, against the OpenAPI document | `TestSettingsOverREST`, `TestValuesOverREST`, `TestYearlyTaxesOverREST`, `TestAllocationOfAKindOverREST` in [connectapi/values_test.go](../../internal/views/connectapi/values_test.go) | REQ-views-three-values, REQ-views-settings, TEST-views-connectapi-values |
| Pages in date then key order, cursor round trip, page size 1 to 500 | `TestPageTransactions`, `TestPageTransactionsSize`, `TestTransactionCursor` in [page_test.go](../../internal/views/page_test.go) | REQ-views-transaction-pages, SCEN-views-transaction-pages, TEST-views-page |
| Use cases over the ports: allocation inputs, cash without a rate, valuations, totals, status, filters, unknown account versus store failure, PLN only | `TestAllocation`, `TestAllocationFailures`, `TestReportingCurrency`, `TestAccountValuations`, `TestValuationTotals`, `TestValuationStatus`, `TestAssertionChecks`, `TestTransactions`, `TestUnknownAccount`, `TestStoreFailure` in [views_test.go](../../internal/views/app/views_test.go) | REQ-views-app-queries, REQ-views-allocation-inputs, REQ-views-valuation-queries, REQ-views-valuation-status, REQ-views-reporting-currency, with their SCEN-views-* scenarios, TEST-views-app |
| Time-weighted return: a price move counts; a deposit, a withdrawal, a transfer between own accounts, in either order and at the day's rate, and an account seeded later don't; only a confirmed transfer is not a cash flow; a missing valuation, a stale quantity, the valued-through date or a missing rate stops the series | `TestPerformance`, `TestPerformanceParts`, `TestPerformanceRejects`, `TestPerformanceGapString`, `TestExternalFlows` in [performance_test.go](../../internal/views/performance_test.go) | REQ-views-performance-twr, REQ-views-performance-value, REQ-views-performance-confirmed-transfers, REQ-views-performance-gaps, with their SCEN-views-performance-* scenarios, TEST-views-performance |
| Transfers: suggestions take the nearest free deposit within five days and skip taken, rejected and opening-cash legs; only confirmed links leave the cash flows, and a link whose legs no longer match counts for nothing; `GET /v1/transfers` matches the OpenAPI document | `TestSuggestTransfers`, `TestConfirmedTransfers` in [performance_test.go](../../internal/views/performance_test.go), `TestTransfers`, `TestTransfersFailWithTheLedger` in [app/transfers_test.go](../../internal/views/app/transfers_test.go), `TestRESTTransfers` in [connectapi/transfers_test.go](../../internal/views/connectapi/transfers_test.go) | REQ-views-transfer-suggestions, REQ-views-performance-confirmed-transfers, REQ-ledger-transfer-suggestion, REQ-portfolio-twr-confirmed-transfers-only |
| Properties over seeded random cases: the index is the same with an invested deposit of any size; cash alone stays at 100; the drift over 3700 days stays within its bound | `TestPerformanceIgnoresDepositSize`, `TestPerformanceOfCashIsFlat`, `TestPerformanceDrift` in [performance_property_test.go](../../internal/views/performance_property_test.go) | REQ-views-performance-twr, TEST-views-performance-properties |
| The performance query: go-live and valued-through defaults, a never-valued account, no transactions, pending revaluations, failures | `TestPerformance`, `TestPerformanceFailures` in [app/performance_test.go](../../internal/views/app/performance_test.go) | REQ-views-performance-query, SCEN-views-performance-query, TEST-views-app-performance |
| A generated Connect client and Connect's GET form; error codes; every method read-only in the descriptor; enums match the domain | `TestConnectQueries`, `TestConnectValuation`, `TestConnectPerformance`, `TestConnectPages`, `TestConnectErrors`, `TestConnectGET`, `TestServiceIsReadOnly`, `TestEnumsMatch` in [connect_test.go](../../internal/views/connectapi/connect_test.go) | REQ-views-api-contract, REQ-views-api-read-only, TEST-views-connectapi-connect |
| The model portfolio: its range, the rebased line, off and not served, a reversed range | `TestModelPortfolio`, `TestModelPortfolioRefusals` in [app/model_test.go](../../internal/views/app/model_test.go) | REQ-views-model-portfolio, REQ-views-model-line-from-go-live, REQ-portfolio-model-optional |
| The model portfolio over Connect and REST: the whole answer, 404 while off, the OpenAPI shape, enum numbers | `TestConnectModelPortfolio`, `TestModelPortfolioOff`, `TestRESTModelPortfolio`, `TestModelEnumsMatch` in [model_test.go](../../internal/views/connectapi/model_test.go) | REQ-views-model-portfolio, REQ-views-api-contract |
| Trust levels from the operator's status, phases, digests; reading the instance back from `ingest.ConnectorKeySource`; removed marks, and none while unknown | `TestPluginOf`, `TestImageDigest`, `TestConnectorInstance`, `TestOrigins` in [plugins_test.go](../../internal/views/plugins_test.go) | REQ-views-plugin-trust-level, REQ-views-removed-instance-mark, TEST-views-plugins |
| The plugin list in name order; unknown without the port or before the watch synced, and transactions still listed then; removed marks over the ports | `TestPlugins`, `TestPluginsUnknown`, `TestTransactionsMarkRemovedInstances` in [app/plugins_test.go](../../internal/views/app/plugins_test.go) | REQ-views-installed-plugins, REQ-views-removed-instance-mark, TEST-views-app-plugins |
| Plugins read from the operator's own `v1alpha1.Plugin`, deleting ones included, other namespaces not, later ones once watched | `TestRecordOf`, `TestPlugins`, `TestStartGivesUp` in [kube/plugins_test.go](../../internal/views/kube/plugins_test.go) | REQ-views-installed-plugins, TEST-views-kube |
| Price checks filtered by status, with the threshold; none without the port; a failure passed on | `TestPriceChecks` in [app/pricechecks_test.go](../../internal/views/app/pricechecks_test.go) | REQ-views-price-checks, TEST-views-app-pricechecks |
| `ListPriceChecks` over Connect and REST, against the OpenAPI document; an unknown status refused; no token refused; enum numbers | `TestConnectListPriceChecks`, `TestRESTPriceChecks`, `TestPriceCheckEnumsMatch` in [pricechecks_test.go](../../internal/views/connectapi/pricechecks_test.go) | REQ-views-price-checks, SCEN-views-price-checks, TEST-views-connectapi-pricechecks |
| `ListPlugins` and the transaction origins over Connect and REST, against the OpenAPI document; 503 without the list; enum numbers | `TestConnectListPlugins`, `TestConnectTransactionOrigins`, `TestRESTPlugins`, `TestRESTTransactionOrigins`, `TestRESTPluginsUnavailable`, `TestPluginEnumsMatch` in [connectapi/plugins_test.go](../../internal/views/connectapi/plugins_test.go) | REQ-views-installed-plugins, REQ-views-removed-instance-mark, TEST-views-connectapi-plugins |
| Sleeve kinds, bindings and broker accounts, all or unmapped, over the ports, Connect and REST against the OpenAPI document | `TestConnections`, `TestConnectionsWithoutThePortOrFailing` in [app/connections_test.go](../../internal/views/app/connections_test.go); `TestConnectListConnections`, `TestRESTConnections` in [connectapi/connections_test.go](../../internal/views/connectapi/connections_test.go) | REQ-views-onboarding-lists, TEST-views-app-connections, TEST-views-connectapi-connections |
| REST responses validated against the generated OpenAPI; 405 with `Allow: GET`; tokens; headers; money as strings; a log without data | `TestREST`, `TestRESTInternalError`, `TestRESTUnauthenticated`, `TestRESTReadOnly`, `TestOpenAPIIsReadOnly`, `TestSecurityHeaders`, `TestMoneyIsADecimalString`, `TestLogNamesNoData`, `TestRESTPerformance` in [handler_test.go](../../internal/views/connectapi/handler_test.go) | REQ-views-api-contract, REQ-views-api-read-only, REQ-views-api-auth, REQ-views-api-logs, TEST-views-connectapi-rest |

These requirements refine the firm ones imported from D-6, D-7, D-30, D-36 #6, D-50 (REQ-views-read-api-protobuf) and D-61 (REQ-portfolio-twr-comparison). Every test runs over in-memory fakes of the ports with synthetic accounts and ISINs, offline, under `task test`. The REST tests go through the real Vanguard transcoder and check each response body against `gen/openapi` with kin-openapi. `task proto:check` fails when the generated Go or OpenAPI differs from what the protobuf produces. Over the real stores, `TestCoreServesTheReadAPI` in [cmd/core](../../cmd/core/core_test.go), under `task test:pg`, starts the core on a fresh PostgreSQL with part of the demo dataset, issues a script token with `core token issue`, decodes REST answers strictly into the protobuf messages, checks that the performance series runs from 100 through today with no gap once the demo data is valued, and gets 401 without a token or with an unknown or revoked one.

## Not built yet

- **D-62**: a reporting currency other than PLN, waiting on T-12.
