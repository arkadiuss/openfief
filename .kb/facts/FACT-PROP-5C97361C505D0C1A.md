---
title: Lot Booked Under Provisional Isin Until Listed = true
status: active
text_ref: design-decisions:D-214#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-214-3
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.treasury_bonds
property_key: lot_booked_under_provisional_isin_until_listed
canonical_key: mcp-kibi-model-requirement:plugins.treasury_bonds:lot_booked_under_provisional_isin_until_listed:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1331A4B38FD77167
claim_text: A lot of a series the Ministry of Finance's data doesn't list yet must be booked under a provisional ISIN, PL, the series code, 00 and a check digit, and must move to the registry's ISIN when the data lists it.
origin:
  kind: agent
  recorded_at: '2026-11-08T01:31:52.065Z'
id: FACT-PROP-5C97361C505D0C1A
type: fact
---
