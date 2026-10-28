---
title: Records Trust Decision = true
status: active
text_ref: docs/DESIGN.md#D-56
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-56
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: operator.policy_verifier
property_key: records_trust_decision
canonical_key: mcp-kibi-model-requirement:operator.policy_verifier:records_trust_decision:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-80A3EB5FAC5DCC7A
claim_text: 'The operator must decide every plugin image with the trust policy and record the outcome in the Plugin''s SignatureVerified condition: True with reason Verified or SignedByTrustedSigner, False with the refusal reason, or Unknown when the registry could not be read, which is retried.'
origin:
  kind: agent
  recorded_at: '2026-10-28T09:29:33.794Z'
id: FACT-PROP-6C9952AA7F21A722
type: fact
---
