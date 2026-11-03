---
title: Index Readings Need Declaration And Approval = true
status: active
text_ref: design-decisions:D-186#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-186-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.connector_delivery
property_key: index_readings_need_declaration_and_approval
canonical_key: mcp-kibi-model-requirement:ingest.connector_delivery:index_readings_need_declaration_and_approval:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0A97212A82AF36BD
claim_text: Only a connector whose signed manifest declares index readings, such as CPI and the NBP reference rate, and whose grant approves that may deliver them; any other connector's index readings must be dropped with an attention item, as D-179 drops bond terms.
origin:
  kind: agent
  recorded_at: '2026-11-03T10:07:40.038Z'
id: FACT-PROP-DAF108D8CB0AD9EC
type: fact
---
