---
title: Sdk Go Connector Server
status: archived
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: sdk_go.connector_server
canonical_key: sdk_go.connector_server
claim_key: CLAIM-CD00D9B65C32BD72
claim_text: The Go SDK's connector server must read a message of at most 4 MiB unless the plugin says otherwise, a body of at most 65 MiB, and bound how long a request may take to arrive and a connection may idle.
origin:
  kind: agent
  recorded_at: '2026-11-02T19:46:08.852Z'
id: FACT-SUBJECT-8986F598EDB94BD6
type: fact
part_of:
  - sdk.go.connector_server
---
The bounds the Go SDK's connector server puts on a message, a body and the time a request may take or a connection idle. They are a part of sdk.go.connector_server, which also covers refusing unauthenticated calls.