---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Plugins

3. **How a release's catalog reaches `main`, and where a beta's CLI goes.**
   - *Problem:* D-88 has the release workflow add each first-party plugin to the catalog, a file in this repository embedded in the operator and CLI, and docs/dev/sdlc.md has every release ship the signed CLI. The release starts from `main`, which is never committed to directly, and a beta so far had no GitHub release.
   - *Today:* the plugin images are built and signed first; the workflow commits the catalog to a new branch `release/<version>`, builds the other images and the CLI from that commit, tags it, and opens a pull request to `main`. A beta becomes a GitHub pre-release carrying only the signed CLI.
   - *Options:* keep; push the catalog commit to `main` from the workflow; or keep betas without a GitHub release and ship their CLI as a workflow artifact only.

## Market data

7. **Which connector instances are asked what to price.**
   - *Problem:* D-93 has each sync request to a market-data connector carry the instruments the owner holds and the model portfolio's targets. No manifest field says that a connector is a market-data connector, and the list tells a connector what the owner holds at every institution.
   - *Today:* the core sends the list to every connector instance bound to no account, since such an instance can deliver only market data; a broker connector, which is bound to the accounts it feeds, gets it empty. A market-data connector the owner bound to an account would get nothing. Held means a non-zero holding today, treasury bonds with series terms left out; the model's targets count only while it is on (D-52).
   - *Options:* keep; or a manifest field declaring a market-data connector, approved with the grant like the cross-check prices of D-112; or send it to every connector.

8. **What a market-data source is, and what a ranking covers.**
   - *Problem:* D-94 has the owner rank "the sources" of prices, exchange rates and index readings, with the first-party connector of each first by default, and keep the lower-ranked values as cross-checks (D-22). It doesn't say whether a source is a connector instance or a plugin, how the core knows which instance is the first-party one, whether D-22's cross-check prices are a ranked source, or whether index levels are ranked.
   - *Today:* a source is the connector instance that delivered the value (D-48). Its default place comes from the institution it names on its deliveries: `eodhd` ranks first for prices, `nbp` for exchange rates and `treasury-bonds` for index readings, the rest by name. Cross-check prices are stored apart and never win (D-112 gates who may send them); lower-ranked prices are checked against the winner like cross-check prices. Index levels of market indices keep one value per index and date, the last delivered. A change of ranking revalues every account from its first transaction.
   - *Options:* keep; or rank by plugin image rather than instance, or by a manifest field naming the kinds a connector serves; or rank IBKR's mark prices as one more price source, so they fill days the market-data connector lacks; or rank index levels too.

## Exposure

9. **Which of two views keeps the browser path.**
   - *Problem:* D-114 has the operator not run a plugin with a browser path while another instance holds one. Two such plugins applied at once, as GitOps may, can both start before either's status records a path, and then each sees the other holding one.
   - *Today:* a plugin holds a browser path when its status records a deployed version with an effective browser path. One that holds none yet yields to any holder; of two that both hold one, the older Plugin (by creation time, then by name) keeps it and the other is torn down with `Ready=False/BrowserPathHeld`. The core serves no view while it reads more than one.
   - *Options:* keep; or let the newer one keep it; or refuse both until the owner deletes one.

10. **How `openfief enroll --origin` knows the origins the core allows.**
   - *Problem:* D-80 requires the origin to be one the core allows, but `core enroll-link` runs as its own process in the core's pod, without `serve`'s flags.
   - *Today:* the CLI reads the `--auth-origin` flags from the running core pod's container spec (none means the core's default, `http://localhost:8080`) and refuses any other origin before it runs `core enroll-link`.
   - *Options:* keep; or move the allowed origins into a ConfigMap or environment variable that both `serve` and `enroll-link` read, so the core itself refuses.

## Ghostfolio view

11. **Cash in a currency other than the owned account's.**
   - *Problem:* D-108 has each owned account's cash balance follow OpenFief's cash, so Ghostfolio's totals include it. A Ghostfolio account has one balance in one currency, and the view creates its accounts in the Ghostfolio user's base currency; an account such as IBKR's holds cash in several.
   - *Today:* the balance history follows the account's cash in the owned account's currency, one balance at the end of each date it changes. Cash in any other currency is left out and counted among the sync's problems, so Ghostfolio's totals miss it.
   - *Options:* keep; or give each other currency an owned cash account of its own, its UUID derived from the instance, the account and the currency, holding only a balance; or convert the cash at OpenFief's exchange rates, which the read API would have to serve; or create each owned account in its account's cash currency when it has only one.

13. **What the bond and fee assets' symbols are derived from.**
   - *Problem:* D-64 (amended) derives a bond's symbol from the plugin instance and the series code, so two installs sharing one Ghostfolio never collide. Ghostfolio's symbols are unique across all its users, and two installs can name their instances alike (an install names an instance after the plugin's catalog name unless told otherwise); then they would collide. D-108's fee asset per currency has the same problem.
   - *Today:* both symbols are UUIDs derived from the instance's name, the Ghostfolio user's ID and the series code or currency. Two instances writing to the same Ghostfolio user still meet in Ghostfolio's name-based reuse of `MANUAL` assets, which the dry run catches.
   - *Options:* keep; or the instance's name alone, as D-64 says; or a per-install ID, which a plugin can't read today.

## Values

14. **OKI's act: its in-force date, limit and rate.**
   - *Problem:* D-76 computes OKI's yearly tax "only once the act is in force, with the limit and rate taken from the act". The code knows no act in force, nor its numbers, and doesn't know how a year the act enters mid-way is counted, or whether the average balance is taken over every calendar day or only the days with a valuation.
   - *Today:* `pl.OKIAct` holds the in-force date, the limit and the rate as data, and `pl.CurrentOKIAct` is not in force (no date), so no tax is ever computed and the read API says so. Under an act, the tax would be the rate of the part above the limit of the average of the daily balances (gross values with cash, summed over every OKI account) of the days of the year on or after the in-force date that are valued, rounded half up to grosze.
   - *Needed:* the act's in-force date, limit and rate, and how it defines the average balance; then `CurrentOKIAct` gets them.

15. **How the taxes inside the values are counted.**
   - *Problem:* D-77 has the pack compute net and liquidation values "after the taxes due then", per holding. Polish tax counts gains per tax year over all disposals (PIT-38), rounds the base and the tax to whole złoty, and taxes a return from IKE on what is withdrawn above what was paid in.
   - *Today:* each holding is taxed on its own gain over its open lots' cost, at 19%, rounded half up to grosze (as D-37 rounds bond interest), and a loss is untaxed but offsets nothing. An IKE's early return is taxed the same way per holding, so dividends and realized gains kept as cash in the IKE are not taxed in its liquidation value. IKZE's 10% exit tax and the owner's marginal rate before 65 apply to every holding and to cash, each rounded half up to grosze.
   - *Options:* keep; or compute each account's taxes over its total (losses offsetting gains, IKE on payments in), which needs the account's contributions; or round as PIT-38 does.

16. **What a sale's cost covers.**
   - *Problem:* D-78's fee rule is "a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency". It doesn't say which currency the minimum is in, whether bonds, whose redemption costs no commission, pay it, or what foreign is relative to.
   - *Today:* the minimum is an amount in any currency, converted at the FX rate of the valuation's day (a missing rate leaves the liquidation value missing); the commission applies to holdings priced at market only, not to bonds or cash; and foreign means priced in another currency than the pack's, PLN for Poland. The minimum applies to the commission alone, before the conversion percentage is added.
   - *Options:* keep; or tie the minimum to the instrument's currency; or charge the conversion percentage on foreign cash too.

## Performance

17. **What happens to a transfer decision when a leg is corrected.**
   - *Problem:* D-105 stores the owner's confirmed link append-only, keyed by the two transactions. A broker can later correct either leg (D-18), say its amount or date, so that the two are no longer one possible transfer.
   - *Today:* the link stays stored, but it counts only while the legs' current versions are still a withdrawal and a deposit of the opposite amount in different accounts; otherwise both legs are cash flows again and the pair is no longer listed as confirmed. The owner's rejections are stored the same way, so a rejected pair is never suggested again, and a leg can be in only one confirmed transfer.
   - *Options:* keep; flag a broken link for the owner to look at, like a flagged balance assertion; or ask the owner again by suggesting the corrected pair.

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
