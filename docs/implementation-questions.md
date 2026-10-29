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

4. **Seeding a broker account the owner maps after its connector synced.**
   - *Problem:* D-23 seeds an account from its broker's positions at go-live, and D-72 books nothing for an unmapped broker account. A connector usually syncs before the owner maps its accounts in the first-run wizard (D-73), so its first opening batch seeds nothing, and the ledger refuses to seed an account again with changed content.
   - *Today:* an instance counts as seeded only while every broker account mapped for it is seeded, so mapping one makes the next sync an opening batch again; that batch seeds it and leaves out the opening lots and cash of broker accounts already seeded, keeping their balance assertions ([Ingest](concepts/ingest.md#broker-accounts)). For IBKR, whose Flex Query sets its own period, nothing is lost; a connector whose watermark moves on an opening batch could skip the transactions of other accounts between the two syncs.
   - *Options:* keep; or have the connector contract seed one account at a time; or seed a newly mapped account only from an uploaded opening statement (D-109).

5. **The treasury-bonds connector still names its OpenFief account in its config.**
   - *Problem:* D-72's broker accounts cover broker connectors. The treasury-bonds plugin books every lot into the account its config's `account` names, so the owner types an OpenFief account ID at install, before the wizard creates the account.
   - *Today:* unchanged: the account must exist and carry that ID, and the wizard binds the instance to it.
   - *Options:* the core sends the lot's account with `ImportBondLots`, as it knows it; or the plugin reports one broker account per registry account like a broker connector.

6. **Removing or remapping sleeves, kinds, accounts and broker accounts.**
   - *Problem:* D-71 and D-72 let the owner add and rename; they don't say whether the owner can delete a sleeve, kind or account, or map a broker account to another account once mapped.
   - *Today:* nothing is deleted, and a mapped broker account stays mapped (`FAILED_PRECONDITION` when mapped again elsewhere). An unwanted default sleeve can only be renamed or left empty.
   - *Options:* allow deleting an empty sleeve, an unused kind and an account without transactions, and remapping a broker account whose account has no transactions from it; or keep everything append-only.

7. **Who may list broker account numbers.**
   - *Problem:* D-72 has the UI show unmapped broker accounts, which the read API lists (`ListBrokerAccounts`). Every read caller reads it: owner and read sessions, script tokens and view plugins such as Ghostfolio's.
   - *Today:* any read caller may list them, as it may read the rest of the record. The demo has none.
   - *Options:* keep; or list them only through the owner's `AccountService`, so read sessions and plugins never see an account number.
