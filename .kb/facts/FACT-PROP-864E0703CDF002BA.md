---
title: Deletion Never Overlaps A Sync = true
status: active
text_ref: docs/DESIGN.md#D-154
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-154
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.bond_lot_entry
property_key: deletion_never_overlaps_a_sync
canonical_key: mcp-kibi-model-requirement:ingest.bond_lot_entry:deletion_never_overlaps_a_sync:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E49FF22615591447
claim_text: 'A bond lot deletion and a sync of the account''s connector instance must never overlap: the deletion must hold the instance''s sync claim while it runs and be refused while a sync holds it, and a sync must read what the bound accounts hold only under its claim, so no sync books a coupon of a lot deleted after it read the lots.'
origin:
  kind: agent
  recorded_at: '2026-11-02T11:50:03.996Z'
id: FACT-PROP-864E0703CDF002BA
type: fact
---
