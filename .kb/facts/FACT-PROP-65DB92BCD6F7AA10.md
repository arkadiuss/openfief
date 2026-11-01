---
title: Books Maturity As Redemption = true
status: active
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: treasurybonds.maturities
property_key: books_maturity_as_redemption
canonical_key: mcp-kibi-model-requirement:treasurybonds.maturities:books_maturity_as_redemption:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2825F44C127438A9
claim_text: The treasury-bonds connector must book each open lot that matured on or before today as a redemption on its maturity date at the redemption value the Polish pack gives it, with Belka tax on its interest and purchase discount by the account's wrapper, keyed per lot so it is booked once.
origin:
  kind: agent
  recorded_at: '2026-11-01T07:45:21.696Z'
id: FACT-PROP-65DB92BCD6F7AA10
type: fact
---
