---
title: Valuation
description: How OpenFief values every holding on every day, stores those valuations, and rebuilds them through a revaluation queue whenever a price, FX rate, index reading or the ledger changes.
covers:
  - internal/valuation/**
---

The valuation context turns holdings into money. For each account it stores the value of every holding on every day, in the instrument's currency and in PLN, and rebuilds those rows whenever one of their inputs changes. The inputs are end-of-day prices, exchange rates from the National Bank of Poland (NBP), period index readings for treasury bonds, and the ledger's transaction log. Bond values come from the Polish rules in [Treasury bonds](treasury-bonds.md); valuation picks which of them counts and adds them up.

## Concepts

### Prices, FX rates and index readings

A [price](../glossary.md) is the end-of-day close of one unit of an [instrument](../glossary.md), in the currency it trades in, with its source. M0 uses end-of-day prices only (D-21), from a market-data connector on the owner's own API key (D-22, D-51).

An [FX rate](../glossary.md) is the value in PLN of one unit of a currency on a date. [NBP table A](../glossary.md) is the central bank's daily table of average exchange rates against PLN, published on business days. Polish tax law converts foreign income at these rates (R-9), and M0 uses them for every conversion to PLN (D-20). There is no table on weekends or holidays, so the rate for a date is the latest one dated on or before it. NBP's API serves table A from 2002-01-02 (R-17).

A period index reading is one month's published value of a [period index](../glossary.md): CPI from GUS, or the NBP reference rate. Treasury bonds need them for interest periods after the first.

Every amount keeps its original currency ([money](../glossary.md)). Converting to PLN is an explicit step that names the rate and a rounding mode, and rounds once.

### Valuations and valuation bases

A [valuation](../glossary.md) is the value of a holding on a date, with its [valuation basis](../glossary.md) and source:

| Basis | Text form | Value | Used for |
|---|---|---|---|
| `Market` | `market` | Quantity times the end-of-day close, in the price's currency | Instruments with a price |
| `GrossAccrued` | `gross_accrued` | Nominal plus accrued interest, before any fee or tax | Treasury bonds |
| `NetIfRedeemed` | `net_if_redeemed` | What redeeming today would pay, after the early-redemption fee and Belka tax | Treasury bonds |

Polish retail treasury bonds are not traded, so they have no market price. They count at gross accrued or net-if-redeemed, and net-if-redeemed, what a redemption today would actually pay, is the default (D-25).

### Stored valuations and revaluation

D-36 makes valuations stored rows rather than values computed on each read. A stored row, `HoldingValuation`, is one holding of one account on one day on one basis: its quantity, its value in the instrument's currency and in PLN, its source, and the dates of the price and rate it used. When an input is missing, the row says which one (`price`, `fx_rate` or `index_reading`) and why, and has no value. A gap is recorded, not raised as an error.

Every row derives from the transaction log and the stored inputs, so it can be rebuilt at any time. A revaluation rebuilds one account's rows from a date to today. Any change to an input enqueues one, in the same database transaction as the change. Pending revaluations of one account coalesce into one, from the earliest affected date.

An account's valued-through date is the last day whose holdings all have a valuation: the day before its first missing input, or today. Together with "pending" and "running", it tells a reader how current the account's numbers are.

## How it works

The context follows the layering in [code-organization.md](../dev/code-organization.md):

| Package | Layer | Contents |
|---|---|---|
| `valuation` ([valuation.go](../../internal/valuation/valuation.go), [fx.go](../../internal/valuation/fx.go), [index.go](../../internal/valuation/index.go), [revaluation.go](../../internal/valuation/revaluation.go)) | Domain, pure | Prices, FX rates, bases, index readings; valuing one holding; `HoldingValuation`, `ValuedThrough`, `Revaluation`, `Coalesce` |
| `valuation/app` ([plan.go](../../internal/valuation/app/plan.go), [record.go](../../internal/valuation/app/record.go), [revalue.go](../../internal/valuation/app/revalue.go), [ports.go](../../internal/valuation/app/ports.go), [status.go](../../internal/valuation/app/status.go), [rates.go](../../internal/valuation/app/rates.go)) | Use cases and ports | `Planner`, `InputRecorder`, `Revaluer`, `ValuationStatus`, `FetchRates` |
| `valuation/postgres` ([store.go](../../internal/valuation/postgres/store.go), [queue.go](../../internal/valuation/postgres/queue.go), [worker.go](../../internal/valuation/postgres/worker.go), [status.go](../../internal/valuation/postgres/status.go), [0001_valuation.sql](../../internal/valuation/postgres/migrations/0001_valuation.sql)) | Adapter | The `valuation` schema, the River queue and its workers |
| `valuation/nbp` ([client.go](../../internal/valuation/nbp/client.go)) | Adapter | `Client`, which fetches table A for the `RateProvider` port |

### Valuing one holding

`Revaluer.Revalue(account, from)` loads the account's log, derives its holdings for each day with `ledger.HoldingsOn`, and values each one:

```mermaid
flowchart TD
  h["ledger.Holding on a day"] --> bond{"bond_terms stored<br/>for its ISIN?"}
  bond -- no --> po["PriceOnOrBefore(prices, instrument, day)"]
  po -- "none" --> mp["row: Missing = price"]
  po --> mv["MarketValuation: quantity × close, exact"]
  mv --> isPLN{"price in PLN?"}
  isPLN -- yes --> rnd["round to grosze"]
  isPLN -- no --> rate["RateOnOrBefore(rates, currency, day)"]
  rate -- "none" --> mf["row: Missing = fx_rate"]
  rate -- "two different rates" --> err["ErrConflictingRates: the run fails"]
  rate --> topln["ToPLN, half up"]
  rnd --> mrow["row: Basis market, Value, PLN, PriceDate, RateDate"]
  topln --> mrow
  bond -- yes --> lots["ledger.LotsOn: open lots"]
  lots --> each["per lot: LotIndices picks readings,<br/>price per bond = cost / quantity,<br/>pl.ValueLot in the account's TaxWrapper"]
  each -- "pl.MissingIndexError" --> mi["row: Missing = index_reading"]
  each --> bv["BondValuation: Net or Gross by BondBasis"]
  bv --> brow["row: sum of lots, in PLN, Source statutory"]
```

`ValueMarketHolding` takes the latest price on or before the day, multiplies it by the quantity exactly, and converts with the latest FX rate on or before the day, rounding half up (`ValueRounding`). `ValueBondHolding` values each open lot with `pl.ValueLot`: purchase date is the lot's open date, price per bond its PLN cost over its quantity, and `LotIndices` supplies a reading for each started period after the first (CPI of `pl.CPIMonth`, or the reference rate recorded for the month the period starts). An account's tax wrapper is `Taxable` unless one was recorded. The bond basis is the one the `Revaluer` was built with (`BondBasis`: net-if-redeemed when unset, `Market` rejected).

`RateOnOrBefore` skips other currencies and later dates. No rate is `ErrNoRate`; equal duplicates are fine; different rates for one date are `ErrConflictingRates`. With synthetic rates of 4.0000 on Friday 2026-03-06 and 4.0100 on Monday 2026-03-09, a Sunday valuation takes Friday's rate. Tax on income received that Monday also takes Friday's, because `pl.TaxRateCutoff` moves the cutoff to Sunday ([the tax FX rate date](treasury-bonds.md#the-tax-fx-rate-date)).

### Recording inputs and planning revaluations

`InputRecorder` stores each input through the `Recorder` port, which writes it and enqueues its revaluations in one transaction. It rejects bad input with `ErrInvalidInput`: a price that names no instrument or isn't positive, an FX rate for PLN or not positive, an unknown index kind or month, bond terms without an ISIN, an unknown tax wrapper. The `Planner` decides which accounts each change affects, by fixed rules:

| Change | Entry point | Accounts | From |
|---|---|---|---|
| Prices | `RecordPrices` | Each account whose log ever held the instrument | The earliest price's date per instrument |
| FX rates | `RecordFXRates` | Every account with a log | The earliest rate's date |
| Index readings | `RecordIndexReadings` | Every account with a log | The first day of the earliest reading's month |
| Bond series terms | `RecordBondTerms` | Each account whose log ever held that ISIN | Its first transaction |
| Tax wrapper | `RecordTaxWrapper` | That account | Its first transaction |
| Bond basis | `ApplyBondBasis` | Every account, only if the basis changed | Its first transaction |
| Ledger import | `postgres.Store.ImportHook` | Each account the import changed | `ImportSummary.EarliestAffected` |
| End of day | `EndOfDay` River job | Every account with a log | Its valued-through date, or its first transaction |

The `valuation` schema keys prices by instrument and date, rates by currency and date, readings by kind and month, and terms by ISIN; recording one again replaces it. An instrument whose ISIN has stored terms is valued as a bond.

### The revaluation queue and worker

Revaluations run on [River](https://github.com/riverqueue/river), a job queue that keeps its jobs in PostgreSQL, so a job can be inserted in the same transaction as the change that causes it. River's tables live in the `river` schema, and `postgres.Migrate` creates both schemas. The sequence for a ledger import:

```mermaid
sequenceDiagram
  participant L as ledger app.Ledger.Import
  participant H as postgres.Store.ImportHook
  participant DB as PostgreSQL
  participant W as postgres.RevalueWorker
  participant R as app.Revaluer
  L->>DB: append transaction versions, check balance assertions
  L->>H: hook(ctx, tx, ImportSummary), same transaction
  H->>H: app.ForImport(EarliestAffected), then Coalesce
  H->>DB: pg_advisory_xact_lock for the account
  H->>DB: valuation.dirty from_date = LEAST(from_date, new date)
  H->>DB: InsertManyTx RevalueAccount, unique per account while waiting or running
  L->>DB: COMMIT: the import and its job together
  DB-->>W: River hands the job to a worker on queue "valuation"
  loop at most 5 rounds
    W->>DB: claim: claimed_from = LEAST(claimed_from, from_date), from_date = NULL
    W->>R: Revalue(account, claimed date), if a date was pending
    R->>DB: read the log, prices, FX rates, readings, terms, tax wrapper
    R->>R: value every holding on every day to today, then ValuedThrough
    R->>DB: ReplaceValuations: delete rows from that date, insert new rows, set valued_through
    W->>DB: finish, under the account's lock: is from_date set again?
    alt a change arrived during the run
      DB-->>W: yes: another round
    else nothing new
      W->>DB: DELETE the dirty row and JobCompleteTx, one transaction
    end
  end
```

Coalescing has two halves. River keeps at most one `RevalueAccount` job per account among waiting and running jobs (unique by its args, which name only the account). The date lives beside it in `valuation.dirty`, and every enqueue lowers it with `LEAST`. Enqueueing and finishing take the same per-account advisory lock, so a change either lands before `finish` checks (and the job goes another round) or after the job completed (and River accepts a new job). A failed run returns an error and River retries it; `claimed_from` keeps the date. After five rounds of fresh changes the job snoozes and runs again at once without using up an attempt.

`Revalue` starts at `from`, or at the account's first transaction when that is later, and runs to `today()`. It replaces every stored row dated on or after `from` in one transaction, so running it again gives the same rows. `ValuedThrough` keeps an earlier stored date if it lies before the day before the start, since that gap was not revisited.

Worked example, from the synthetic `TestWorkerRevaluesEndToEnd`: 10 units of a USD ETF are bought on 2026-03-02 and today is 2026-03-04. A USD rate of 4 is recorded for 2026-03-02 and a close of 100 USD for 2026-03-03. The run stores 2026-03-02 as missing a price, then 4000 PLN on both later days, so the account is valued through 2026-03-01. When a close of 99 USD for 2026-03-02 arrives, its own revaluation stores 3960 PLN for that day and the account is valued through 2026-03-04.

### Status and configuration

`postgres.Store.Statuses` returns an `app.ValuationStatus` per account: `ValuedThrough` from `valuation.account_status` (zero if never revalued); `Pending` when a change waits unclaimed, a job is waiting or retryable, or claimed work has no running job; `Running` when a job runs now.

`postgres.Config` holds what D-36 makes configurable: `Concurrency` (River workers on the queue), `EndOfDay` (a daily time in a time zone), `BondBasis` and the `Today` date source. `NewClient` rejects invalid settings and builds the River client with both workers and the periodic end-of-day job. The basis is fixed per process: `ApplyBondBasis` at start-up records it in `valuation.settings` and revalues every account only when it changed.

### Fetching NBP rates

`app.RateProvider` is D-20's swappable rate source. `FetchRates` splits a range with `SplitRange` into pieces of at most the provider's limit, calls them in date order, stops at the first failure, and sorts the result by date and currency.

```mermaid
sequenceDiagram
  participant F as app.FetchRates
  participant N as nbp.Client
  participant API as api.nbp.pl
  F->>F: SplitRange(from, to, nbp.MaxRangeDays = 93)
  loop each DateRange, in date order
    F->>N: Rates(ctx, r.From, r.To)
    N->>API: GET /api/exchangerates/tables/A/from/to/?format=json
    alt 200
      API-->>N: JSON tables
      N-->>F: FXRates, mid parsed from its text, Source "nbp-a"
    else 404, no table in the range
      N-->>F: empty result
    else any other status
      N-->>F: ErrUnexpectedStatus, and FetchRates stops
    end
  end
  F->>F: stable sort by date, then currency code
```

The adapter keeps NBP's JSON shape inside the package. It parses `mid` from the number's text into a `kernel.Decimal`, never through a float. It rejects, as `ErrMalformedTable`, a body that isn't JSON, a table other than `A`, a bad date or currency code, and a rate that isn't positive. It sends `User-Agent: openfief`, because `api.nbp.pl` answers 403 to Go's default, and reads at most 4 MiB. The caller supplies the `http.Client`, including its proxy.

## Decisions

- **D-36**: stored valuations, River queue, coalescing, fixed triggers, configurable settings, basis change revalues all. Built in the valuation context, but not yet wired or exposed: no binary in `cmd/` runs the store, the import hook or the River client, and no read API or UI shows the status.
- **D-20**: NBP table A to PLN, swappable rate provider. Two gaps remain: NBP is a package of the core module (`internal/valuation/nbp`), not the FX plugin of D-20 and DESIGN.md's M0 diagram; and "tax always uses NBP rates" isn't enforced, since `RateOnOrBefore` ignores `Source`.
- **D-21**: end-of-day prices only.
- **D-25**: configurable bond basis, net-if-redeemed by default. The code also withholds no tax inside IKZE (see [Treasury bonds](treasury-bonds.md#decisions)).
- **D-14**: PostgreSQL holds the `valuation` and `river` schemas.
- **D-62** and **T-12**: conversion targets PLN only, and every stored row has a PLN value; D-62 allows assuming PLN only in the Polish pack's tax rules. T-12 (proposal P-49) is open.
- Rounding market values half up to grosze (`ValueRounding`) is not decided anywhere; its doc comment borrows D-37's bond rounding.

## Verification

| Behavior | Tests | kibi |
|---|---|---|
| Market value is exact; bond basis defaults to net and rejects market | `TestBasisNames`, `TestMarketValuation`, `TestBondBasis`, `TestBondValuation` in [valuation_test.go](../../internal/valuation/valuation_test.go) | REQ-valuation-market-value, REQ-valuation-bond-basis, TEST-valuation-valuation |
| Latest rate on or before a date; tax takes the day before; conversion rounds once | `TestRateOnOrBefore`, `TestRateOnOrBeforeConflicts`, `TestTaxRateUsesTheTableBeforeTheEvent`, `TestToPLN`, `TestConvertToPLN` in [fx_test.go](../../internal/valuation/fx_test.go) | REQ-valuation-fx-rate-lookup, REQ-pl-tax-fx-rate-date, REQ-valuation-pln-conversion, TEST-valuation-fx |
| Each started period gets its CPI or reference-rate reading | `TestIndexKindNames`, `TestLotIndices` in [index_test.go](../../internal/valuation/index_test.go) | REQ-valuation-lot-indices, SCEN-valuation-lot-indices, TEST-valuation-index |
| Holdings valued at market or as bond lots; gaps become rows; valued-through rule; coalescing | `TestPriceOnOrBefore`, `TestValueMarketHolding`, `TestValueMarketHoldingConflictingRates`, `TestValueBondHolding`, `TestValueBondHoldingRejects`, `TestValuedThrough`, `TestCoalesce` in [revaluation_test.go](../../internal/valuation/revaluation_test.go) | REQ-valuation-market-holding, REQ-valuation-bond-holding, REQ-valuation-missing-input, REQ-valuation-coalesce, TEST-valuation-revaluation |
| Fixed trigger rules, input validation, basis change | `TestPlanner`, `TestRecordInputs`, `TestRecordRejectsInvalidInputs`, `TestApplyBondBasis` in [plan_test.go](../../internal/valuation/app/plan_test.go) | REQ-valuation-revaluation-triggers, REQ-valuation-input-validation, REQ-valuation-basis-change, TEST-valuation-app-plan |
| A revaluation rebuilds rows idempotently from its start and keeps an earlier gap | `TestRevalueStoresMarketValues`, `TestRevalueIsIdempotent`, `TestRevalueStartsAtTheFirstTransaction`, `TestRevalueKeepsAnEarlierGap`, `TestRevalueMissingInputs`, `TestRevalueBonds` in [revalue_test.go](../../internal/valuation/app/revalue_test.go) | REQ-valuation-revaluation-rows, SCEN-valuation-revaluation-rebuild, TEST-valuation-app-revalue |
| Inputs and imports enqueue in their own transaction; coalescing in `dirty`; status flags | `TestSaveEnqueuesInTheSameTransaction`, `TestEnqueueCoalescesFromTheEarliestDate`, `TestImportHookEnqueuesInTheImportsTransaction`, `TestSaveBondBasisRevaluesOnlyOnAChange`, `TestStatuses` in [queue_test.go](../../internal/valuation/postgres/queue_test.go) | REQ-valuation-enqueue-same-transaction, REQ-valuation-river-jobs, REQ-valuation-status, TEST-valuation-postgres-queue |
| Inputs and valuations round-trip; replacing rows | `TestInputsRoundTrip`, `TestReplaceValuations` in [store_test.go](../../internal/valuation/postgres/store_test.go) | REQ-valuation-schema, TEST-valuation-postgres-store |
| Real River workers revalue, close a gap, run end of day; settings checked | `TestWorkerRevaluesEndToEnd`, `TestNewClientChecksTheConfig`, `TestDailyAt` in [worker_test.go](../../internal/valuation/postgres/worker_test.go) | REQ-valuation-settings, SCEN-valuation-river-jobs, TEST-valuation-postgres-worker |
| Rate ranges split per provider limit; NBP parses exactly, no table is empty | `TestSplitRange`, `TestFetchRates` in [rates_test.go](../../internal/valuation/app/rates_test.go); `TestRatesTableA`, `TestRatesNoData`, `TestRatesKeepsEveryDigit`, `TestRatesRejects` in [client_test.go](../../internal/valuation/nbp/client_test.go) | REQ-valuation-fx-range-splitting, REQ-valuation-nbp-decimal-exact, REQ-valuation-nbp-request, TEST-valuation-app-rates, TEST-valuation-nbp-client |

These kibi tests also validate the firm D-36 requirements (REQ-decision-d36-*), except REQ-decision-d36-6-api-valuation-status and REQ-decision-d36-7-ui-valuation-status, and those imported from D-20, D-21 and D-25. The app tests use in-memory fakes. The NBP tests replay [hypert](https://github.com/areknoster/hypert) recordings of the public `api.nbp.pl` from `nbp/testdata/`, with no network.

`task test` runs everything offline; the PostgreSQL tests skip there, because `OPENFIEF_PG_DSN` is unset. `task test:pg` runs every `internal/*/postgres` package against a disposable server: the one `OPENFIEF_PG_DSN` names, or a throwaway cluster started from local PostgreSQL binaries on a Unix socket. Each test creates its own database with every valuation and River migration applied, and drops it.

One behavior differs from its doc comment. `RateOnOrBefore` says two different rates are an error "on the chosen date", but it checks while scanning, against the latest date seen so far: a conflict on an earlier date fails the lookup if it comes first in the slice. Stored rates are unique per currency and date, so a revaluation can't meet it.

## Not built yet

- **D-36**: the read API and UI parts. `Store.Statuses` computes each account's status, but no read API or UI exists to show it. Nothing in `cmd/` runs `Migrate`, the store, the import hook, the River client or `ApplyBondBasis`, or reads `Config` from settings.
- **D-49**, **D-20**: passing connector deliveries on. The ingest context's `PriceStore`, `FXRateStore` and `IndexReadingStore` ports have no implementation that calls `InputRecorder`, and the NBP adapter is not a connector plugin. Nothing fetches rates, CPI or reference rates on a schedule.
- **D-22**, **D-51**: the EODHD market-data connector, and the cross-check against IBKR Flex prices.
- **D-62**: the reporting currency setting (`valuation.ReportingCurrency` in the glossary).
- **D-52**, **D-53**: market index levels (`valuation.Index` in the glossary) for the model portfolio.
