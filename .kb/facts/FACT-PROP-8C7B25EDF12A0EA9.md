---
title: Confirmed Once By Owner = true
status: active
text_ref: docs/DESIGN.md#D-116
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-116
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ticker_entries
property_key: confirmed_once_by_owner
canonical_key: mcp-kibi-model-requirement:ingest.ticker_entries:confirmed_once_by_owner:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-55C9BD1974A13EC9
claim_text: The owner must confirm a ticker's ISIN once, proposed or typed, through an owner-only service on the core's plugin port, which stores it and books every entry waiting for it.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:44.112Z'
id: FACT-PROP-8C7B25EDF12A0EA9
type: fact
---
