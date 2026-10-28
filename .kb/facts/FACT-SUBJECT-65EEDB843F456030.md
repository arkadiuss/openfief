---
title: Sdk Go Plugin Token
status: active
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: sdk.go.plugin_token
canonical_key: sdk.go.plugin_token
claim_key: CLAIM-4779E08C64B66221
claim_text: The Go SDK must read a plugin's projected ServiceAccount token from its file on every call that presents it, never caching it, and must fail the call without sending anything when the file is missing or empty.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:02:09.336Z'
id: FACT-SUBJECT-65EEDB843F456030
type: fact
---
