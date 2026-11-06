---
title: 'A bond lot deletion and a sync of the account''s connector instance must never overlap: the deletion must hold the instance''s sync claim while it runs and be refused while a sync holds it, and a sync must read what the bound accounts hold only under its claim, so no sync books a coupon of a lot deleted after it read the lots.'
status: open
text_ref: docs/DESIGN.md#D-154
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-154
  - lane:strict
semantic_text: 'A bond lot deletion and a sync of the account''s connector instance must never overlap: the deletion must hold the instance''s sync claim while it runs and be refused while a sync holds it, and a sync must read what the bound accounts hold only under its claim, so no sync books a coupon of a lot deleted after it read the lots.'
logic_claims:
  - CLAIM-E49FF22615591447
semantic_clauses:
  - 'A bond lot deletion and a sync of the account''s connector instance must never overlap: the deletion must hold the instance''s sync claim while it runs and be refused while a sync holds it, and a sync must read what the bound accounts hold only under its claim, so no sync books a coupon of a lot deleted after it read the lots.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a5563e9de615a212c85a2d090615987f31da1bd6e7c3c9cc20e156608a090a54
semantic_inventory:
  - claim_key: CLAIM-E49FF22615591447
    claim_text: 'A bond lot deletion and a sync of the account''s connector instance must never overlap: the deletion must hold the instance''s sync claim while it runs and be refused while a sync holds it, and a sync must read what the bound accounts hold only under its claim, so no sync books a coupon of a lot deleted after it read the lots'
    role: normative
    status: modeled
    span:
      start: 0
      end: 325
origin:
  kind: agent
  recorded_at: '2026-11-02T11:50:07.324Z'
id: REQ-ingest-bond-lot-delete-excludes-sync
type: req
---
A bond lot deletion and a sync of the account's connector instance must never overlap: the deletion must hold the instance's sync claim while it runs and be refused while a sync holds it, and a sync must read what the bound accounts hold only under its claim, so no sync books a coupon of a lot deleted after it read the lots.

## Context

D-154 has deleting a bond lot also delete the coupons booked for it and their cash. A sync reads the account's open lots when it starts and imports the coupons when the connector answers, so a deletion in between left a coupon of the deleted lot booked: cash that was never paid, and silent. On 2026-11-02 the design session, for the author, asked for the race to be closed rather than documented, choosing between serializing the deletion with the instance's sync and refusing coupons of a lot no longer open at import. This requirement takes the first: the scheduler's single-flight claim in the ingest schema already serializes syncs, and refusing at import would need the core to read the treasury-bonds connector's keys, while the core holds no bond logic (D-117).

## Source

> **[D-154]** Deleting a bond lot as a correction also deletes the coupons booked for it, and with them their cash, since those coupons were never paid.

docs/DESIGN.md#D-154
