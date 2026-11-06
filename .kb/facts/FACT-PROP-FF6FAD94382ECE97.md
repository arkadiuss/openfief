---
title: Sensitive Value Logged = fail
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
property_key: sensitive_value_logged
canonical_key: docs-dev-code-organization-md:tooling.archtest:sensitive_value_logged:eq:fail
operator: eq
value_type: string
value_string: fail
claim_key: CLAIM-F41C06A280383C4C
claim_text: The architecture tests must fail when a logging call takes a value of a money, quantity, decimal, price, FX rate, ISIN or script token type, or an error that is not passed through a LogSafe function.
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:55.673Z'
id: FACT-PROP-FF6FAD94382ECE97
type: fact
---
