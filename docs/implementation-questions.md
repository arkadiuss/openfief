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
