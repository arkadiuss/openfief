---
title: Reimport Is Noop = true
status: active
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.import
property_key: reimport_is_noop
canonical_key: mcp-kibi-model-requirement:ledger.import:reimport_is_noop:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5913080FD43262D2
claim_text: Re-importing a transaction whose source and institution ID are already in the log with the same content as its current version must change nothing, and content equality must not depend on posting order or on how a number is written.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:32:18.005Z'
id: FACT-PROP-F9CC4203228FD890
type: fact
---
