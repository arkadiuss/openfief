---
title: Deletes Leftover Assets After Import = true
status: active
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.ghostfolio_plugin
property_key: deletes_leftover_assets_after_import
canonical_key: mcp-kibi-model-requirement:views.ghostfolio_plugin:deletes_leftover_assets_after_import:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5909FE07AD300656
claim_text: A Ghostfolio sync must delete an activity on a MANUAL asset an earlier version left after the import, like any stale activity, and delete before the import only the stale activities an import duplicates.
origin:
  kind: agent
  recorded_at: '2026-10-31T15:07:56.487Z'
id: FACT-PROP-F38264BBE7387E71
type: fact
---
