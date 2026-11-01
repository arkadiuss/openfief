---
title: Lot Deletion Deletes Its Coupons And Cash = true
status: active
text_ref: design-decisions:D-154#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-154-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.bond_lot_entry
property_key: lot_deletion_deletes_its_coupons_and_cash
canonical_key: mcp-kibi-model-requirement:ingest.bond_lot_entry:lot_deletion_deletes_its_coupons_and_cash:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-98D0E777122D264B
claim_text: Deleting a bond lot as a correction must also delete the coupons booked for it, and with them their cash.
origin:
  kind: agent
  recorded_at: '2026-11-01T12:29:54.180Z'
id: FACT-PROP-CC4B098488DA5155
type: fact
---
