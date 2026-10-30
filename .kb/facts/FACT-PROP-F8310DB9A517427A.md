---
title: Mapped Entries Booked = true
status: active
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.broker_accounts
property_key: mapped_entries_booked
canonical_key: mcp-kibi-model-requirement:ingest.broker_accounts:mapped_entries_booked:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EAE5FFF3BA0B86BE
claim_text: Entries a connector names by broker account must be booked into the OpenFief account the broker account is mapped to, with ':' and that account ID appended to each transaction's key source, and nothing of a broker account that is not mapped may be booked.
origin:
  kind: agent
  recorded_at: '2026-10-29T11:36:50.781Z'
id: FACT-PROP-F8310DB9A517427A
type: fact
---
