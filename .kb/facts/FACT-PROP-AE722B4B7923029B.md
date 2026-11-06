---
title: Sleep Or Skip In Test = fail
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
property_key: sleep_or_skip_in_test
canonical_key: docs-dev-code-organization-md:tooling.archtest:sleep_or_skip_in_test:eq:fail
operator: eq
value_type: string
value_string: fail
claim_key: CLAIM-21ACD3A725C8AF19
claim_text: The architecture tests must fail when a test calls time.Sleep or skips itself, unless the exceptions list that test with a reason.
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:58.947Z'
id: FACT-PROP-AE722B4B7923029B
type: fact
---
