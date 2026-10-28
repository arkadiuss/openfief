---
title: Sdk Serve Connector
status: active
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: sdk.serve_connector
canonical_key: sdk.serve_connector
claim_key: CLAIM-396DD329FED54C85
claim_text: The Go SDK's ServeConnector must serve a connector's contract on the plugin's listen port only to callers its core verifier accepts, refuse to start without its listen port or instance name, and on cancellation stop taking calls and let calls in flight finish within a bound.
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:11.362Z'
id: FACT-SUBJECT-D9990F90A0163DCF
type: fact
---
