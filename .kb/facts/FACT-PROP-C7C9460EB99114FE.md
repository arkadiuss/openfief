---
title: Checked Before Connector Reads = true
status: active
text_ref: docs/DESIGN.md#D-109
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-109
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.statement_upload
property_key: checked_before_connector_reads
canonical_key: mcp-kibi-model-requirement:ingest.statement_upload:checked_before_connector_reads:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-76780D141DF12E32
claim_text: An uploaded statement must be refused before a connector reads it when the caller is not the owner or the file is empty, and the connector must get only the cleaned last element of its name and a sniffed media type.
origin:
  kind: agent
  recorded_at: '2026-10-30T04:11:09.218Z'
id: FACT-PROP-C7C9460EB99114FE
type: fact
---
