---
title: When the owner deletes a bond lot, the core must send the connector bound to its account the account's current transactions that instance delivered, keyed as the connector keyed them and dated, and the treasury-bonds connector must deliver a deletion of each one booked under the lot's key, its coupons, its redemption and their payout withdrawals, besides the lot's own transactions.
status: open
text_ref: docs/DESIGN.md#D-154
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-154
  - lane:strict
semantic_text: When the owner deletes a bond lot, the core must send the connector bound to its account the account's current transactions that instance delivered, keyed as the connector keyed them and dated, and the treasury-bonds connector must deliver a deletion of each one booked under the lot's key, its coupons, its redemption and their payout withdrawals, besides the lot's own transactions.
logic_claims:
  - CLAIM-8DC45663FB43E2E4
semantic_clauses:
  - When the owner deletes a bond lot, the core must send the connector bound to its account the account's current transactions that instance delivered, keyed as the connector keyed them and dated, and the treasury-bonds connector must deliver a deletion of each one booked under the lot's key, its coupons, its redemption and their payout withdrawals, besides the lot's own transactions.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2b97ed9820bafffb96d8b732cde17d1a3c7af0c05b15086efdce86d6cacc3d31
semantic_inventory:
  - claim_key: CLAIM-8DC45663FB43E2E4
    claim_text: When the owner deletes a bond lot, the core must send the connector bound to its account the account's current transactions that instance delivered, keyed as the connector keyed them and dated, and the treasury-bonds connector must deliver a deletion of each one booked under the lot's key, its coupons, its redemption and their payout withdrawals, besides the lot's own transactions
    role: condition
    status: modeled
    span:
      start: 0
      end: 383
origin:
  kind: agent
  recorded_at: '2026-11-02T10:24:36.615Z'
id: REQ-treasurybonds-lot-delete-deletes-booked
type: req
---
When the owner deletes a bond lot, the core must send the connector bound to its account the account's current transactions that instance delivered, keyed as the connector keyed them and dated, and the treasury-bonds connector must deliver a deletion of each one booked under the lot's key, its coupons, its redemption and their payout withdrawals, besides the lot's own transactions.

## Context

D-154 has deleting a bond lot as a correction also delete the coupons booked for it and their cash. The connector stores nothing (D-117), so the core tells it what it booked, as a sync request tells it what the account holds. Deleting the lot's redemption at maturity and the payout withdrawals of D-135 and D-162 too is this implementation's reading of the same rule: they were booked for the lot and would leave its cash behind. Reason for sending only current transactions: a superseded one needs no deletion.

## Source

> **[D-154]** Deleting a bond lot as a correction also deletes the coupons booked for it, and with them their cash, since those coupons were never paid.

docs/DESIGN.md#D-154
