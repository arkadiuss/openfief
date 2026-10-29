---
title: Trusts The Operators Signers = true
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
subject_key: plugins.install_cli
property_key: trusts_the_operators_signers
canonical_key: mcp-kibi-model-requirement:plugins.install_cli:trusts_the_operators_signers:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-92EF0BDBAE01BB44
claim_text: 'The CLI must trust exactly the signers the operator trusts: OpenFief''s CI, the catalog''s publishers and the owner''s signers from the openfief-trusted-signers ConfigMap (or a file the owner names), and an invalid signer document must stop it.'
origin:
  kind: agent
  recorded_at: '2026-10-28T23:14:11.413Z'
id: FACT-PROP-2E7F4C4A82E3888D
type: fact
---
