---
title: Sends Unfiltered Bulk Delete = false
status: active
text_ref: design-decisions:D-64#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-2
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.ghostfolio_plugin
property_key: sends_unfiltered_bulk_delete
canonical_key: docs-design-md:views.ghostfolio_plugin:sends_unfiltered_bulk_delete:eq:false
operator: eq
value_type: bool
value_bool: false
claim_key: CLAIM-6883370B156E4528
claim_text: The Ghostfolio plugin must never send a bulk delete without an account filter.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:18:54.388Z'
id: FACT-PROP-8AE99A2D4FA6E1B9
type: fact
---
