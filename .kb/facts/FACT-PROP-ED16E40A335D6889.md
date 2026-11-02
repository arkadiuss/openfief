---
title: Condition Per Unclassified Isin = true
status: active
text_ref: docs/DESIGN.md#D-148
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-148
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: attention.unclassified_instrument_check
property_key: condition_per_unclassified_isin
canonical_key: mcp-kibi-model-requirement:attention.unclassified_instrument_check:condition_per_unclassified_isin:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B6F9E9F631DC9798
claim_text: The attention store must run a check whose conditions are the instruments priced at market held in the tax year an account in a wrapper with a yearly tax is valued through, which the pack can't classify and the owner hasn't, one per ISIN, so the item resolves once the owner classifies the instrument.
origin:
  kind: agent
  recorded_at: '2026-11-01T20:12:12.160Z'
id: FACT-PROP-ED16E40A335D6889
type: fact
---
