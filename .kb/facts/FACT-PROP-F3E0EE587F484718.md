---
title: Sends Go Live Only When Seeded = true
status: active
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.bond_lot_entry
property_key: sends_go_live_only_when_seeded
canonical_key: mcp-kibi-model-requirement:ingest.bond_lot_entry:sends_go_live_only_when_seeded:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-BBF2D4BFEAC92A88
claim_text: The core must send no go-live date with the bond lots of an account never seeded with opening lots, and an account seeded with them its seeding date.
origin:
  kind: agent
  recorded_at: '2026-11-01T07:46:08.972Z'
id: FACT-PROP-F3E0EE587F484718
type: fact
---
