---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Built-in UI on a phone

27. **What the installed app shows offline.**
   - *Problem:* the author asked for a mobile PWA. D-69 makes the UI a backend-for-frontend whose pages hold the owner's record and are never cached; no decision says whether the app may keep any of the record on the device for offline reading.
   - *Today:* the service worker caches only static files (styles, scripts, fonts, icons, the manifest) and a static offline page with Kasztan asleep; every page comes from the network, the cache is emptied on sign-out, and nothing of the record is stored on the device ([Built-in UI](concepts/ui.md#on-a-phone-and-as-an-app)).
   - *Options:* keep; or keep the last overview (totals and accounts, no transactions) for offline reading, encrypted or not, cleared on sign-out, which puts portfolio data on the phone; or push notifications, which need a push service outside the owner's cluster.

## XTB

32. **How closed CFD results and free-funds interest are booked.**
   - *Problem:* D-24 imports closed CFD results as cash profit or loss, but the ledger has no kind for a trading result or for interest on cash (`NewInterest` needs a holding), and R-24's export had no CFD, so which `Cash Operations` type carries a CFD's result, and whether `Closed Positions`' `Profit/Loss` repeats it, wasn't seen. `Free funds interest` and its tax are in R-24.
   - *Today:* each closed CFD position is an unsupported ingest problem naming its category, as is each free-funds interest row and its tax; cash is short by their sum.
   - *Options:* a ledger kind for a cash result, posted against an income ledger account so performance doesn't count it as a deposit, for CFD results and interest on cash alike; or book losses as fees and profits as deposits, which misstates performance.

33. **Holding assertions from `Open Positions`.**
   - *Problem:* D-18 stores a broker's position snapshot as balance assertions. An XTB position names its ticker only, so an assertion needs its ticker's confirmed ISIN, and the sheet's `Volume` format wasn't seen.
   - *Today:* none are made from `Open Positions`; only its CFD rows are read, to flag them.
   - *Options:* assert the summed volume per ticker on the report's date once the ticker is confirmed, leaving out unconfirmed ones; or wait for an export with open positions.

34. **What a ticker confirmation covers, and its US listings.**
   - *Problem:* D-116 has the owner confirm a ticker's ISIN once. A confirmation could hold for one connector instance or for every instance of the institution, and a wrong confirmation can't be changed. EODHD's supported exchanges (D-51) leave out US markets, which XTB lists with `.US`.
   - *Today:* a confirmation holds for the instance that reported the ticker, is set once, and a different ISIN later is refused; a ticker of a market with no supported exchange, `.US` among them, gets no proposal and the owner types its ISIN.
   - *Options:* keep; or share confirmations across instances of one institution; or let the owner correct a confirmation, superseding the transactions it booked (D-18).

## Treasury bonds

36. **A coupon booked on a lot the owner later deletes.**
   - *Problem:* D-100 lets the owner delete a lot as a correction, which supersedes its transactions. Coupons the connector booked on it before are separate transactions, keyed by the lot and coupon date, and the connector can't tell which were booked: deleting a key never booked is refused (D-18).
   - *Today:* a deleted lot is no longer open, so it gets no further coupons, but those booked earlier stay.
   - *Options:* keep; or deliver, with a lot's deletion, a deletion of each coupon the core says was booked on it; or let a deletion of a key never booked be a no-op for coupons only.

## Accounts

38. **What "no transactions from it" means for a broker account's remap.**
   - *Problem:* D-120 lets the owner remap a broker account whose OpenFief account has no transactions from it. The ledger records which account and source each transaction has, not which broker account it came from.
   - *Today:* every transaction a broker account brings is scoped to its connector instance's key source (D-48), so a remap is refused while the mapped account holds any version of a transaction whose source starts with `plugin:<instance>:`; balance assertions don't count. The instance stays bound to the old account and is bound to the new one; the broker account is no longer seeded, so the next opening batch seeds the new account.
   - *Options:* keep; or record the broker account on each transaction and check exactly; or refuse while the account holds any transaction at all.

39. **How the read API drops broker account numbers.**
   - *Problem:* D-121 has the read API never serve broker account numbers. `ListBrokerAccounts` was part of the read API, and `buf breaking` runs on every change (docs/dev/pipeline.md), so removing the method or its fields fails the check; no rule says how a read API method is retired.
   - *Today:* the method and its messages stay in the contract, marked deprecated, and the core answers it with `UNIMPLEMENTED` (HTTP 501 on `GET /v1/broker-accounts`); the owner's `AccountService.ListBrokerAccounts` lists them, and the built-in UI uses it.
   - *Options:* keep until a versioned break of the read API (`/v2`); or remove them now with an exception to the breaking-change check.

