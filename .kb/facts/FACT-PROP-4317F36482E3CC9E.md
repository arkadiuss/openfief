---
title: Verifies Signature Before Deploy Offline = true
status: active
text_ref: design-decisions:D-56#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-56-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.operator
property_key: verifies_signature_before_deploy_offline
canonical_key: docs-design-md:plugins.operator:verifies_signature_before_deploy_offline:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-FD6E0F61B0D0C94C
claim_text: The operator must verify every plugin image's signature before deploying it, including a Plugin applied without the CLI, and verification must work offline.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:22.636Z'
id: FACT-PROP-4317F36482E3CC9E
type: fact
---
