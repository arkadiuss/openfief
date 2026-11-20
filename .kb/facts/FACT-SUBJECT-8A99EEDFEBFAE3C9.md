---
title: Views Balance Check
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: views.balance_check
canonical_key: views.balance_check
claim_key: CLAIM-206B121C6E083E44
claim_text: 'The owner''s balance check must show, for each balance assertion, the reported balance, the balance derived from the transaction log, reported minus derived, and whether it passes: a cash assertion within 0.01 of its currency either way, a holding assertion only with equal units.'
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:18.750Z'
id: FACT-SUBJECT-8A99EEDFEBFAE3C9
type: fact
---

Definition not stated: the subject views.balance_check is named by requirements REQ-views-balance-check-owner-only, REQ-views-balance-check-tolerance, REQ-views-balance-check-unsupported, REQ-views-cli-check-balances, which link to the subject fact FACT-SUBJECT-8A99EEDFEBFAE3C9 through constrains. The people who own that source can say what it is in the product and what it is not.
