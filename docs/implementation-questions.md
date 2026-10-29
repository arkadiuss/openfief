---
title: Implementation questions
description: Questions the implementation raised that no decision answers yet, each with what is built today and the options, for the author to decide.
---

The implementation of M0 ran into these questions. No decision in [DESIGN.md](DESIGN.md) answers them, so each was built on the narrowest reading (or not built), and the code works as described under **Today**. None of this is decided: when the author settles an item, it becomes a decision in DESIGN.md (or a change to one) and leaves this page.

Items are grouped by area and numbered for discussion only; the numbers are not design IDs.

## Blocking real use

1. **Nothing can create an account or sleeve outside the demo.**
   - *Problem:* only `cmd/demo-seed` calls `SaveAccount` and `SaveSleeve`, and `core connector bind` needs an account that already exists. A real installation therefore has no account for any connector to write into, so it can't hold data. No decision says who creates accounts.
   - *Today:* nothing is built.
   - *Options:* the owner creates accounts and sleeves in the built-in UI, through the owner-only write path of item 4; the `openfief` CLI does it; or a connector creates its account on its first sync (and the owner assigns the sleeve).

2. **The Ghostfolio view can't reach the owner's Ghostfolio.**
   - *Problem:* a plugin's manifest is its signed image label (D-42), and its egress allow-list holds only the hosts the manifest requests and the grant approves (D-16). An image built and signed in CI can't name a host the owner picks at install time.
   - *Today:* the Ghostfolio manifest requests no host, so the plugin refuses to start without one. The owner would have to rebuild the image with their host and sign it themselves (D-45).
   - *Options:* a manifest placeholder (for example "the host of config field `url`") that `openfief plugin install` fills from the config and the grant approves explicitly; or an owner-signed overlay of the manifest; or something else.

3. **XTB import needs the export's real layout.**
   - *Problem:* D-9 says XTB data arrives as uploaded xStation export files, but the export's columns, and whether it separates CFDs from stocks, are unverified (OQ-22).
   - *Today:* no XTB parser exists.
   - *Needed:* a header row and two or three rows with made-up values, copied from a real export. The parser's fixtures will be synthetic.

## Write paths for the owner

4. **How the owner's entries reach the core.**
   - *Problem:* D-52 says the model portfolio is entered in the UI and handed to the core "like manual bond lots" (D-49), but no decision gives the shape of that path.
   - *Today:* the core serves `ModelPortfolioService` and `BondLotService` next to the AuthService. Only view plugins may call them (the AuthService's caller check), and each call carries an owner-scope session for an allowed origin. The read API stays read-only. Statement uploads (`StatementService`) work the same way.
   - *Options:* keep this; or one owner write service for all such entries; or something else.

5. **The first passkey from the `openfief` CLI.**
   - *Problem:* D-60 says the CLI prints the one-time link for the first passkey. Today `core enroll-link` prints it, run inside the core's pod, and it needs the database.
   - *Today:* the owner runs `kubectl exec ... core enroll-link`.
   - *Options:* `openfief` wraps it by executing the command in the core's pod (two new dependencies); or it stays a `kubectl exec`, and D-60's wording changes.

6. **Statement uploads (D-7, D-49).**
   - *Today:* an upload goes to the one connector instance bound to the account (none or several is refused), and its delivery must stay within that account, so a Flex statement covering several accounts imports only the uploaded account's part. The owner picks the mode (opening or transactions) on each upload. The form shows on every account page, because the read API doesn't say which connector is bound; a connector that takes no files refuses with its reason. An opening upload doesn't mark the instance as seeded for its scheduled syncs. The file name and the core's sniffed media type are hints only. The core caps a file at 64 MiB, the largest any connector accepts today.
   - *Question:* keep these?

## Identity and the UI

7. **View plugins that browsers reach.**
   - *Problem:* D-59 says only the core may call a plugin. The built-in UI is a view plugin that browsers reach through the Ingress.
   - *Today:* the demo allows it with a demo-only NetworkPolicy from Traefik (like the core's own). The operator's rules are unchanged.
   - *Options:* an explicit exception in D-59 for view plugins that serve browsers, and what the operator then compiles for them; or serve the UI from the core.

8. **The core's own cluster role for TokenReview.**
   - *Problem:* the core checks view plugins' tokens with TokenReview, so it has a narrow ClusterRole (create `tokenreviews` only). That is the second cluster-scoped grant after the egress proxy's (D-63).
   - *Options:* keep it; or another way to authenticate view plugins.

9. **Which tokens the read API accepts (D-27).**
   - *Today:* owner sessions, script tokens and view plugins' projected tokens, checked by the same rules as the AuthService's callers. Connector tokens are refused, since the core calls connectors and not the other way round.
   - *Question:* is that the intended scope?

10. **`OPENFIEF_AUTH_CALLER_TOKEN`.**
   - *Today:* a development aid that lets a UI running outside a cluster call the AuthService and the read API with a fixed token. It is ignored in a cluster.
   - *Options:* keep; or remove.

11. **The UI's session cookie on `http://localhost`.**
    - *Today:* the cookie is `__Host-openfief_session`, which must be `Secure`, so it is `Secure` on localhost too. Chrome and Firefox accept that on localhost; Safari is unverified.
    - *Options:* keep; or drop the `__Host-` prefix on localhost only.

12. **Passkey login in the end-to-end tests.**
    - *Problem:* `--auth-origin` accepts plain HTTP only on localhost, so the e2e host `http://openfief.e2e.test` isn't allowed, and the e2e suite checks only that the login page renders.
    - *Options:* a test-only flag that allows one insecure origin; or keep the e2e check as it is.

13. **Who sees the list of installed plugins.**
    - *Today:* `/plugins` in the UI, and `GET /v1/plugins`, are open to any session, including the public demo's read-only auto-login. They show image digests and signer names.
    - *Options:* keep; or owner scope only.

14. **Signature trust shown for each plugin (D-46, D-56).**
    - *Today:* the operator writes an additive `status.signature` (trust, signer kind, signer) on the v1alpha1 `Plugin`, because the `SignatureVerified` condition alone can't tell OpenFief's CI from the owner's own signers. The UI shows seven levels: verified, signed by OpenFief CI, by a catalog publisher, by the owner, by a trusted signer of unknown kind, refused, not checked.
    - *Question:* is the status field and that set of levels right?

## Plugins and the CLI

15. **First-party plugins in the catalog (D-43, D-44).**
    - *Today:* `nbp-fx`, `ibkr-flex`, `eodhd`, `treasury-bonds`, `ghostfolio` and `ui` are built and signed by CI but not listed in the catalog, so they show as "signed by OpenFief CI", not "verified".
    - *Options:* add each with its signed digest per release; or leave first-party plugins as "signed by OpenFief CI".

16. **Secrets shared between instances of one image.**
    - *Problem:* Secret names come from the signed manifest, so two instances of one image (one per IBKR account, D-48) read the same Secrets.
    - *Today:* uninstall keeps a Secret another instance still names, and says so.
    - *Options:* a per-instance naming convention (for example a prefix the operator adds); or keep one Secret per image.

17. **Data from a removed instance (D-58).**
    - *Today:* a transaction is marked as coming from a removed instance when no `Plugin` of that name exists. The mark is worked out on every read, so the log stays append-only and reinstalling under the same name clears it. Only transactions are marked, not lots or holdings, and nothing is marked outside a cluster.
    - *Question:* mark lots and holdings too?

18. **CLI behaviours.**
    - *Today:* a catalog name shared by several versions is refused, with the list to choose from, instead of picking the newest. Installing over an existing name is refused; there is no `upgrade` command (D-47 leaves upgrades to the owner's own edit). `--print` without `--trusted-signers` trusts only OpenFief's CI and the catalog.
    - *Question:* keep these?

19. **Shipping the `openfief` binary.**
    - *Today:* it builds with `go build ./cmd/openfief`; no release step publishes it.
    - *Options:* a signed release (goreleaser or similar) per tag; or later.

20. **Core egress to the Kubernetes API.**
    - *Problem:* k3s serves the API on the node's address, so the product's NetworkPolicy lets the core reach port 6443 on any address.
    - *Options:* narrow it per install (an overlay with the API server's address); or document it.

## Market data and connectors

21. **Who names what a market-data connector prices (D-22).**
    - *Today:* the owner lists ISINs and indices in the `eodhd` instance's config.
    - *Options:* the core sends the held instruments and the model portfolio's indices in the sync request (a contract change); or keep the config list.

22. **Shared market data.**
    - *Problem:* any connector instance can write prices, FX rates and index readings, each stamped with its source, and a later write replaces an earlier one.
    - *Options:* restrict each kind of data to some instances (for example only an FX connector writes FX rates); or rank sources; or keep the last write.

23. **EODHD details.**
    - *Today:* a market index is identified by the owner's name for it (for example "WIG20"). Euronext means EODHD's PA, AS, BR, LS, IR, MI and OL markets. An ISIN listed on several supported exchanges is reported as ambiguous until the owner sets `exchange`. A series that can't be priced is reported again on every sync.
    - *Question:* keep these?

24. **Real or synthetic FX rates on the demo.**
    - *Today:* the demo runs the `nbp` instance of `nbp-fx`, which writes real NBP rates next to the demo dataset's own rates.
    - *Options:* keep the plugin on the demo; or let the demo dataset own FX.

25. **The first sync of an account seeded from opening lots.**
    - *Today:* it fails until `core connector sync --mode transactions` is run once.
    - *Options:* the core detects an already-seeded account and switches to transactions mode; or keep the manual step.

## Treasury bonds

26. **Series terms are entered with each lot.**
    - *Problem:* D-9 lists series, quantity and purchase date, but valuing a lot also needs the series' first-period rate, margin and early-redemption fee.
    - *Today:* the owner enters them with each lot, and the latest entry for a series wins.
    - *Options:* keep; or fetch a catalogue of series terms.

27. **A bond lot is a deposit and a buy, not an opening lot.**
    - *Problem:* the ledger refuses to seed an account twice, so lots entered one at a time can't all be opening lots.
    - *Today:* each lot is a deposit of its cost plus a buy on its purchase date. An old lot therefore moves go-live, and the start of the time-weighted return, back to its date.
    - *Question:* acceptable?

28. **The bond's ISIN is derived from its series code.**
    - *Today:* `PL` + series code + `00` + a check digit (for example `PLEDO0536000`), not the ISIN the registry assigns, which owners don't see. A future import of the bond service's own export must use the same derivation, or a series would be held twice.
    - *Question:* acceptable?

29. **Bond lot limits.**
    - *Today:* only cash purchases at 100 PLN (not exchanges at 99.90 PLN, whose discount is taxed). No way to delete a lot; entering the same series and day replaces it. "The right connector" (D-49) is the one instance bound to the account. The UI shows the form only for accounts of institution `treasury-bonds`. Entry works only in a cluster with connectors on.
    - *Question:* keep these?

30. **COI coupons are booked as dividends.**
    - *Problem:* the ledger has no interest transaction kind.
    - *Options:* add an interest kind; or keep dividends.

31. **Who picks the NBP reference rate in force.**
    - *Today:* the treasury-bonds connector applies the rule itself: the rate in force on the tenth business day before a month is that month's reading, dated on the month's first day. The comment on `IndexReading.date` in the connector contract now says so; the wire format is unchanged.
    - *Options:* keep; or the connector sends NBP's raw rate changes and the core picks.

32. **The Polish business-day calendar.**
    - *Today:* Saturdays, Sundays and the holidays in art. 1 of the act on days off work, year by year (6 January from 2011, 24 December from 2025, Easter-based days). Epidemic days off (art. 1a) are not modelled. Dates before 1991 are refused.
    - *Question:* enough for M0?

33. **GUS CPI defaults.**
    - *Today:* only final monthly values are used; preliminary, estimated and flash values are skipped and arrive later. The first sync starts January 13 years back (`start` can change it), fetching at most 4 years per sync to stay within GUS's anonymous limits. GUS's indicator 1832 is used up to 2025 and 2728 from 2026, a boundary written into the code. A reference-rate month is delivered once its day has passed and is not fetched again.
    - *Question:* keep these?

## Performance and the model portfolio

34. **Transfers between the owner's own accounts (D-61).**
    - *Problem:* the ledger doesn't record which withdrawal became which deposit, but the time-weighted return must not count such transfers as cash flows.
    - *Today:* a withdrawal matches a deposit into another account of the same amount and currency at most 5 calendar days apart, the nearest unmatched deposit first. The performance response lists every match.
    - *Options:* keep the rule (and its window); or record transfers explicitly in the ledger.

35. **Time-weighted return conventions.**
    - *Today:* cash flows count at the end of their day; each daily factor and the running product are rounded half-even to 12 decimal places; opening lots count as inflows at their market value on their date.
    - *Question:* keep these?

36. **What the value-over-time chart draws.**
    - *Today:* the chart draws the time-weighted return index from 100 at go-live (D-30 and D-53 read as one chart with two lines from 100), with the value including cash shown alongside. `ListValuationTotals` still sums holdings only, without cash.
    - *Options:* keep; or include cash in `ListValuationTotals` too.

37. **Model portfolio details (D-52 to D-54).**
    - *Today:* the model rebalances at the end of each 31 December. A target's own drawdown is in the reporting currency (PLN) over the span where it has both closes and FX rates; P-48, not decided, would use its own currency over its whole history. Weights go in 0.01% steps, at most 50 targets. While the model is off, the read API answers 404.
    - *Question:* keep these?

## Ghostfolio

38. **Where the Ghostfolio view departs from D-64.**
    - *Today:* the plugin derives its owned accounts' IDs (UUIDv5 of instance and account) instead of generating and storing them, because a plugin has no volume and no Kubernetes API. A stale activity that Ghostfolio's duplicate check would skip is deleted before the import, or the loop never converges. The plugin runs its own timer, since D-49 has the core schedule only connectors.
    - *Question:* acceptable?

39. **How transactions become Ghostfolio activities.**
    - *Today:* a dividend is quantity 1 at the gross amount, with the withheld tax as its fee. Standalone fees go against a manual "OpenFief fees" asset per currency. Cash movements produce no activity, so Ghostfolio shows holdings without cash. Listed instruments need the owner's mapping from ISIN to Ghostfolio's data source and symbol. Bond assets are keyed by the series code only, so two users of one Ghostfolio would collide.
    - *Question:* keep these?
