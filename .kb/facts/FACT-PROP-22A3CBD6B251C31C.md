---
title: Opening Lots Join Seeding = true
status: active
text_ref: docs/DESIGN.md#D-136
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-136
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ticker_entries
property_key: opening_lots_join_seeding
canonical_key: mcp-kibi-model-requirement:ingest.ticker_entries:opening_lots_join_seeding:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5DC10116F6894036
claim_text: A ticker entry may be an opening lot carrying the date its units were first acquired; once its ticker's ISIN is confirmed it must be added to its account's seeding on the go-live date, and an opening batch of a broker account already seeded must leave it out.
origin:
  kind: agent
  recorded_at: '2026-11-01T07:36:06.182Z'
id: FACT-PROP-22A3CBD6B251C31C
type: fact
---
