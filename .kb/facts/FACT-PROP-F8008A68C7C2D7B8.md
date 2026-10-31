---
title: Serves Every Bond Account = true
status: active
text_ref: docs/DESIGN.md#D-118
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-118
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.treasury_bonds
property_key: serves_every_bond_account
canonical_key: mcp-kibi-model-requirement:plugins.treasury_bonds:serves_every_bond_account:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0C67A9ED96FA810F
claim_text: The core must hand each bond lot to the treasury-bonds connector with the account the owner entered it on, the connector must book it there, and the plugin's config must name no account.
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:48.369Z'
id: FACT-PROP-F8008A68C7C2D7B8
type: fact
---
