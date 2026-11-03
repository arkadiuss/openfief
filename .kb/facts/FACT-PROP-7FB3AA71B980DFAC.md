---
title: Declared Institution Only = true
status: active
text_ref: docs/DESIGN.md#D-182
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-182
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.deliveries
property_key: declared_institution_only
canonical_key: mcp-kibi-model-requirement:ingest.deliveries:declared_institution_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C0B7D02B561F51CD
claim_text: A connector's signed manifest must declare the institutions it books for, and the core must refuse, whole, a delivery that names an institution the running version's manifest doesn't declare, importing and storing nothing of it.
origin:
  kind: agent
  recorded_at: '2026-11-02T22:20:01.690Z'
id: FACT-PROP-7FB3AA71B980DFAC
type: fact
---
