---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Blocking real use

1. **XTB import needs the export's real layout.**
   - *Problem:* D-9 says XTB data arrives as uploaded xStation export files, but the export's columns, and whether it separates CFDs from stocks, are unverified (OQ-22).
   - *Today:* no XTB parser exists.
   - *Needed:* a header row and two or three rows with made-up values, copied from a real export. The parser's fixtures will be synthetic.

## Write paths for the owner

2. **Statement uploads (D-7, D-49).**
   - *Today:* an upload goes to the one connector instance bound to the account (none or several is refused), and its delivery must stay within that account, so a Flex statement covering several accounts imports only the uploaded account's part. The owner picks the mode (opening or transactions) on each upload. The form shows on every account page, because the read API doesn't say which connector is bound; a connector that takes no files refuses with its reason. An opening upload doesn't mark the instance as seeded for its scheduled syncs. The file name and the core's sniffed media type are hints only. The core caps a file at 64 MiB, the largest any connector accepts today.
   - *Question:* keep these?

## Identity and the UI

3. **How the built-in UI formats amounts.**
   - *Problem:* the design system says amounts are formatted for the reader's locale, Polish being `5 000,00 PLN`, but no decision says which locale the UI uses: its pages are in English, and OpenFief is built for Polish investors first.
   - *Today:* every amount keeps its exact digits and its currency, with thousands separated by a narrow no-break space and a decimal point, such as `5 000.00 PLN`, whatever the browser's language. The charts' axes use Chart.js's own number format.
   - *Options:* keep one format; follow the browser's language (`Accept-Language`); or a setting next to the reporting currency.

## Market data and connectors

4. **Who names what a market-data connector prices (D-22).**
    - *Today:* the owner lists ISINs and indices in the `eodhd` instance's config.
    - *Options:* the core sends the held instruments and the model portfolio's indices in the sync request (a contract change); or keep the config list.

5. **Shared market data.**
    - *Problem:* any connector instance can write prices, FX rates and index readings, each stamped with its source, and a later write replaces an earlier one.
    - *Options:* restrict each kind of data to some instances (for example only an FX connector writes FX rates); or rank sources; or keep the last write.

6. **EODHD details.**
    - *Today:* a market index is identified by the owner's name for it (for example "WIG20"). Euronext means EODHD's PA, AS, BR, LS, IR, MI and OL markets. An ISIN listed on several supported exchanges is reported as ambiguous until the owner sets `exchange`. A series that can't be priced is reported again on every sync.
    - *Question:* keep these?

7. **Real or synthetic FX rates on the demo.**
    - *Today:* the demo runs the `nbp` instance of `nbp-fx`, which writes real NBP rates next to the demo dataset's own rates.
    - *Options:* keep the plugin on the demo; or let the demo dataset own FX.

8. **The first sync of an account seeded from opening lots.**
    - *Today:* it fails until `core connector sync --mode transactions` is run once.
    - *Options:* the core detects an already-seeded account and switches to transactions mode; or keep the manual step.

9. **The price check threshold (D-22).**
    - *Today:* a price check is flagged when IBKR's price differs from the stored price by more than 1% of it, in either direction; the core's `--price-check-threshold` changes it (a fraction greater than 0 and less than 1). One percent was picked as well above the few basis points that separate an aggregated close from an exchange's official close on a liquid listing (D-51), and well below the errors the cross-check exists to catch: a wrong listing or currency, pence read as pounds, a missed split, a stale day after a big move. A thinly traded Warsaw listing may flag more often.
    - *Options:* keep 1%; another value; or a threshold per exchange or per instrument.

10. **What is compared with what (D-22).**
    - *Today:* the IBKR connector reads only each open position's `markPrice`, from the SUMMARY row or else from LOT rows that agree, dated by `reportDate`, and only for positions held. A cross-check price is compared only with the stored price of the same instrument (ISIN and symbol), on the same date, in the same currency. One in another currency, such as an ETF priced from its Xetra listing in EUR while IBKR reports the holding in GBP, is not compared and not shown, and a day without a stored price gives no check. A check is relative to the stored price.
    - *Options:* keep; also show pairs in different currencies, as their own status, since they may mean the market-data connector prices another listing than the one held; compare with the latest price on or before the day; read other Flex prices, such as `closePrice` of prior-period positions.

11. **Who sends cross-check prices, and how long checks stay.**
    - *Today:* any connector instance may send `cross_check_prices`, not only IBKR's; each is stored under its instance, and nothing is ever pruned. `ListPriceChecks` returns every check of every day, and a flagged check stays flagged: there is no way to mark one as seen. The UI lists every flagged one, of any day, on `/price-checks`. The in-core mapping that loads the demo dataset drops cross-check prices, so the demo shows none.
    - *Options:* keep; only connectors that declare it may send them; a date range or the latest check per instrument; an acknowledgement the owner records; synthetic cross-check prices in the demo dataset.

## Treasury bonds

12. **Series terms are entered with each lot.**
    - *Problem:* D-9 lists series, quantity and purchase date, but valuing a lot also needs the series' first-period rate, margin and early-redemption fee.
    - *Today:* the owner enters them with each lot, and the latest entry for a series wins.
    - *Options:* keep; or fetch a catalogue of series terms.

13. **A bond lot is a deposit and a buy, not an opening lot.**
    - *Problem:* the ledger refuses to seed an account twice, so lots entered one at a time can't all be opening lots.
    - *Today:* each lot is a deposit of its cost plus a buy on its purchase date. An old lot therefore moves go-live, and the start of the time-weighted return, back to its date.
    - *Question:* acceptable?

14. **The bond's ISIN is derived from its series code.**
    - *Today:* `PL` + series code + `00` + a check digit (for example `PLEDO0536000`), not the ISIN the registry assigns, which owners don't see. A future import of the bond service's own export must use the same derivation, or a series would be held twice.
    - *Question:* acceptable?

15. **Bond lot limits.**
    - *Today:* only cash purchases at 100 PLN (not exchanges at 99.90 PLN, whose discount is taxed). No way to delete a lot; entering the same series and day replaces it. "The right connector" (D-49) is the one instance bound to the account. The UI shows the form only for accounts of institution `treasury-bonds`. Entry works only in a cluster with connectors on.
    - *Question:* keep these?

16. **COI coupons are booked as dividends.**
    - *Problem:* the ledger has no interest transaction kind.
    - *Options:* add an interest kind; or keep dividends.

17. **Who picks the NBP reference rate in force.**
    - *Today:* the treasury-bonds connector applies the rule itself: the rate in force on the tenth business day before a month is that month's reading, dated on the month's first day. The comment on `IndexReading.date` in the connector contract now says so; the wire format is unchanged.
    - *Options:* keep; or the connector sends NBP's raw rate changes and the core picks.

18. **The Polish business-day calendar.**
    - *Today:* Saturdays, Sundays and the holidays in art. 1 of the act on days off work, year by year (6 January from 2011, 24 December from 2025, Easter-based days). Epidemic days off (art. 1a) are not modelled. Dates before 1991 are refused.
    - *Question:* enough for M0?

19. **GUS CPI defaults.**
    - *Today:* only final monthly values are used; preliminary, estimated and flash values are skipped and arrive later. The first sync starts January 13 years back (`start` can change it), fetching at most 4 years per sync to stay within GUS's anonymous limits. GUS's indicator 1832 is used up to 2025 and 2728 from 2026, a boundary written into the code. A reference-rate month is delivered once its day has passed and is not fetched again.
    - *Question:* keep these?

## Performance and the model portfolio

20. **Transfers between the owner's own accounts (D-61).**
    - *Problem:* the ledger doesn't record which withdrawal became which deposit, but the time-weighted return must not count such transfers as cash flows.
    - *Today:* a withdrawal matches a deposit into another account of the same amount and currency at most 5 calendar days apart, the nearest unmatched deposit first. The performance response lists every match.
    - *Options:* keep the rule (and its window); or record transfers explicitly in the ledger.

21. **Time-weighted return conventions.**
    - *Today:* cash flows count at the end of their day; each daily factor and the running product are rounded half-even to 12 decimal places; opening lots count as inflows at their market value on their date.
    - *Question:* keep these?

22. **What the value-over-time chart draws.**
    - *Today:* the chart draws the time-weighted return index from 100 at go-live (D-30 and D-53 read as one chart with two lines from 100), with the value including cash shown alongside. `ListValuationTotals` still sums holdings only, without cash.
    - *Options:* keep; or include cash in `ListValuationTotals` too.

23. **Model portfolio details (D-52 to D-54).**
    - *Today:* the model rebalances at the end of each 31 December. A target's own drawdown is in the reporting currency (PLN) over the span where it has both closes and FX rates; P-48, not decided, would use its own currency over its whole history. Weights go in 0.01% steps, at most 50 targets. While the model is off, the read API answers 404.
    - *Question:* keep these?

## Ghostfolio

24. **Where the Ghostfolio view departs from D-64.**
    - *Today:* the plugin derives its owned accounts' IDs (UUIDv5 of instance and account) instead of generating and storing them, because a plugin has no volume and no Kubernetes API. A stale activity that Ghostfolio's duplicate check would skip is deleted before the import, or the loop never converges. The plugin runs its own timer, since D-49 has the core schedule only connectors.
    - *Question:* acceptable?

25. **How transactions become Ghostfolio activities.**
    - *Today:* a dividend is quantity 1 at the gross amount, with the withheld tax as its fee. Standalone fees go against a manual "OpenFief fees" asset per currency. Cash movements produce no activity, so Ghostfolio shows holdings without cash. Listed instruments need the owner's mapping from ISIN to Ghostfolio's data source and symbol. Bond assets are keyed by the series code only, so two users of one Ghostfolio would collide.
    - *Question:* keep these?

## End-to-end tests

26. **How the end-to-end suite gets signed plugin images to the operator.**
    - *Problem:* the operator deploys only images whose signature it reads from their registry (D-56, D-57), and kind and k3d configure their registries differently (D-12).
    - *Today:* `task e2e:images` runs a `registry:3.0.0` container on the cluster's Docker network and pushes `nbp-fx` there; the suite signs it with a key made for the run, trusted through `openfief-trusted-signers` (D-45). The operator reads that registry over plain HTTP only because its address is a private IPv4 address, for which go-containerregistry allows HTTP; the nodes get the image through `ctr --plain-http` pulls, so containerd needs no registry configuration. No operator flag or setting was added for the tests.
    - *Options:* keep; or give the registry TLS with a test CA the operator and the nodes trust, closer to a real registry but with per-distribution node configuration.
