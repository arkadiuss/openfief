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

## XTB

31. **How an xStation export seeds an account.**
   - *Problem:* D-23 seeds an account from the broker's lot-level positions and its cash at go-live, and D-109 has the core choose opening mode for an account not yet seeded. R-24's export holds no cash balance, and `Open Positions` gives each position's open price in the instrument's currency with no conversion rate, so neither the opening cash nor a foreign position's cost in the account currency is in it; that sheet's value formats were not seen either.
   - *Today:* the xtb connector refuses an opening batch (`ErrOpeningMode`), and the owner overrides the mode to transactions and uploads the account's whole history, from its first deposit.
   - *Options:* keep; or seed opening lots from `Open Positions` for positions quoted in the account currency and take the opening cash from the owner; or take the cost of every position from the owner.

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

35. **What a sync request carries besides the holdings and open lots.**
   - *Problem:* D-117 has every sync request carry the holdings and open lots of the bound accounts, so the treasury-bonds connector can book each coupon with the Polish pack's rules. A coupon's rate needs the series' terms, which the owner enters with a lot and only the core stores (D-97), its tax needs the account's tax wrapper (D-76), and what fell due before go-live is history the opening lots hold (D-98); a plugin stores nothing.
   - *Today:* each account in the request also carries its go-live date and tax wrapper, and the request carries the stored terms of the bonds among the lots (`SyncRequest.accounts`, `bond_terms`, [Ingest](concepts/ingest.md#what-a-sync-tells-a-bound-connector)). The connector books every coupon due after go-live on every sync, keyed per lot and coupon date, which covers those due since its last sync and stays idempotent; it asks GUS and NBP for the readings the coupons need.
   - *Options:* keep; or keep the request to holdings and lots and have the connector deliver the coupons gross, leaving the tax to the core; or book only coupons dated after the last sync's watermark, which misses those of a lot entered later with an earlier purchase date.

36. **A coupon booked on a lot the owner later deletes.**
   - *Problem:* D-100 lets the owner delete a lot as a correction, which supersedes its transactions. Coupons the connector booked on it before are separate transactions, keyed by the lot and coupon date, and the connector can't tell which were booked: deleting a key never booked is refused (D-18).
   - *Today:* a deleted lot is no longer open, so it gets no further coupons, but those booked earlier stay.
   - *Options:* keep; or deliver, with a lot's deletion, a deletion of each coupon the core says was booked on it; or let a deletion of a key never booked be a no-op for coupons only.

37. **The last period's interest and the redemption at maturity.**
   - *Problem:* a coupon series (ROR, DOR, COI) pays its last period's interest with the nominal at maturity, and OpenFief books no redemption yet; the valuation keeps a matured lot at its redemption value.
   - *Today:* the last period's interest is not booked as a coupon, so a matured lot is valued with it and nothing is counted twice; the redemption itself is not booked.
   - *Options:* keep until redemptions are decided; or book the redemption at maturity, the nominal and the last period's interest, as a sell of the lot.

## Accounts

38. **What "no transactions from it" means for a broker account's remap.**
   - *Problem:* D-120 lets the owner remap a broker account whose OpenFief account has no transactions from it. The ledger records which account and source each transaction has, not which broker account it came from.
   - *Today:* every transaction a broker account brings is scoped to its connector instance's key source (D-48), so a remap is refused while the mapped account holds any version of a transaction whose source starts with `plugin:<instance>:`; balance assertions don't count. The instance stays bound to the old account and is bound to the new one; the broker account is no longer seeded, so the next opening batch seeds the new account.
   - *Options:* keep; or record the broker account on each transaction and check exactly; or refuse while the account holds any transaction at all.

39. **How the read API drops broker account numbers.**
   - *Problem:* D-121 has the read API never serve broker account numbers. `ListBrokerAccounts` was part of the read API, and `buf breaking` runs on every change (docs/dev/pipeline.md), so removing the method or its fields fails the check; no rule says how a read API method is retired.
   - *Today:* the method and its messages stay in the contract, marked deprecated, and the core answers it with `UNIMPLEMENTED` (HTTP 501 on `GET /v1/broker-accounts`); the owner's `AccountService.ListBrokerAccounts` lists them, and the built-in UI uses it.
   - *Options:* keep until a versioned break of the read API (`/v2`); or remove them now with an exception to the breaking-change check.

## Taxes

40. **How OKI's averages and tax are rounded.**
   - *Problem:* D-127 and the act on OKI (R-18) don't say how an account's average value, the base or the tax on asset value is rounded. Art. 63 § 1 of the tax ordinance rounds bases and taxes to whole złoty, but the act has its own base (a sum of averages) and its own return.
   - *Today:* each account's average per exemption class is rounded half up to grosze, and so is the tax; the exemptions and the taxed part are exact sums of those averages ([Valuation](concepts/valuation.md#okis-yearly-tax)).
   - *Options:* keep; or round the base and the tax to whole złoty as art. 63 § 1 does; or keep the averages exact and round only the tax.

41. **Where OKI's later rates and exempt amounts come from.**
   - *Problem:* from 2028 the rate is 19% of the NBP reference rate in force on 31 October of the year before, and from 2030 the exempt amounts rise by the CPI of the first three quarters of the year before over the same quarters a year earlier (art. 25, 26 ust. 5). The core stores reference rate readings keyed by the bond periods they set and GUS's monthly CPI, neither of which is that figure; the minister also announces both by 30 November. Nor does the act say what an index below 100 does to amounts it "raises".
   - *Today:* `OKIAct` holds 2027's rate and the 2027 to 2029 exempt amounts; `OKIRate` and `IndexOKIExemption` compute later years from a reference rate change list and CPI factors the act data carries, both empty in `CurrentOKIAct`, so 2028's tax reports its averages and taxed part with no rate or tax, and from 2030 no exemptions either. A factor below 1 would lower the amounts.
   - *Options:* add the announced rate and amounts to the act data with each release; or store the NBP reference rate history and GUS's quarterly index as inputs and compute them; and decide whether the amounts may fall.

42. **When an OKI account counts as held, and a year in progress.**
   - *Problem:* the average divides by the days of the year the account was held (art. 23 ust. 2). OpenFief knows an account from its first transaction, which for an account seeded at go-live (D-23) is later than its opening; and a year in progress has no year-end.
   - *Today:* an account is held from its first transaction through the year's end, counted through the earliest date the OKI accounts are valued through: the current year's tax is an estimate over the days so far. Opening cash (the seeding deposit) is no payment in.
   - *Options:* keep; or let the owner enter an OKI account's opening date; or project the current year to its end.

43. **Whether PLN cash in OKI is a savings or an investment asset.**
   - *Problem:* PLN cash is a savings asset on a bank account or on the cash account of a treasury bond account, and an investment asset on any other cash account (art. 26 ust. 1 pkt 1 lit. a and e, pkt 2 lit. h). OpenFief doesn't know an account's form.
   - *Today:* PLN cash is a savings asset on an OKI account that held a treasury savings bond on any counted day of the year, and an investment asset otherwise; cash in another currency gets no exemption.
   - *Options:* keep; or let the owner set an OKI account's form (bank, bond account, brokerage, fund register) with the account.

44. **What an IKE's early return is taxed above.**
   - *Problem:* D-128 taxes an IKE's early return on what is withdrawn above what was paid in. For an account seeded at go-live the payments before it aren't in the record, and a partial return isn't modelled.
   - *Today:* the liquidation total of an IKE is a return of the whole account: its proceeds after sale costs less what was paid in, its deposits (opening cash included) less its withdrawals plus the cost of the lots it was seeded with, each at the NBP rate of the day before, taxed at 19% half up to grosze when positive.
   - *Options:* keep; or let the owner enter an IKE's total payments in at go-live.

45. **What the totals' realized gains count.**
   - *Problem:* D-128 offsets the tax year's realized gains and losses with the unrealized ones. The ledger keeps buy fees out of lot cost (D-34), but PIT-38 deducts them; foreign dividends and their withholding are PIT-38 income too.
   - *Today:* realized gains are the sells of the year through the day: the proceeds less the sell's fees, at the NBP rate of the day before the sale, less the consumed lots' cost at the rate of the day before each was opened, without buy fees, as each row's own tax counts them. Dividends and bond coupons stay out of the netting; a bond keeps its own row's tax, which the issuer withholds.
   - *Options:* keep; or add buy fees to the cost basis for tax; or bring foreign dividends into the netting.

46. **Raising the attention item for an unclassified instrument.**
   - *Problem:* D-132 asks the owner to classify an instrument once through an attention item (D-130), whose store isn't built yet.
   - *Today:* a revaluation of an OKI account raises `instrument-unclassified` for each holding the pack can't classify and the owner hasn't, and clears it once the owner has, through `app.AttentionRaiser` (`Raise`, `Clear`); `cmd/core` wires a stand-in that only logs them, and the settings page lists the instruments to classify.
   - *Options:* wire the attention store's port once D-130 lands; or have the attention app find every unclassified instrument at once (a check with a find function), so items resolve on their own.

## Kibi evidence

28. **Keeping the kibi sync patch, which makes the store defect rarer but doesn't fix it.**
    - *Today:* `patches/kibi-cli+2.36.0.patch` lets `kibi sync`'s compiling process exit on its own before the store is published, and refuses to publish if it had to be killed. In stress runs (three clones looping rebuild and check) failing checks went from 9 of 36 unpatched to 2 of 36 and 3 of 24 patched, and in every patched failure the process had exited cleanly, so the kill is at most part of the cause. The one-recompile stopgap in `scripts/kibi-check.sh` stays. The upstream report draft includes the patch, the numbers and the recurring missing edges.
    - *Options:* keep carrying the patch until upstream fixes the defect; or drop it and rely on the stopgap alone.

29. **The docs site waits for both end-to-end runs on main.**
    - *Today:* the docs-build job ingests the kibi proof receipts of the kind and k3d e2e runs and publishes only when both passed, so a failing e2e run on either cluster leaves the site at the last commit whose runs passed.
    - *Options:* keep; or publish whenever `check` passes, with whatever receipts exist (a failed run then shows as a failed receipt).

30. **A new scenario linked to a firm requirement.**
    - *Today:* `SCEN-e2e-first-passkey-localhost` (P-10's passkey criterion) specifies D-60's firm requirement; only the relationship was added, the requirement's statement is unchanged.
    - *Question:* is adding a scenario link to a firm requirement acceptable without the author, or should such links wait for the design session?
