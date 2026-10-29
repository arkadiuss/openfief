---
title: Prices Requested Instruments = true
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
subject_key: plugins.eodhd
property_key: prices_requested_instruments
canonical_key: mcp-kibi-model-requirement:plugins.eodhd:prices_requested_instruments:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E76BDF40795B8125
claim_text: The EODHD plugin must price each instrument a sync request names, delivering its closes under the symbol the request gives as well as a watched entry's, and must fetch each requested market index as the ticker of its own name plus .INDX unless its config names the ticker, reporting a name that can't be a ticker as an ingest problem.
origin:
  kind: agent
  recorded_at: '2026-10-29T16:58:09.257Z'
id: FACT-PROP-3781BDC39E2D8EEF
type: fact
---
