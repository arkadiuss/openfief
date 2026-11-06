---
title: Baseline Not Exact = fail
status: active
text_ref: docs/dev/code-organization.md#architecture-tests
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-architecture-tests
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.archtest
property_key: baseline_not_exact
canonical_key: docs-dev-code-organization-md:tooling.archtest:baseline_not_exact:eq:fail
operator: eq
value_type: string
value_string: fail
claim_key: CLAIM-E35CA450D9E0E5E1
claim_text: The architecture tests must fail on any violation that neither the baseline nor the exceptions list, and on any listed line that no longer matches a violation, so the baseline only shrinks and both lists stay exact.
origin:
  kind: agent
  recorded_at: '2026-11-06T17:50:05.509Z'
id: FACT-PROP-EC5AACDF9AFA2A34
type: fact
---
