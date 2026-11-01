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

## Treasury bonds

54. **Moving bond lots entered as opening lots to their purchase dates.**
   - *Problem:* D-98 (rewritten) books every hand-entered lot on its purchase date as a deposit and a buy, and an account no longer goes live on its first entry. Accounts whose lots were entered before hold them as opening lots dated on the day the first was entered, keyed `lot:<series>:<purchase date>`; booking them again on their purchase dates would hold them twice, and the connector can't tell which lots were booked which way.
   - *Today:* such an account keeps its go-live: the core sends it with each entry, a lot bought on or before it is still mapped to its old opening lot and key, and its coupons and maturities count from go-live. An account never seeded gets no go-live, so all its lots are deposits and buys and everything since purchase is booked. The owner can move an old account by deleting every opening lot (D-100), which leaves it unseeded, and entering the lots again. Nothing migrates on its own.
   - *Options:* keep; or have the core migrate such an account once, delivering a deletion of each opening lot with its deposit and buy in one batch; or refuse entries on a seeded bond account until the owner re-enters it.

55. **When a change of the payout setting takes effect.**
   - *Problem:* D-135 gives each hand-entered bond account a payout setting, but doesn't say what a change does to coupons and maturities that fell due before it. The connector books everything due on every sync, so a setting applied to the past would add, or stop delivering, withdrawals for payouts already booked.
   - *Today:* a change takes effect on the day it is made (`PayoutSetting.Since`): a coupon or redemption follows the setting in force on the day it fell due, so what was booked is never rewritten. A setting chosen when the account is created applies from that day.
   - *Options:* keep; or let the owner give the day the registry option changed, which may be in the past; or apply the current setting to all payouts.

## Charts

56. **Whether the charts' start the owner chose is kept.**
   - *Problem:* D-138 lets the owner choose the date every time chart starts on, but doesn't say whether the choice is remembered. D-113 calls the number format a setting; D-138 doesn't use that word.
   - *Today:* the choice is a view parameter, `start` in the page's URL, shared by the page's two time charts: a reload or a bookmark keeps it, another browser or a link without it shows the default, and *Back to the default* drops it. The read API takes it as `from` and names the default as `default_from` ([Built-in UI](concepts/ui.md#pages)).
   - *Options:* keep; or store it as a setting next to the reporting currency and the number format, one for the portfolio and one per account; or store only the portfolio's.

57. **What "opened later" means for an account seeded from a starting balance.**
   - *Problem:* D-138's portfolio default is the latest day before which some account's history is unknown because it was seeded from a starting balance, and accounts "opened later" don't move it. An account the owner adds months later seeded from its current positions (an IBKR or XTB opening statement) also has history before its seeding day that the record doesn't know.
   - *Today:* every account seeded from a starting balance (opening lots or opening cash, D-23) moves the default to its seeding day if that is later; only accounts with their whole history, such as bond accounts booked from their purchases (D-98) or an XTB account started from its whole history (D-136), never move it, however late they opened (`views.ChartStart`).
   - *Options:* keep; or let an account seeded more than some time after the others not move it, and mark its seeding day on the chart instead; or let the owner mark an account as opened on its seeding day.
