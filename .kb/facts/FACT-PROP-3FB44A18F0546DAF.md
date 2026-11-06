---
title: Float Or Impure Domain = fail
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
property_key: float_or_impure_domain
canonical_key: docs-dev-code-organization-md:tooling.archtest:float_or_impure_domain:eq:fail
operator: eq
value_type: string
value_string: fail
claim_key: CLAIM-0EF278FC493C9404
claim_text: The architecture tests must fail when a domain, app or kernel package uses float32 or float64, when a contract under proto/ has a float or double field, or when a domain or kernel package starts a goroutine or holds a package-level variable other than an error.
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:52.294Z'
id: FACT-PROP-3FB44A18F0546DAF
type: fact
---
