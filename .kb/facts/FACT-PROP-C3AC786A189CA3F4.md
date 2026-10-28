---
title: Bundled And Tuf Updated = true
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
subject_key: plugins.trust_root
property_key: bundled_and_tuf_updated
canonical_key: mcp-kibi-model-requirement:plugins.trust_root:bundled_and_tuf_updated:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-73CEC9C6009BDDAF
claim_text: The operator must verify signatures against a Sigstore trust root embedded in its binary, without network access to Sigstore services, and the embedded root must be refreshed only through a TUF-verified update procedure.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:17:18.599Z'
id: FACT-PROP-C3AC786A189CA3F4
type: fact
---
