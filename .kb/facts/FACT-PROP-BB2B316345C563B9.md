---
title: Sync Fills Missing Terms = true
status: active
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: treasurybonds.terms
property_key: sync_fills_missing_terms
canonical_key: mcp-kibi-model-requirement:treasurybonds.terms:sync_fills_missing_terms:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-819A0058C87E75E1
claim_text: A lot that leaves every term to the connector, of a series the ministry's data doesn't list yet, must be booked without terms, and a later sync must fill in the terms of each held series the core stores none of once the data lists it, never replacing stored terms.
origin:
  kind: agent
  recorded_at: '2026-11-07T13:27:07.465Z'
id: FACT-PROP-BB2B316345C563B9
type: fact
---
