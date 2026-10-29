---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Blocking real use

1. **XTB import needs the export's real layout.**
   - *Problem:* D-9 says XTB data arrives as uploaded xStation export files, but the export's columns, and whether it separates CFDs from stocks, are unverified (OQ-22).
   - *Today:* no XTB parser exists.
   - *Needed:* a header row and two or three rows with made-up values, copied from a real export. The parser's fixtures will be synthetic.
   - *Known so far:* the author holds stocks and ETFs only at XTB (so D-24's CFD handling stays as written) and will send the headers from a real export. The export's language is not known yet. Public parsers disagree on the sheet names (R-19).

## Plugins

2. **How the operator reads an in-cluster service in another namespace.**
   - *Problem:* D-74 has the operator add a NetworkPolicy rule to an in-cluster service's pods and hand the plugin its ClusterIP, which means reading that Service; D-15 confines the operator to a Role in `openfief`, and D-82 counts the cluster-wide permissions OpenFief holds.
   - *Today:* the owner's install (with the owner's kubeconfig) creates, in the service's namespace, a Role and RoleBinding `openfief-plugin-<instance>` that let the operator's ServiceAccount `get` that one Service; the operator reads it uncached, retries every minute when it can't, and doesn't watch it, so a changed selector or ClusterIP takes effect at the next reconcile. Uninstall deletes them.
   - *Options:* keep; have the install record the Service's ClusterIP, selector and target port in the grant so the operator reads nothing outside `openfief`; or a ClusterRole to read Services.

3. **How a release's catalog reaches `main`, and where a beta's CLI goes.**
   - *Problem:* D-88 has the release workflow add each first-party plugin to the catalog, a file in this repository embedded in the operator and CLI, and docs/dev/sdlc.md has every release ship the signed CLI. The release starts from `main`, which is never committed to directly, and a beta so far had no GitHub release.
   - *Today:* the plugin images are built and signed first; the workflow commits the catalog to a new branch `release/<version>`, builds the other images and the CLI from that commit, tags it, and opens a pull request to `main`. A beta becomes a GitHub pre-release carrying only the signed CLI.
   - *Options:* keep; push the catalog commit to `main` from the workflow; or keep betas without a GitHub release and ship their CLI as a workflow artifact only.

## Onboarding

4. **The treasury-bonds connector still names its OpenFief account in its config.**
   - *Problem:* D-72's broker accounts cover broker connectors. The treasury-bonds plugin books every lot into the account its config's `account` names, so the owner types an OpenFief account ID at install, before the wizard creates the account.
   - *Today:* unchanged: the account must exist and carry that ID, and the wizard binds the instance to it.
   - *Options:* the core sends the lot's account with `ImportBondLots`, as it knows it; or the plugin reports one broker account per registry account like a broker connector.

5. **Removing or remapping sleeves, kinds, accounts and broker accounts.**
   - *Problem:* D-71 and D-72 let the owner add and rename; they don't say whether the owner can delete a sleeve, kind or account, or map a broker account to another account once mapped.
   - *Today:* nothing is deleted, and a mapped broker account stays mapped (`FAILED_PRECONDITION` when mapped again elsewhere). An unwanted default sleeve can only be renamed or left empty.
   - *Options:* allow deleting an empty sleeve, an unused kind and an account without transactions, and remapping a broker account whose account has no transactions from it; or keep everything append-only.

6. **Who may list broker account numbers.**
   - *Problem:* D-72 has the UI show unmapped broker accounts, which the read API lists (`ListBrokerAccounts`). Every read caller reads it: owner and read sessions, script tokens and view plugins such as Ghostfolio's.
   - *Today:* any read caller may list them, as it may read the rest of the record. The demo has none.
   - *Options:* keep; or list them only through the owner's `AccountService`, so read sessions and plugins never see an account number.

## Market data

7. **Which connector instances are asked what to price.**
   - *Problem:* D-93 has each sync request to a market-data connector carry the instruments the owner holds and the model portfolio's targets. No manifest field says that a connector is a market-data connector, and the list tells a connector what the owner holds at every institution.
   - *Today:* the core sends the list to every connector instance bound to no account, since such an instance can deliver only market data; a broker connector, which is bound to the accounts it feeds, gets it empty. A market-data connector the owner bound to an account would get nothing. Held means a non-zero holding today, treasury bonds with series terms left out; the model's targets count only while it is on (D-52).
   - *Options:* keep; or a manifest field declaring a market-data connector, approved with the grant like the cross-check prices of D-112; or send it to every connector.

8. **What a market-data source is, and what a ranking covers.**
   - *Problem:* D-94 has the owner rank "the sources" of prices, exchange rates and index readings, with the first-party connector of each first by default, and keep the lower-ranked values as cross-checks (D-22). It doesn't say whether a source is a connector instance or a plugin, how the core knows which instance is the first-party one, whether D-22's cross-check prices are a ranked source, or whether index levels are ranked.
   - *Today:* a source is the connector instance that delivered the value (D-48). Its default place comes from the institution it names on its deliveries: `eodhd` ranks first for prices, `nbp` for exchange rates and `treasury-bonds` for index readings, the rest by name. Cross-check prices are stored apart and never win (D-112 gates who may send them); lower-ranked prices are checked against the winner like cross-check prices. Index levels of market indices keep one value per index and date, the last delivered. A change of ranking revalues every account from its first transaction.
   - *Options:* keep; or rank by plugin image rather than instance, or by a manifest field naming the kinds a connector serves; or rank IBKR's mark prices as one more price source, so they fill days the market-data connector lacks; or rank index levels too.

## Values

9. **OKI's act: its in-force date, limit and rate.**
   - *Problem:* D-76 computes OKI's yearly tax "only once the act is in force, with the limit and rate taken from the act". The code knows no act in force, nor its numbers, and doesn't know how a year the act enters mid-way is counted, or whether the average balance is taken over every calendar day or only the days with a valuation.
   - *Today:* `pl.OKIAct` holds the in-force date, the limit and the rate as data, and `pl.CurrentOKIAct` is not in force (no date), so no tax is ever computed and the read API says so. Under an act, the tax would be the rate of the part above the limit of the average of the daily balances (gross values with cash, summed over every OKI account) of the days of the year on or after the in-force date that are valued, rounded half up to grosze.
   - *Needed:* the act's in-force date, limit and rate, and how it defines the average balance; then `CurrentOKIAct` gets them.

10. **How the taxes inside the values are counted.**
   - *Problem:* D-77 has the pack compute net and liquidation values "after the taxes due then", per holding. Polish tax counts gains per tax year over all disposals (PIT-38), rounds the base and the tax to whole złoty, and taxes a return from IKE on what is withdrawn above what was paid in.
   - *Today:* each holding is taxed on its own gain over its open lots' cost, at 19%, rounded half up to grosze (as D-37 rounds bond interest), and a loss is untaxed but offsets nothing. An IKE's early return is taxed the same way per holding, so dividends and realized gains kept as cash in the IKE are not taxed in its liquidation value. IKZE's 10% exit tax and the owner's marginal rate before 65 apply to every holding and to cash, each rounded half up to grosze.
   - *Options:* keep; or compute each account's taxes over its total (losses offsetting gains, IKE on payments in), which needs the account's contributions; or round as PIT-38 does.

11. **What a sale's cost covers.**
   - *Problem:* D-78's fee rule is "a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency". It doesn't say which currency the minimum is in, whether bonds, whose redemption costs no commission, pay it, or what foreign is relative to.
   - *Today:* the minimum is an amount in any currency, converted at the FX rate of the valuation's day (a missing rate leaves the liquidation value missing); the commission applies to holdings priced at market only, not to bonds or cash; and foreign means priced in another currency than the pack's, PLN for Poland. The minimum applies to the commission alone, before the conversion percentage is added.
   - *Options:* keep; or tie the minimum to the instrument's currency; or charge the conversion percentage on foreign cash too.
