---
title: Valuation Revaluation
status: active
text_ref: design-decisions:D-36#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-3
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: valuation.revaluation
canonical_key: valuation.revaluation
claim_key: CLAIM-55D889CC632D8C71
claim_text: Any change to the transaction log, prices, FX rates or index readings must enqueue a revaluation in the same database transaction.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:29:16.270Z'
id: FACT-SUBJECT-201BF18D8A3B745A
type: fact
---

Definition not stated: the subject valuation.revaluation is named by requirements REQ-decision-d36-10-fixed-revaluation-triggers, REQ-decision-d36-3-revaluation-same-transaction, REQ-decision-d36-4-river-job-queue, REQ-decision-d36-5-revaluation-coalescing, REQ-decision-d36-8-concurrency-and-schedule, REQ-decision-d36-8-revaluation-settings, REQ-valuation-revaluation-rows, which link to the subject fact FACT-SUBJECT-201BF18D8A3B745A through constrains. The people who own that source can say what it is in the product and what it is not.
