---
title: Valid Only If Valuable On Purchase Date = true
status: active
text_ref: docs/DESIGN.md#D-100
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-100
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.bond_lot
property_key: valid_only_if_valuable_on_purchase_date
canonical_key: mcp-kibi-model-requirement:ingest.bond_lot:valid_only_if_valuable_on_purchase_date:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B2DC2C0B0656E67B
claim_text: A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, a purchase for cash at the nominal or by exchange at the exchange price, and can be valued on its purchase date by the Polish bond rules at that price.
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:10.892Z'
id: FACT-PROP-5F3B2B18543AF0ED
type: fact
---
