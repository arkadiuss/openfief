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

## Releases

69. **Whether a release ships the database.**
   - *Problem:* the install file runs the core against a CloudNativePG cluster `openfief-db` in `openfief` and reads its connection string from the Secret CloudNativePG makes for it (D-14), but docs/dev/sdlc.md names only the CRDs and the bootstrap install as release assets, and a `Cluster` manifest carries choices that depend on the owner's cluster: the storage class and size, the instance count, and the backups D-29 asks to be encrypted.
   - *Today:* the bundle ships neither CloudNativePG nor the cluster. [Install a release](guides/install.md#1-install-cloudnativepg-and-the-database) has the owner install CloudNativePG at the version the end-to-end tests run and apply a minimal one-instance `openfief-db` on the default storage class by hand; the end-to-end bundle run applies the demo's `deploy/demo/database` instead.
   - *Options:* keep the manual step; or ship a minimal `openfief-database-<version>.yaml` next to the bundle, the guide's cluster, for the owner to edit or replace; or ship it and CloudNativePG's release manifest pinned by sha256 as assets, and have the end-to-end run install from those.
