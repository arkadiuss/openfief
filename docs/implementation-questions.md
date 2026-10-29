---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Blocking real use

3. **XTB import needs the export's real layout.**
   - *Problem:* D-9 says XTB data arrives as uploaded xStation export files, but the export's columns, and whether it separates CFDs from stocks, are unverified (OQ-22).
   - *Today:* no XTB parser exists.
   - *Needed:* a header row and two or three rows with made-up values, copied from a real export. The parser's fixtures will be synthetic.
   - *Known so far:* the author holds stocks and ETFs only at XTB (so D-24's CFD handling stays as written) and will send the headers from a real export. The export's language is not known yet. Public parsers disagree on the sheet names (R-19).

## Market data and connectors

20. **Who names what a market-data connector prices (D-22).**
    - *Today:* the owner lists ISINs and indices in the `eodhd` instance's config.
    - *Options:* the core sends the held instruments and the model portfolio's indices in the sync request (a contract change); or keep the config list.

21. **Shared market data.**
    - *Problem:* any connector instance can write prices, FX rates and index readings, each stamped with its source, and a later write replaces an earlier one.
    - *Options:* restrict each kind of data to some instances (for example only an FX connector writes FX rates); or rank sources; or keep the last write.

22. **EODHD details.**
    - *Today:* a market index is identified by the owner's name for it (for example "WIG20"). Euronext means EODHD's PA, AS, BR, LS, IR, MI and OL markets. An ISIN listed on several supported exchanges is reported as ambiguous until the owner sets `exchange`. A series that can't be priced is reported again on every sync.
    - *Question:* keep these?

23. **Real or synthetic FX rates on the demo.**
    - *Today:* the demo runs the `nbp` instance of `nbp-fx`, which writes real NBP rates next to the demo dataset's own rates.
    - *Options:* keep the plugin on the demo; or let the demo dataset own FX.

24. **The first sync of an account seeded from opening lots.**
    - *Today:* it fails until `core connector sync --mode transactions` is run once.
    - *Options:* the core detects an already-seeded account and switches to transactions mode; or keep the manual step.

## Treasury bonds

25. **Series terms are entered with each lot.**
    - *Problem:* D-9 lists series, quantity and purchase date, but valuing a lot also needs the series' first-period rate, margin and early-redemption fee.
    - *Today:* the owner enters them with each lot, and the latest entry for a series wins.
    - *Options:* keep; or fetch a catalogue of series terms.

26. **A bond lot is a deposit and a buy, not an opening lot.**
    - *Problem:* the ledger refuses to seed an account twice, so lots entered one at a time can't all be opening lots.
    - *Today:* each lot is a deposit of its cost plus a buy on its purchase date. An old lot therefore moves go-live, and the start of the time-weighted return, back to its date.
    - *Question:* acceptable?

27. **The bond's ISIN is derived from its series code.**
    - *Today:* `PL` + series code + `00` + a check digit (for example `PLEDO0536000`), not the ISIN the registry assigns, which owners don't see. A future import of the bond service's own export must use the same derivation, or a series would be held twice.
    - *Question:* acceptable?

28. **Bond lot limits.**
    - *Today:* only cash purchases at 100 PLN (not exchanges at 99.90 PLN, whose discount is taxed). No way to delete a lot; entering the same series and day replaces it. "The right connector" (D-49) is the one instance bound to the account. The UI shows the form only for accounts of institution `treasury-bonds`. Entry works only in a cluster with connectors on.
    - *Question:* keep these?

29. **COI coupons are booked as dividends.**
    - *Problem:* the ledger has no interest transaction kind.
    - *Options:* add an interest kind; or keep dividends.

30. **Who picks the NBP reference rate in force.**
    - *Today:* the treasury-bonds connector applies the rule itself: the rate in force on the tenth business day before a month is that month's reading, dated on the month's first day. The comment on `IndexReading.date` in the connector contract now says so; the wire format is unchanged.
    - *Options:* keep; or the connector sends NBP's raw rate changes and the core picks.

31. **The Polish business-day calendar.**
    - *Today:* Saturdays, Sundays and the holidays in art. 1 of the act on days off work, year by year (6 January from 2011, 24 December from 2025, Easter-based days). Epidemic days off (art. 1a) are not modelled. Dates before 1991 are refused.
    - *Question:* enough for M0?

32. **GUS CPI defaults.**
    - *Today:* only final monthly values are used; preliminary, estimated and flash values are skipped and arrive later. The first sync starts January 13 years back (`start` can change it), fetching at most 4 years per sync to stay within GUS's anonymous limits. GUS's indicator 1832 is used up to 2025 and 2728 from 2026, a boundary written into the code. A reference-rate month is delivered once its day has passed and is not fetched again.
    - *Question:* keep these?

## Performance and the model portfolio

33. **Transfers between the owner's own accounts (D-61).**
    - *Problem:* the ledger doesn't record which withdrawal became which deposit, but the time-weighted return must not count such transfers as cash flows.
    - *Today:* a withdrawal matches a deposit into another account of the same amount and currency at most 5 calendar days apart, the nearest unmatched deposit first. The performance response lists every match.
    - *Options:* keep the rule (and its window); or record transfers explicitly in the ledger.

34. **Time-weighted return conventions.**
    - *Today:* cash flows count at the end of their day; each daily factor and the running product are rounded half-even to 12 decimal places; opening lots count as inflows at their market value on their date.
    - *Question:* keep these?

35. **What the value-over-time chart draws.**
    - *Today:* the chart draws the time-weighted return index from 100 at go-live (D-30 and D-53 read as one chart with two lines from 100), with the value including cash shown alongside. `ListValuationTotals` still sums holdings only, without cash.
    - *Options:* keep; or include cash in `ListValuationTotals` too.

36. **Model portfolio details (D-52 to D-54).**
    - *Today:* the model rebalances at the end of each 31 December. A target's own drawdown is in the reporting currency (PLN) over the span where it has both closes and FX rates; P-48, not decided, would use its own currency over its whole history. Weights go in 0.01% steps, at most 50 targets. While the model is off, the read API answers 404.
    - *Question:* keep these?

## Ghostfolio

37. **Where the Ghostfolio view departs from D-64.**
    - *Today:* the plugin derives its owned accounts' IDs (UUIDv5 of instance and account) instead of generating and storing them, because a plugin has no volume and no Kubernetes API. A stale activity that Ghostfolio's duplicate check would skip is deleted before the import, or the loop never converges. The plugin runs its own timer, since D-49 has the core schedule only connectors.
    - *Question:* acceptable?

38. **How transactions become Ghostfolio activities.**
    - *Today:* a dividend is quantity 1 at the gross amount, with the withheld tax as its fee. Standalone fees go against a manual "OpenFief fees" asset per currency. Cash movements produce no activity, so Ghostfolio shows holdings without cash. Listed instruments need the owner's mapping from ISIN to Ghostfolio's data source and symbol. Bond assets are keyed by the series code only, so two users of one Ghostfolio would collide.
    - *Question:* keep these?
