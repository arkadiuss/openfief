---
title: Sync Scope Held And Targets = true
status: active
text_ref: docs/DESIGN.md#D-93
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-93
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.market_data
property_key: sync_scope_held_and_targets
canonical_key: mcp-kibi-model-requirement:ingest.market_data:sync_scope_held_and_targets:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EF14AC4F99088435
claim_text: The core must ask every connector instance bound to no account to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask an instance bound to an account for nothing.
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:58.626Z'
id: FACT-PROP-F09792EDC8A6B33A
type: fact
---
