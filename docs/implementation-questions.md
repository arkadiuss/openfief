---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## XTB

81. **Where a closed CFD's result comes from.**
   - *Problem:* D-24 and D-152 book a closed CFD's result as cash income so the account's cash reconciles. The export R-24 read held no CFD, so it isn't known whether `Closed Positions`' `Profit/Loss` is net of the position's commission, swap and rollover, nor whether `Cash Operations` repeats the result, or those charges, in rows of its own.
   - *Today:* each closed position of a category other than a stock or ETF books its `Profit/Loss`, in the account's currency, as cash income on the UTC date it closed, keyed `closed-position:` plus its position ID. `Cash Operations` rows of such a category stay unsupported ingest problems, so nothing is booked twice ([Ingest](concepts/ingest.md#mapping-an-xstation-export)).
   - *Options:* keep; or book `Gross Profit` with commission, swap and rollover as fees of their own; or book the result from `Cash Operations` once a real CFD export shows its rows.

## Ghostfolio

80. **Cash income in Ghostfolio.**
   - *Problem:* D-152's cash income has no holding, and Ghostfolio's `INTEREST` activity needs an asset. D-108 names a manual asset only for fees.
   - *Today:* cash income has no activity: it moves only the owned account's cash balance, which follows OpenFief's cash (D-108), so Ghostfolio's totals include it but its activities don't list it ([Ghostfolio](concepts/ghostfolio.md)).
   - *Options:* keep; or book it as an `INTEREST` against a manual "OpenFief cash income" asset per currency, derived like the fee asset (D-108, D-134).
