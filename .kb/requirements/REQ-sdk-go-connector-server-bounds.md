---
title: The Go SDK's connector server must read a message of at most 4 MiB unless the plugin says otherwise, a body of at most 65 MiB, and bound how long a request may take to arrive and a connection may idle.
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: The Go SDK's connector server must read a message of at most 4 MiB unless the plugin says otherwise, a body of at most 65 MiB, and bound how long a request may take to arrive and a connection may idle.
logic_claims:
  - CLAIM-CD00D9B65C32BD72
semantic_clauses:
  - The Go SDK's connector server must read a message of at most 4 MiB unless the plugin says otherwise, a body of at most 65 MiB, and bound how long a request may take to arrive and a connection may idle.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9ac7af2121f59ccc7ee7cece804957870b4fd117b982f90204caf0c1c5dd545a
semantic_inventory:
  - claim_key: CLAIM-CD00D9B65C32BD72
    claim_text: The Go SDK's connector server must read a message of at most 4 MiB unless the plugin says otherwise, a body of at most 65 MiB, and bound how long a request may take to arrive and a connection may idle
    role: exception
    status: modeled
    span:
      start: 0
      end: 200
origin:
  kind: agent
  recorded_at: '2026-11-02T19:46:13.230Z'
id: REQ-sdk-go-connector-server-bounds
type: req
---
The Go SDK's connector server must read a message of at most 4 MiB unless the plugin says otherwise, a body of at most 65 MiB, and bound how long a request may take to arrive and a connection may idle.

## Context

The core's sync requests and bond lots are far smaller, and a statement reaches a plugin in 64 KiB chunks of a file of at most 64 MiB. Only the core passes RequireCore, so this guards the plugin against a broken core as much as against a hostile caller. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p0-dos), from the security review's finding P0-2: a ~400 KB gzip body to ReadService/ListAccounts allocated about 713 MiB in the core before its 401, because no Connect handler capped what it read and authentication runs only after the message is decoded. The author made the fix a decision, D-184. No firm requirement for D-184 existed in kibi on this branch when it was built.

## Source

> The core bounds every request body and uploaded file before authenticating or parsing it, and every parser bounds what it allocates, so no unauthenticated request or small file can exhaust its memory.

docs/DESIGN.md#D-184
