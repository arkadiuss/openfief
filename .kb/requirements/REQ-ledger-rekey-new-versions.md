---
title: The ledger must re-key an instrument by appending, in one database transaction, a new version under the same key of each current transaction that posts to it, with only its holding postings moved, running the import hooks, carrying its display names over, and changing nothing when run again.
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: The ledger must re-key an instrument by appending, in one database transaction, a new version under the same key of each current transaction that posts to it, with only its holding postings moved, running the import hooks, carrying its display names over, and changing nothing when run again.
logic_claims:
  - CLAIM-5730629E84BFDF73
semantic_clauses:
  - The ledger must re-key an instrument by appending, in one database transaction, a new version under the same key of each current transaction that posts to it, with only its holding postings moved, running the import hooks, carrying its display names over, and changing nothing when run again.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b784cb5f02cd3ae4013e08ef0356f5b7bf0e995a34a48cba0a7d7dfd63d35296
semantic_inventory:
  - claim_key: CLAIM-5730629E84BFDF73
    claim_text: The ledger must re-key an instrument by appending, in one database transaction, a new version under the same key of each current transaction that posts to it, with only its holding postings moved, running the import hooks, carrying its display names over, and changing nothing when run again
    role: normative
    status: modeled
    span:
      start: 0
      end: 291
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:22.475Z'
id: REQ-ledger-rekey-new-versions
type: req
---
The ledger must re-key an instrument by appending, in one database transaction, a new version under the same key of each current transaction that posts to it, with only its holding postings moved, running the import hooks, carrying its display names over, and changing nothing when run again.

## Context

D-214 has a treasury bond identified by the registry's ISIN from the Ministry of Finance's data, and a lot of a series the data doesn't list yet wait for its ISIN as it waits for its terms; installs from before it hold every series under the ISIN D-99 derived from the series code, and D-161 asks that stored data be migrated. The ledger is append-only (D-18), so a re-key supersedes and keeps the old versions. Modeled by Claude Code for D-214; the author has not reviewed the wording.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97), never one derived from its series code; the series code is its ticker.

docs/DESIGN.md#D-214
