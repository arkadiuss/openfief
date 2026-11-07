---
title: Rekeys Provisional Instruments = true
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
subject_key: ingest.delivery
property_key: rekeys_provisional_instruments
canonical_key: mcp-kibi-model-requirement:ingest.delivery:rekeys_provisional_instruments:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-820480C3199862C6
claim_text: The core must re-key each provisional instrument a sync brings the registry ISIN of, after importing the rest of the delivery, in every context that stores it with the ledger last, idempotently, refusing an invalid registry ISIN and dropping one from a connector that may not deliver bond terms.
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:15.419Z'
id: FACT-PROP-DFC5F05A9976FCE7
type: fact
---
