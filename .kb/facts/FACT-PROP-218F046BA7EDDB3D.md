---
title: Only On Fresh Install = true
status: active
text_ref: docs/DESIGN.md#D-71
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-71
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: portfolio.default_sleeves
property_key: only_on_fresh_install
canonical_key: mcp-kibi-model-requirement:portfolio.default_sleeves:only_on_fresh_install:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-572B3C5D2E082FC7
claim_text: Ensuring the defaults must store every missing built-in sleeve kind and, only while no sleeve is stored at all, the default sleeves Bonds, Core and Satellite, in one database transaction, and must never change a stored kind or sleeve.
origin:
  kind: agent
  recorded_at: '2026-10-29T11:36:23.854Z'
id: FACT-PROP-218F046BA7EDDB3D
type: fact
---
