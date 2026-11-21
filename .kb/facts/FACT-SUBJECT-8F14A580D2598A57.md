---
title: Attention Outbox
status: active
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: attention.outbox
canonical_key: attention.outbox
claim_key: CLAIM-E661487AD21DB6E1
claim_text: Each change to an attention item and its outbox event must be written in one database transaction, and the items, their history and the outbox must be append-only.
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:57.207Z'
id: FACT-SUBJECT-8F14A580D2598A57
type: fact
---

Definition not stated: the subject attention.outbox is named by requirements REQ-attention-outbox-read-subject, REQ-attention-outbox-same-transaction, REQ-attention-relay, which link to the subject fact FACT-SUBJECT-8F14A580D2598A57 through constrains. The people who own that source can say what it is in the product and what it is not.
