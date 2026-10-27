---
title: Offline Behavior = warn locally, fail in CI
status: active
text_ref: docs/dev/sdlc.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.vulnerability_check
property_key: offline_behavior
canonical_key: docs-dev-sdlc-md:tooling.vulnerability_check:offline_behavior:eq:warn_locally_fail_in_ci
operator: eq
value_type: string
value_string: warn locally, fail in CI
claim_key: CLAIM-1C02253D43CC2675
claim_text: When a vulnerability database is unreachable, the vulnerability check must warn and pass locally, and must fail when CI is true.
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:46.154Z'
id: FACT-PROP-15CE2C278EFBB91D
type: fact
---
