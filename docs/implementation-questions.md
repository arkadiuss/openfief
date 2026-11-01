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

## Read access

65. **Whether owner sessions hold read access.**
   - *Problem:* D-137 says script tokens, view plugins and demo sessions hold read access "and nothing else does", while D-83 has the read API accept sessions of either scope, so an owner session reads it too.
   - *Today:* the read API accepts owner and read sessions, script tokens and view plugins' tokens, and refuses connectors' (`readTokens` in `cmd/core`); the built-in UI reads with the owner's session. Read sessions come only from the demo authenticator (D-68), so a read session is a demo session.
   - *Options:* keep, reading "nothing else" as nothing other than these and the owner; or refuse owner sessions on the read API, so the UI reads with its own token and the owner's session only proves the owner.

66. **Read API methods that D-137's list doesn't name.**
   - *Problem:* D-137 says read access is everything the read API serves and lists what that is. The read API also serves the reporting currency, number format and tax wrappers (`GetSettings`), the yearly taxes (`ListYearlyTaxes`) and the balance assertion checks (`ListAssertionChecks`), which the list doesn't name; sleeve kinds and connector bindings are placed under sleeves and accounts.
   - *Today:* all of them stay; none holds a secret, a broker account number or a change of state. `TestReadAccessMethods` places every method against D-137 and fails on a new one that isn't placed.
   - *Options:* keep, reading the list as kinds of records; or take the methods outside it off the read API, as D-157 does with `ListBrokerAccounts`.

67. **What an unmapped broker account's attention item shows on the read API.**
   - *Problem:* the item's subject is the connector instance and the broker account number, and read access never includes a broker account number (D-121, D-137).
   - *Today:* the read API serves that subject as the instance alone (`attention.ReadSubject`), so two unmapped accounts of one instance share a subject and differ by item ID; the owner finds the broker account on the accounts page, through the account service. The stored subject is unchanged.
   - *Options:* keep; or give the item a subject derived from the broker account that can't be read back, such as a hash; or leave these items off the read API.

68. **Which connectors look up a ticker's ISIN.**
   - *Problem:* D-123 restricts D-93's list of held instruments to declared and approved market-data connectors; D-116's lookup of a ticker's ISIN, which also tells a connector what the owner holds, isn't named.
   - *Today:* the core asks only market-data connectors, as D-123 defines them, to look up a ticker; before D-123 it asked every instance bound to no account.
   - *Options:* keep; or ask every running connector that implements the lookup.

