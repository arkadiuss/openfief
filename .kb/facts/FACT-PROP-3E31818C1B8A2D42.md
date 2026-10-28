---
title: Reassembles Bounded File = true
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
subject_key: sdk.go.import_file
property_key: reassembles_bounded_file
canonical_key: mcp-kibi-model-requirement:sdk.go.import_file:reassembles_bounded_file:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A8E67C08803D1EF7
claim_text: The Go SDK's ReceiveFile must take exactly one header as the first message of an ImportFile stream, then the file's chunks in order, and must refuse a stream without a header, with a second header, or with more bytes than the connector's limit.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:43.598Z'
id: FACT-PROP-3E31818C1B8A2D42
type: fact
---
