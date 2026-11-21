---
title: Core Attention Items
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-130-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
text_ref: design-decisions:D-130#1
fact_kind: subject
subject_key: core.attention_items
canonical_key: core.attention_items
claim_key: CLAIM-C09B827CC1A61D1B
claim_text: Every check that needs the owner must raise an attention item in one append-only store that keeps its kind, subject, state (open, acknowledged or resolved) and history.
origin:
  kind: agent
  recorded_at: '2026-10-31T14:44:37.417Z'
id: FACT-SUBJECT-A9563048F0AE7021
type: fact
---

Definition not stated: the subject core.attention_items is named by requirements REQ-attention-unclassified-instrument-is-check, REQ-attention-unclassified-instrument-resolves-itself, REQ-core-attention-item-checks, REQ-core-attention-item-outbox-events, REQ-core-attention-item-store, REQ-views-unbound-attention-names-instance-only, REQ-views-unbound-attention-owner-only, which link to the subject fact FACT-SUBJECT-A9563048F0AE7021 through constrains. The people who own that source can say what it is in the product and what it is not.
