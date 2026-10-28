---
title: Secret Printed Once = true
status: active
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: core.token_commands
property_key: secret_printed_once
canonical_key: mcp-kibi-model-requirement:core.token_commands:secret_printed_once:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3A844105AFEAAD85
claim_text: core token issue must print a new token's secret once, as the first line on stdout, and log nothing about it; core token list must show every token's ID, name, scope, times and status and never a secret; core token revoke must revoke the token with the given ID.
origin:
  kind: agent
  recorded_at: '2026-10-28T14:33:23.192Z'
id: FACT-PROP-61C320A8FE33CBCB
type: fact
---
