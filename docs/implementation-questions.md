---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## XTB

33. **Holding assertions from `Open Positions`.**
   - *Problem:* D-18 stores a broker's position snapshot as balance assertions. An XTB position names its ticker only, so an assertion needs its ticker's confirmed ISIN, and the sheet's `Volume` format wasn't seen.
   - *Today:* none are made from `Open Positions`; only its CFD rows are read, to flag them.
   - *Options:* assert the summed volume per ticker on the report's date once the ticker is confirmed, leaving out unconfirmed ones; or wait for an export with open positions.

## Accounts

38. **What "no transactions from it" means for a broker account's remap.**
   - *Problem:* D-120 lets the owner remap a broker account whose OpenFief account has no transactions from it. The ledger records which account and source each transaction has, not which broker account it came from.
   - *Today:* every transaction a broker account brings is scoped to its connector instance's key source (D-48), so a remap is refused while the mapped account holds any version of a transaction whose source starts with `plugin:<instance>:`; balance assertions don't count. The instance stays bound to the old account and is bound to the new one; the broker account is no longer seeded, so the next opening batch seeds the new account.
   - *Options:* keep; or record the broker account on each transaction and check exactly; or refuse while the account holds any transaction at all.

39. **How the read API drops broker account numbers.**
   - *Problem:* D-121 has the read API never serve broker account numbers. `ListBrokerAccounts` was part of the read API, and `buf breaking` runs on every change (docs/dev/pipeline.md), so removing the method or its fields fails the check; no rule says how a read API method is retired.
   - *Today:* the method and its messages stay in the contract, marked deprecated, and the core answers it with `UNIMPLEMENTED` (HTTP 501 on `GET /v1/broker-accounts`); the owner's `AccountService.ListBrokerAccounts` lists them, and the built-in UI uses it.
   - *Options:* keep until a versioned break of the read API (`/v2`); or remove them now with an exception to the breaking-change check.

