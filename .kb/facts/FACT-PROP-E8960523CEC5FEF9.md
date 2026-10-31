---
title: Purchase Kinds = cash_purchase_at_100_pln, exchange_at_99_90_pln_except_ots, ots_exchange_at_100_00_pln
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-100-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-100#1
fact_kind: property_value
subject_key: ingest.bond_lot
property_key: purchase_kinds
operator: eq
value_type: set
value_set:
  - cash_purchase_at_100_pln
  - exchange_at_99_90_pln_except_ots
  - ots_exchange_at_100_00_pln
canonical_key: kb-requirements-req-ingest-bond-lot-exchange-price-by-letter-md:ingest.bond_lot:purchase_kinds:eq:cash_purchase_at_100_pln_exchange_at_99_90_pln_except_ots_ots_exchange_at_100_00_pln
claim_key: CLAIM-05F5E88D80E3AC36
claim_text: 'A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at the exchange price its series'' issue letter sets: 99.90 PLN per bond, or 100.00 PLN per bond for OTS.'
origin:
  kind: agent
  recorded_at: '2026-10-31T12:07:52.310Z'
id: FACT-PROP-E8960523CEC5FEF9
type: fact
---
