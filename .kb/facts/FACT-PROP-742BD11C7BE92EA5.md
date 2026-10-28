---
title: Matches Exactly = true
status: active
text_ref: docs/DESIGN.md#D-45
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-45
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.signer_identity
property_key: matches_exactly
canonical_key: mcp-kibi-model-requirement:plugins.signer_identity:matches_exactly:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-37A68FC72DBF9EAF
claim_text: A signature must count as a trusted signer's only when its certificate has the signer's exact OIDC issuer and either the signer's exact subject or a subject that the signer's pattern matches in full, or when the signer's public key verified it.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:51:18.739Z'
id: FACT-PROP-742BD11C7BE92EA5
type: fact
---
