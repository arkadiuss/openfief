---
title: Untrusted Signer Named = true
status: active
text_ref: design-decisions:D-169#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-169-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.signature_trust
property_key: untrusted_signer_named
canonical_key: mcp-kibi-model-requirement:plugins.signature_trust:untrusted_signer_named:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-D1B8E9C11D1FEDD6
claim_text: When the operator refuses an image signed only by signers it doesn't trust, the Plugin's status must name each signer's identity, distinct from an image with no valid signature.
origin:
  kind: agent
  recorded_at: '2026-11-01T21:51:20.587Z'
id: FACT-PROP-A72D820C82F478EF
type: fact
---
