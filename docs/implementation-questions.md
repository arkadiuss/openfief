---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Charts

56. **Whether the charts' start the owner chose is kept.**
   - *Problem:* D-138 lets the owner choose the date every time chart starts on, but doesn't say whether the choice is remembered. D-113 calls the number format a setting; D-138 doesn't use that word.
   - *Today:* the choice is a view parameter, `start` in the page's URL, shared by the page's two time charts: a reload or a bookmark keeps it, another browser or a link without it shows the default, and *Back to the default* drops it. The read API takes it as `from` and names the default as `default_from` ([Built-in UI](concepts/ui.md#pages)).
   - *Options:* keep; or store it as a setting next to the reporting currency and the number format, one for the portfolio and one per account; or store only the portfolio's.

57. **What "opened later" means for an account seeded from a starting balance.**
   - *Problem:* D-138's portfolio default is the latest day before which some account's history is unknown because it was seeded from a starting balance, and accounts "opened later" don't move it. An account the owner adds months later seeded from its current positions (an IBKR or XTB opening statement) also has history before its seeding day that the record doesn't know.
   - *Today:* every account seeded from a starting balance (opening lots or opening cash, D-23) moves the default to its seeding day if that is later; only accounts with their whole history, such as bond accounts booked from their purchases (D-98) or an XTB account started from its whole history (D-136), never move it, however late they opened (`views.ChartStart`).
   - *Options:* keep; or let an account seeded more than some time after the others not move it, and mark its seeding day on the chart instead; or let the owner mark an account as opened on its seeding day.
