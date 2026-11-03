---
title: Names Never Shared = true
status: active
text_ref: docs/DESIGN.md#D-89
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-89
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.secrets
property_key: names_never_shared
canonical_key: mcp-kibi-model-requirement:plugins.secrets:names_never_shared:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3CCC62ACDA78AE97
claim_text: The CLI must refuse an install or an upgrade, prepared or applied, whose grant names a Secret another installed instance names, and the operator must not run a plugin version that would mount a Secret another instance mounts, unless it mounts that Secret too and is the older Plugin.
origin:
  kind: agent
  recorded_at: '2026-11-03T08:22:15.926Z'
id: FACT-PROP-1FFFBC87388E7998
type: fact
---
