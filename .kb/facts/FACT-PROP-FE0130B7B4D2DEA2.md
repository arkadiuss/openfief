---
title: Checks Only Affected Packages = true
status: active
text_ref: docs/dev/sdlc.md#tooling
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-tooling
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.pre_commit_hook
property_key: checks_only_affected_packages
canonical_key: mcp-kibi-model-requirement:tooling.pre_commit_hook:checks_only_affected_packages:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EAD1408AC02201D4
claim_text: 'The pre-commit hook must lint, vulnerability-check and test only the Go packages a staged change affects: the changed packages and every package that imports them directly or not, plus the packages whose tests import one of them, found from Go''s import graph.'
origin:
  kind: agent
  recorded_at: '2026-10-30T03:33:20.573Z'
id: FACT-PROP-FE0130B7B4D2DEA2
type: fact
---
