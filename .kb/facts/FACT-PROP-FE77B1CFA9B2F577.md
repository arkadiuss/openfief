---
title: Stores Accounts And Log Exactly = true
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
subject_key: ledger.postgres_store
property_key: stores_accounts_and_log_exactly
canonical_key: mcp-kibi-model-requirement:ledger.postgres_store:stores_accounts_and_log_exactly:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EDF3153345BEC045
claim_text: The ledger must store accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.
origin:
  kind: agent
  recorded_at: '2026-11-03T11:50:18.680Z'
id: FACT-PROP-FE77B1CFA9B2F577
type: fact
---
