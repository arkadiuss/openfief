---
title: Applies Only On Confirmation = true
status: active
text_ref: docs/DESIGN.md#D-90
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-90
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.cli_upgrade
property_key: applies_only_on_confirmation
canonical_key: mcp-kibi-model-requirement:plugins.cli_upgrade:applies_only_on_confirmation:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C9034985B4EB49BC
claim_text: openfief plugin upgrade must change nothing unless the owner confirms, and on confirmation replace the Plugin's spec before its grant's, granting the operator read access to a new in-cluster service first and taking it back from one no longer named.
origin:
  kind: agent
  recorded_at: '2026-10-29T17:04:19.035Z'
id: FACT-PROP-79E056D05DB28B90
type: fact
---
