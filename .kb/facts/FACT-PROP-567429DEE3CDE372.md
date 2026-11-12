---
title: Transient Failure = retried_with_backoff
status: active
text_ref: docs/dev/sdlc.md#gate-tiers
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-gate-tiers
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.downloads
property_key: transient_failure
canonical_key: docs-dev-sdlc-md:tooling.downloads:transient_failure:eq:retried_with_backoff
operator: eq
value_type: string
value_string: retried_with_backoff
claim_key: CLAIM-A4193AACD985DF55
claim_text: Downloads CI makes of its own, tool installs, test assets, release manifests and the cosign installer included, must be retried with backoff when the server answers 429 or 5xx or the connection fails, must never be retried on any other 4xx answer, and must keep their checksum checks as strict as before.
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:56.989Z'
id: FACT-PROP-567429DEE3CDE372
type: fact
---
