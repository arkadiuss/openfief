---
title: Calls Ghostfolio Safely = true
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
property_key: calls_ghostfolio_safely
canonical_key: mcp-kibi-model-requirement:views.ghostfolio_plugin:calls_ghostfolio_safely:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-698921AEF5C0BEB1
claim_text: The Ghostfolio client must log in by exchanging the security token, read from its Secret on every login, for a bearer token; must refuse to list or bulk-delete activities without an account filter, sending nothing; must count deleting an activity Ghostfolio no longer has as done; must send numbers as JSON numbers with their exact digits and dates as midnight UTC; and must report an error status by call and status only.
origin:
  kind: agent
  recorded_at: '2026-10-29T00:55:11.911Z'
id: FACT-PROP-80F5723BEFC82156
type: fact
---
