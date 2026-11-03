---
title: Ledger Postgres Store
status: active
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: ledger.postgres_store
canonical_key: ledger.postgres_store
claim_key: CLAIM-EDF3153345BEC045
claim_text: The ledger must store accounts, transaction versions with their postings and balance assertion checks in PostgreSQL, in a ledger schema it owns and creates with embedded migrations, with amounts and quantities as exact numerics that read back unchanged.
origin:
  kind: agent
  recorded_at: '2026-11-03T11:50:13.426Z'
id: FACT-SUBJECT-7FAB96CB2ABE1905
type: fact
---
