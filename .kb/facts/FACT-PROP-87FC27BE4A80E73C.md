---
title: Postgres Exact Round Trip = true
status: active
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.store
property_key: postgres_exact_round_trip
canonical_key: mcp-kibi-model-requirement:ledger.store:postgres_exact_round_trip:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2EA08200020CBC78
claim_text: The ledger must store sleeves, accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:29:31.464Z'
id: FACT-PROP-87FC27BE4A80E73C
type: fact
---
