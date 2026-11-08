---
title: Resyncs Rekeyed Bond Idempotently = true
status: active
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.ghostfolio
property_key: resyncs_rekeyed_bond_idempotently
canonical_key: mcp-kibi-model-requirement:views.ghostfolio:resyncs_rekeyed_bond_idempotently:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-8B086CCEFA5B7584
claim_text: After a bond is re-keyed to its registry ISIN, the Ghostfolio view's next sync must import and delete nothing, since a bond's asset is derived from its series code and its activities name transaction keys.
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:31.474Z'
id: FACT-PROP-1A47B9813BA708DA
type: fact
---
