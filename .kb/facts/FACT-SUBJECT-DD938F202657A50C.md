---
title: Valuation River Queue
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: valuation.river_queue
canonical_key: valuation.river_queue
claim_key: CLAIM-8CBD69ADFDF487D2
claim_text: Revaluations must run as River jobs, at most one per account waiting or running, with the account's earliest pending date kept beside them and lowered by every enqueue, and a job must complete only in the transaction that clears that date, so that no change is left without a job.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:57:30.358Z'
id: FACT-SUBJECT-DD938F202657A50C
type: fact
---
