---
title: Sync Step Order = delete stale activities an import would match as duplicates; import missing activities; delete other stale activities
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-64#1
fact_kind: property_value
subject_key: views.ghostfolio_plugin
property_key: sync_step_order
operator: eq
value_type: string
value_string: delete stale activities an import would match as duplicates; import missing activities; delete other stale activities
canonical_key: docs-design-md:views.ghostfolio_plugin:sync_step_order:eq:delete_stale_activities_an_import_would_match_as_duplicates_import_missing_activities_delete_other_stale_activities
claim_key: CLAIM-96669E061B9B47A2
claim_text: Each Ghostfolio sync must first delete the stale activities that Ghostfolio's duplicate check would match against an import, then import missing activities, and only then delete the other stale ones.
origin:
  kind: agent
  recorded_at: '2026-10-29T10:21:47.902Z'
id: FACT-PROP-38562D43BEAB63D8
type: fact
---
