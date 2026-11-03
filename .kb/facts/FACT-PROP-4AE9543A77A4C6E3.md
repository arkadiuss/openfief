---
title: Event Subject Read Form = true
status: active
text_ref: docs/DESIGN.md#D-121
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-121
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: attention.outbox
property_key: event_subject_read_form
canonical_key: mcp-kibi-model-requirement:attention.outbox:event_subject_read_form:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C5256BA3B0044FAA
claim_text: An attention item's outbox event must carry its subject as read access sees it (attention.ReadSubject), so no outbox row holds a broker account number; the item keeps its full subject for the owner.
origin:
  kind: agent
  recorded_at: '2026-11-03T15:54:54.400Z'
id: FACT-PROP-4AE9543A77A4C6E3
type: fact
---
