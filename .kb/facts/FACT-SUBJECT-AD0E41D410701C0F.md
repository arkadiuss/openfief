---
title: Tooling Pre Commit Hook
status: active
text_ref: docs/dev/sdlc.md#tooling
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-tooling
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: tooling.pre_commit_hook
canonical_key: tooling.pre_commit_hook
claim_key: CLAIM-EAD1408AC02201D4
claim_text: 'The pre-commit hook must lint, vulnerability-check and test only the Go packages a staged change affects: the changed packages and every package that imports them directly or not, plus the packages whose tests import one of them, found from Go''s import graph.'
origin:
  kind: agent
  recorded_at: '2026-10-30T03:33:15.538Z'
id: FACT-SUBJECT-AD0E41D410701C0F
type: fact
---
