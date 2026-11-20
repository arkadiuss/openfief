---
title: Valuation Enqueue
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
subject_key: valuation.enqueue
canonical_key: valuation.enqueue
claim_key: CLAIM-97718BFE48BD8939
claim_text: Storing a price, FX rate, index reading, bond series terms or tax wrapper, and every ledger import, must enqueue its revaluations in the same database transaction, so that a rollback undoes both.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:57:06.330Z'
id: FACT-SUBJECT-379B2F59D33D8261
type: fact
---

Definition not stated: the subject valuation.enqueue is named by requirement REQ-valuation-enqueue-same-transaction, which links to the subject fact FACT-SUBJECT-379B2F59D33D8261 through constrains. The people who own that source can say what it is in the product and what it is not.
