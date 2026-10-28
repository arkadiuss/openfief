---
title: Portfolio
description: Sleeves and sleeve kinds, the defensive and offensive parts of the portfolio that every account belongs to, and the model portfolio, its backtest and max drawdowns.
covers:
  - internal/portfolio/**
---

The portfolio context describes the portfolio as a whole: the parts it is divided into, and later the model it is compared with. It holds [sleeves](../glossary.md) and [sleeve kinds](../glossary.md): every account belongs to exactly one sleeve, every sleeve has one kind, and each sleeve is either defensive or offensive. It also holds the optional [model portfolio](../glossary.md) (D-52 to D-54): its [targets](../glossary.md), its [backtest](../glossary.md) and the [max drawdown](../glossary.md) of the model and of each target.

## Concepts

**Sleeves.** In asset management, a sleeve is a part of a portfolio that is run separately, with its own mandate. In OpenFief a sleeve is a part of the portfolio with one role, and it is made of whole [accounts](../glossary.md). Section 3.1 of [DESIGN.md](../DESIGN.md) gives the reason the core keeps sleeve attribution itself: the later lease ledger reconciles against it.

**Why defensive and offensive.** The split resembles core-satellite investing: a large, diversified core held for the long run, plus smaller satellites for active bets. OpenFief turns that habit into a hard line for agents. GR-6 leases each thesis agent's budget only from the offensive part, and a defensive sleeve is never available to agents. Because a sleeve is made of accounts at institutions, the line can later be drawn by the broker too. P-11 (a), a proposal, puts each offensive sleeve in its own IBKR linked cash account, so even a compromised OpenFief host couldn't trade defensive accounts.

**Sleeve kinds.** The role is a two-value switch that GR-6 relies on, so it stays fixed. A sleeve kind is the owner-facing label on top: emergency fund, bonds, retirement, core, satellite, agent pool, or one the owner adds. Each kind has a default role, and the allocation chart can group by kind (D-30). The owner may rename a kind, but its default role never changes, because every sleeve of that kind was checked against it. The role is stored on the sleeve itself, so renaming a kind can't change what agents may lease from. A sleeve may always be *more* defensive than its kind; making it offensive in a defensive kind needs an explicit override.

**Sleeves in future fiefs.** These are proposals in [control-plane.md](../proposals/control-plane.md), not decisions. A [fief](../glossary.md) is a revocable lease from the offensive sleeve to one agent. P-12 sweeps sale proceeds above the lease back to the offensive sleeve. P-11 (f) adds kill switches per agent, per sleeve and global. In P-24, a closed fief returns its capital to the offensive sleeve. Still open: whether there is one offensive sleeve or several (OQ-20, T-7), and who sets its size (OQ-12).

**The model portfolio.** A weighted basket of [targets](../glossary.md) for the whole portfolio, not per sleeve (D-52). A target is an instrument or an [index](../glossary.md) with a weight, and the weights sum to 100%. It is something to compare with; in M0 nothing trades. Its performance is a [backtest](../glossary.md): its value from end-of-day prices, as far back as every target has data, rebalanced to the target weights once a year (D-53). Without rebalancing, a basket slowly becomes a bet on its best performer. The portfolio's own line next to it is its time-weighted return ([TWR](../glossary.md), D-61), so deposits and withdrawals don't count as performance, and transfers between the owner's own accounts, defensive to offensive included, are not cash flows. Both lines are in the [reporting currency](../glossary.md) (D-62) and start at 100 on go-live (D-53); M0 also shows the [max drawdown](../glossary.md) of the model and of each target (D-54).

## How it works

### Built: sleeves and sleeve kinds

Package `portfolio` ([sleeve.go](../../internal/portfolio/sleeve.go), [kind.go](../../internal/portfolio/kind.go)) is pure domain code:

| Name | What it does |
|---|---|
| `Sleeve` | `ID`, `Name`, `Kind`, `Role` and `OffensiveOverride`. `Validate` checks the ID, name and role. |
| `Role`, `ParseRole` | `RoleDefensive` or `RoleOffensive`, stored as those words. The zero `Role` is not a role. |
| `Kind`, `KindKey` | A key of lowercase hyphenated words (at most 64 characters), a name and a `DefaultRole`. |
| `DefaultKinds` | The built-in kinds: `agent-pool` and `satellite` offensive; `bonds`, `core`, `emergency-fund` and `retirement` defensive. |
| `CheckKind` | A new key adds a kind, a known key renames it; changing a stored kind's default role is `ErrKindRoleChange`. |
| `CheckSleeve` | Checks a sleeve against the stored kinds and fills in a missing role (diagram below). |
| `DefaultSleeves`, `MissingDefaults` | An install's starting sleeves, Bonds, Core and Satellite, and which defaults are not stored yet. |
| `KindBySleeve` | Maps each sleeve to its kind, for an allocation grouped by kind (D-30). |

```mermaid
flowchart TD
  s(["CheckSleeve: sleeve and stored kinds"]) --> k{"kind stored?"}
  k -- no --> e1["ErrUnknownKind"]
  k -- yes --> r{"role given?"}
  r -- no --> d["role = kind's default role"] --> v
  r -- yes --> v{"role vs kind's default"}
  v -- "same, or defensive in an offensive kind" --> o1{"OffensiveOverride set?"}
  o1 -- no --> ok["stored as is"]
  o1 -- yes --> e2["ErrRoleOverride: the flag means one thing only"]
  v -- "offensive in a defensive kind" --> o2{"OffensiveOverride set?"}
  o2 -- yes --> ok
  o2 -- no --> e3["ErrRoleOverride: needs the owner's explicit override"]
```

The portfolio context has no store of its own in M0, so the ledger keeps sleeves and kinds next to the accounts that refer to them, and the `ledger` package imports `portfolio`, not the other way round. [0002_sleeve_kinds.sql](../../internal/ledger/postgres/migrations/0002_sleeve_kinds.sql) creates `ledger.sleeve_kinds`, seeds the six built-in kinds, and adds a trigger that refuses any change to a kind's key or default role. Each sleeve gets a required `kind` and an `offensive_override` that a `CHECK` allows only on offensive sleeves; sleeves stored before kinds existed became `core` or `satellite`, whichever matches their role. In `internal/ledger/app`, `SaveKind` runs `CheckKind`, `SaveSleeve` runs `CheckSleeve`, and `SaveAccount` refuses an account whose sleeve isn't stored ([ledger.go](../../internal/ledger/app/ledger.go), [kinds.go](../../internal/ledger/app/kinds.go)). `EnsureDefaults` stores whichever default kinds and sleeves are missing, in one transaction, and never changes a stored one.

```mermaid
flowchart LR
  subgraph built["Built: default sleeves, with their kinds"]
    bonds["Sleeve Bonds<br/>kind bonds, defensive"]
    core["Sleeve Core<br/>kind core, defensive"]
    sat["Sleeve Satellite<br/>kind satellite, offensive"]
    a1["Account: bond registry"] --> bonds
    a2["Account: broker A, main"] --> core
    a3["Account: broker A, linked cash"] --> sat
  end
  subgraph later["Proposed for M1 and later, not decided"]
    pool["An agent-pool sleeve"]
    f1["Fief: thesis X"]
  end
  pool -. "leases, GR-6 and P-11" .-> f1
  sat -. "offensive too: undecided" .-> f1
  core -. "never available to agents" .-x f1
```

### Built: the model portfolio's rules

The rules are pure functions in package `portfolio` ([model.go](../../internal/portfolio/model.go), [backtest.go](../../internal/portfolio/backtest.go), [history.go](../../internal/portfolio/history.go), [drawdown.go](../../internal/portfolio/drawdown.go)):

| Name | What it does |
|---|---|
| `Model`, `Model.Validate` | `Enabled` (off is the zero value, D-52) and `Targets` in the owner's order. At most `MaxTargets` (50), none twice, weights summing to exactly 100%; an off model may have none, an on one needs some. |
| `Target`, `TargetKind`, `Target.Key` | `TargetInstrument`, named `ISIN/symbol` as prices are keyed, or `TargetIndex`, named as the market-data connector names it (such as `WIG20`) and with the currency of its levels. `Weight` is in percent, above 0 and at most 100, in steps of 0.01 (`WeightPlaces`). |
| `Backtest`, `BacktestInput`, `BacktestSeries` | The backtest (D-53), described below. |
| `TargetHistory`, `MissingData` | Each target's value in the reporting currency per day, or why it has none: `MissingCloses` or `MissingRates`. |
| `MaxDrawdown`, `Drawdown` | The worst fall from a peak to a later trough, with both dates and values and the depth as a fraction (D-54). |
| `Since` | A series rebased to 100 on its first day on or after a date: the model's line from go-live (D-53). |

**A target's history.** Each day from the target's first close (and, in a foreign currency, its first FX rate) to its last close is valued at the latest close times the latest rate on or before it, as holdings are valued. A close and a rate hold over days without one, such as weekends. Rates come in as `Rate` values into the reporting currency; with PLN reporting they are NBP's table A (D-20), which starts on 2002-01-02, so a foreign target's history starts no earlier (D-61).

**The backtest.** It starts on the first day every target has a history, at 100, holding each target at its weight: `units = weight / 100 × value / the target's value`. Each day's value is the sum of units times values. At the end of every December 31 before its last day it rebalances to the weights. Without rebalancing, a basket slowly becomes a bet on its best performer. It ends on the last day every target has a history. Units keep 18 decimal places and the index 12, both rounded half-even like the time-weighted return. A target without a history, or histories that don't overlap, give no days.

```mermaid
flowchart LR
  closes["Closes and index levels<br/>per target"] --> hist["targetHistory<br/>close × rate, carried over"]
  rates["FX rates into the<br/>reporting currency"] --> hist
  hist --> span["commonSpan<br/>first and last day all have"]
  span --> sim["simulate<br/>100 at the weights;<br/>rebalance each Dec 31"]
  sim --> days["BacktestSeries.Days"]
  days --> dd["MaxDrawdown: the model"]
  hist --> tdd["MaxDrawdown: each target"]
  days --> since["Since(go-live)<br/>the chart's line from 100"]
```

**Max drawdown.** `MaxDrawdown` walks the series once, keeping the highest value so far; each later value's fall from it is compared exactly, by cross-multiplying, and only the worst one's depth is divided, rounded half-even to 12 places. Of equal falls it keeps the earliest. A series that never falls has depth 0 with its first day as both peak and trough.

### Not built: storing, serving and showing the model

```mermaid
flowchart LR
  owner(["Owner, built-in UI"]) -- "targets and weights, D-52" --> store["Stored model<br/>off by default"]
  store --> bt["portfolio.Backtest"]
  bt --> api["Read API, D-6"]
  api --> chart["Value-over-time chart<br/>two lines from go-live, both at 100"]
  classDef notbuilt stroke-dasharray: 5 5
  class store,api,chart notbuilt
```

## Decisions

- **GR-6**: one fixed role per sleeve; offensive in a defensive kind needs an override.
- **D-30**: `KindBySleeve` lets the allocation group by kind; the chart isn't built.
- **D-52, D-53, D-54**: the model's targets and weights, its backtest and max drawdowns are built as pure rules. When it rebalances (each December 31), the 0.01% weight step, the 50-target limit and the precision are this context's reading, not the author's.
- **D-61, D-62**: the backtest is in the reporting currency and a foreign target starts with its first FX rate.
- **No decision ID for sleeve kinds.** The commit and the migration call them an "author decision", and the glossary defines the term, but DESIGN.md records no D-ID for the built-in kinds, the default sleeves or the override.
- **Which offensive sleeves agents may lease from.** The glossary makes every offensive sleeve "the part fiefs are leased from". [kind.go](../../internal/portfolio/kind.go) says only `agent-pool` is what fiefs are leased from and `satellite` holds the owner's own higher-risk positions, yet both are offensive and the default Satellite sleeve is offensive. Nothing in the code tells them apart; OQ-20 is still open.

## Verification

| Behavior | Tests | kibi |
|---|---|---|
| A sleeve needs an ID, a name and one known role | `TestSleeveValidate`, `TestParseRole` in [sleeve_test.go](../../internal/portfolio/sleeve_test.go) | REQ-portfolio-sleeve-role |
| Kind keys, the built-in kinds, fixed default roles, the override rules, the defaults and kind by sleeve | `TestKindValidate`, `TestDefaultKinds`, `TestCheckKind`, `TestCheckSleeve`, `TestDefaultSleeves`, `TestMissingDefaults`, `TestKindBySleeve` in [kind_test.go](../../internal/portfolio/kind_test.go) | REQ-portfolio-sleeve-kind, REQ-portfolio-default-kinds, REQ-portfolio-kind-role-fixed, REQ-portfolio-sleeve-role-override, REQ-portfolio-allocation-by-kind |
| The ledger checks kinds and sleeves before storing; defaults are ensured once | `TestSaveKind`, `TestEnsureDefaults` in [kinds_test.go](../../internal/ledger/app/kinds_test.go); `TestSaveSleeve`, `TestSaveAccount` in [ledger_test.go](../../internal/ledger/app/ledger_test.go) | REQ-ledger-app-sleeve-kinds, REQ-ledger-app-registry |
| PostgreSQL seeds the kinds, backfills old sleeves, refuses role changes and stray overrides | `TestMigrateStoresTheDefaultKinds`, `TestMigrateGivesStoredSleevesAKind`, `TestKindRoundTrip`, `TestSleeveKindConstraints`, `TestEnsureDefaultsOnPostgres` in [kinds_test.go](../../internal/ledger/postgres/kinds_test.go); `TestRegistryRoundTrip` | REQ-ledger-pg-sleeve-kinds, REQ-ledger-pg-store |
| An account needs a sleeve | `TestAccountValidate` in [account_test.go](../../internal/ledger/account_test.go) | REQ-ledger-account-sleeve |
| Targets, weights and the 100% sum | `TestModelValidate`, `TestTargetKind`, `TestTargetKey` in [model_test.go](../../internal/portfolio/model_test.go) | REQ-portfolio-model-targets, REQ-portfolio-model-weights |
| Histories in PLN, the start, the end, the year-end rebalance, gaps and bad history | `TestBacktest`, `TestBacktestTargetHistories`, `TestMissingDataString` in [backtest_test.go](../../internal/portfolio/backtest_test.go) | REQ-portfolio-backtest-history, REQ-portfolio-backtest-index, REQ-portfolio-backtest-yearly-rebalance |
| Max drawdown and rebasing | `TestMaxDrawdown`, `TestSince` in [drawdown_test.go](../../internal/portfolio/drawdown_test.go) | REQ-portfolio-drawdown-depth, REQ-portfolio-rebase, REQ-portfolio-max-drawdown |
| Properties: one target is its price, price scale doesn't matter, the drawdown is the worst pair | `TestBacktestOfOneTargetIsItsPrice`, `TestBacktestIgnoresPriceScale`, `TestMaxDrawdownIsTheWorstPair` in [backtest_property_test.go](../../internal/portfolio/backtest_property_test.go) | REQ-portfolio-backtest-index, REQ-portfolio-drawdown-depth |

Run `go test ./internal/portfolio/... ./internal/ledger/app/...` or `task test`; the PostgreSQL tests run with `task test:pg`.

From reading the code: the override guards only one path to offensive. Saving a sleeve again with an offensive kind, such as moving Core to `agent-pool`, makes it offensive with no override, and saving an account again can move it to another sleeve. Nothing records either change. That is harmless while no agent exists, but it will matter once fiefs are leased from offensive sleeves.

## Not built yet

- **D-52**: storing the model portfolio (off by default), entry in the built-in UI, and its place in the read API.
- **D-53, D-54**: the model's line in the value-over-time chart (D-30) and its drawdowns in the UI.
- **D-62**: the reporting currency setting and its default from the jurisdiction pack. The backtest also needs FX history into it, which starts no earlier than NBP's first rate on 2002-01-02.
- **D-30, D-6**: sleeves and kinds in the read API, and the allocation chart.
- Nothing calls `EnsureDefaults`, because no core binary exists yet, so no install starts with the default sleeves. Only `cmd/demo-seed` saves sleeves, for the demo dataset. No API or UI adds kinds, renames them or sets the override.
