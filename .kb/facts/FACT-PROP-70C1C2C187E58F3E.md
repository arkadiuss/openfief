---
title: Undeclared Cross Check Item Per Instance = true
status: active
text_ref: docs/DESIGN.md#D-141
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-141
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.connector_delivery
property_key: undeclared_cross_check_item_per_instance
canonical_key: mcp-kibi-model-requirement:ingest.connector_delivery:undeclared_cross_check_item_per_instance:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-852CA6F2C6F66554
claim_text: Undeclared cross-check prices must raise one attention item per connector instance, with a detail that names no count or batch, and the instance's next delivery without undeclared ones must clear it.
origin:
  kind: agent
  recorded_at: '2026-11-02T13:17:46.870Z'
id: FACT-PROP-70C1C2C187E58F3E
type: fact
---
