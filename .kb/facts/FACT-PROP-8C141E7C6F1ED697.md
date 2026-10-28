---
title: Pairs Withholding With Dividend = true
status: active
text_ref: docs/DESIGN.md#D-19
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-19
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ibkr_flex
property_key: pairs_withholding_with_dividend
canonical_key: mcp-kibi-model-requirement:ingest.ibkr_flex:pairs_withholding_with_dividend:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5BCEDC3E9D48F7B9
claim_text: An IBKR Flex Withholding Tax row must be booked with the dividend that has the same actionID, or when either lacks one, the same ISIN, currency and date; an unmatched, ambiguous or positive withholding row must become an unsupported ingest problem.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:53.347Z'
id: FACT-PROP-8C141E7C6F1ED697
type: fact
---
