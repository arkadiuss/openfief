---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Statement uploads and price checks

23. **How the core learns which connectors take statement files.**
   - *Problem:* D-109 has the owner's UI learn through its owner service which accounts have a connector that takes files, but no decision says how the core knows a connector takes files: the connector contract's `ImportFile` only answers `UNIMPLEMENTED` once a file is sent.
   - *Today:* a connector's manifest declares `statementFiles: true`, signed with its image like `crossCheckPrices` (D-112); ibkr-flex declares both. `StatementService.ListStatementAccounts` lists the accounts bound to exactly one running instance that declares it, and the UI shows the upload form only there. Neither declaration is a permission the grant approves.
   - *Options:* keep; or ask each connector through a new contract method; or make the declarations permissions the owner approves in the grant.

24. **The mode of a statement that covers several accounts.**
   - *Problem:* D-109 has the core choose the mode, opening for an account not yet seeded and transactions after, but a statement of linked accounts can cover a seeded account and one that isn't, and an upload is one batch in one mode.
   - *Today:* the mode follows the account the owner uploaded for; the owner may override it. An opening upload leaves out the seeding of linked broker accounts already seeded, as a sync does, so it seeds only the others; a transactions upload books every mapped account's transactions, an unseeded one's too.
   - *Options:* keep; or run an opening batch and then a transactions batch, as a sync does (`PlanSync`), which needs the file twice; or refuse a statement whose accounts disagree.

25. **What a connector's undeclared cross-check prices do.**
   - *Problem:* D-112 lets only a connector whose manifest declares it send cross-check prices, without saying what happens to a delivery from one that doesn't.
   - *Today:* the batch imports everything else; none of its cross-check prices is stored, and one skipped ingest problem says so. Refusing the batch would stop a broker's syncs over prices that only cross-check.
   - *Options:* keep; or refuse the whole delivery.

26. **What an acknowledgement covers, and the owner's thresholds.**
   - *Problem:* D-112 lets the owner acknowledge a flagged price check, recorded append-only, and D-110 lets the owner set a threshold per instrument; neither says how long an acknowledgement holds, whether a matched or currency-mismatch check can be acknowledged, or what thresholds are allowed.
   - *Today:* only a flagged check can be acknowledged; the acknowledgement records both prices and covers the check only while the cross-check price and the price it is compared with keep their dates and closes, so a changed mark or a newer price is flagged anew. A threshold is a fraction greater than 0 and less than 1; clearing it puts the instrument back on 1%; thresholds are not versioned. The overview counts instruments whose latest check is flagged and not acknowledged.
   - *Options:* keep; or let an acknowledgement cover every later check of the instrument until its prices diverge further; or let the owner acknowledge a currency mismatch too.

## Built-in UI on a phone

27. **What the installed app shows offline.**
   - *Problem:* the author asked for a mobile PWA. D-69 makes the UI a backend-for-frontend whose pages hold the owner's record and are never cached; no decision says whether the app may keep any of the record on the device for offline reading.
   - *Today:* the service worker caches only static files (styles, scripts, fonts, icons, the manifest) and a static offline page with Kasztan asleep; every page comes from the network, the cache is emptied on sign-out, and nothing of the record is stored on the device ([Built-in UI](concepts/ui.md#on-a-phone-and-as-an-app)).
   - *Options:* keep; or keep the last overview (totals and accounts, no transactions) for offline reading, encrypted or not, cleared on sign-out, which puts portfolio data on the phone; or push notifications, which need a push service outside the owner's cluster.
