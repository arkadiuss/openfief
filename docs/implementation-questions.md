---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## XTB

70. **The currency of a London listing that names none.**
   - *Problem:* D-159 takes an open position's price currency from its confirmed ticker's listing and, when the listing names none, from the exchange its market suffix names. London quotes most shares in pence (GBX), but many London ETFs trade in USD or GBP.
   - *Today:* a `.UK` ticker confirmed without a listing that names its currency is treated as quoted in pence, so its opening lot's cost is booked in pounds after dividing by 100. A London ETF quoted in USD or GBP comes out wrong in that case; one whose listing names its currency is booked correctly ([Ingest](concepts/ingest.md)).
   - *Options:* keep; or leave such a position waiting as an ingest problem until the owner names the currency; or ask the owner for the currency when confirming any `.UK` ticker whose listing names none.
