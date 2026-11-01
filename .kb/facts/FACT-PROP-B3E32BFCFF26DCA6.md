---
title: Entered Once By Owner = true
status: active
text_ref: docs/DESIGN.md#D-136
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-136
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.opening_cash
property_key: entered_once_by_owner
canonical_key: mcp-kibi-model-requirement:ingest.opening_cash:entered_once_by_owner:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EF69FA5FF51D9578
claim_text: The cash balance the owner enters with the opening upload of a statement that holds none must be booked once as the account's opening cash on its go-live date, and refused with any other upload, for an account already seeded, in another currency than the account's, or by a connector whose statement holds the cash.
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:49.801Z'
id: FACT-PROP-B3E32BFCFF26DCA6
type: fact
---
