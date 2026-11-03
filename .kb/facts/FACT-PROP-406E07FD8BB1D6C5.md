---
title: Text Safe For Terminals = true
status: active
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.manifest
property_key: text_safe_for_terminals
canonical_key: mcp-kibi-model-requirement:plugins.manifest:text_safe_for_terminals:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-6E4948DB14BF642A
claim_text: Manifest validation must refuse, with ErrUnsafeText, every manifest string, the config schema's keys and string values at any depth included, that holds a C0 or C1 control character, DEL, a Unicode bidirectional embedding, override or isolate, or invalid UTF-8.
origin:
  kind: agent
  recorded_at: '2026-11-03T09:18:15.537Z'
id: FACT-PROP-406E07FD8BB1D6C5
type: fact
---
