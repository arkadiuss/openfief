---
title: Ingest Logging
status: active
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.logging
canonical_key: ingest.logging
claim_key: CLAIM-0A05AB1AE74B1EA8
claim_text: The ingest context must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry, and the ledger's refusals of a transaction must not quote its amount or quantity.
origin:
  kind: agent
  recorded_at: '2026-11-03T01:05:21.983Z'
id: FACT-SUBJECT-30FC87DCAAFC7CB1
type: fact
---
