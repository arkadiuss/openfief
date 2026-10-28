---
title: Sdk Go Import File
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: sdk.go.import_file
canonical_key: sdk.go.import_file
claim_key: CLAIM-A8E67C08803D1EF7
claim_text: The Go SDK's ReceiveFile must take exactly one header as the first message of an ImportFile stream, then the file's chunks in order, and must refuse a stream without a header, with a second header, or with more bytes than the connector's limit.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:38.128Z'
id: FACT-SUBJECT-967C451FC87BA375
type: fact
---
