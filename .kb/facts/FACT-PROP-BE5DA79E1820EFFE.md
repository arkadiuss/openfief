---
title: Handed To Bound Connector = true
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.bond_lot_entry
property_key: handed_to_bound_connector
canonical_key: mcp-kibi-model-requirement:ingest.bond_lot_entry:handed_to_bound_connector:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-6AAC65AA02FC492F
claim_text: The core must take a bond lot only from the owner's owner-scoped session, hand it to the one connector instance bound to its account as a new transactions batch, and import what that instance delivers into that account only, recording the delivered series terms.
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:25.109Z'
id: FACT-PROP-BE5DA79E1820EFFE
type: fact
---
