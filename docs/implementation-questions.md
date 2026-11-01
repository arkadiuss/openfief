---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## XTB

52. **The cost of an open position quoted in another currency.**
   - *Problem:* D-136 seeds one opening lot per open position at its open time and price, but `Open Positions` gives the open price in the instrument's currency, names no currency, and gives no conversion rate (R-24), so a position such as a London ETF in a PLN account has a price whose currency the export doesn't say.
   - *Today:* only a position quoted in the account's currency is seeded: one whose `Value`, in the account's currency, is its volume times its current price to within a minor unit, at a cost of volume times open price in the account's currency. Any other is an unsupported ingest problem, so the account is seeded without it and a later sale of it can't be booked ([Ingest](concepts/ingest.md#mapping-an-xstation-export)).
   - *Options:* keep; or take the instrument's currency from the confirmed ticker's listing and book the lot's cost in that currency; or derive the currency from the market suffix (a London listing can be quoted in USD); or have the owner enter each such position's currency, or its cost in the account's currency, with the upload.

53. **An opening upload that books nothing yet.**
   - *Problem:* an account seeded from open positions goes live with its first seeding transaction (D-23). When the owner enters no cash and every position waits for its ticker's ISIN (D-116), the opening upload books nothing, so the account isn't live yet, and an upload before the owner confirms a ticker is chosen as opening again, or, overridden to transactions, books history that the opening lots then hold too.
   - *Today:* the waiting opening lots are held with their go-live date and join the seeding when their ticker is confirmed; until one is, the account is not seeded, and a confirmation after transactions dated on or before the go-live were booked is refused (`ErrBookedBeforeGoLive`). The UI asks for the cash at the start; entering it, even as 0.00, books nothing for zero.
   - *Options:* keep; or record the go-live of a pending opening in the ingest schema so the account counts as started at once; or seed a waiting position as soon as the opening upload, with a placeholder instrument replaced when the ticker is confirmed.

