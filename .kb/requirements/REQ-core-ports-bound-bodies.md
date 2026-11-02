---
title: 'Both core ports must refuse a compressed request body with 415 and cap every body by its path before the API behind it reads it: 2 MiB for the core''s own APIs and 65 MiB for a statement upload; every Connect service must read a message of at most 1 MiB; and both servers must bound how long a request may take to arrive and a connection may idle.'
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: 'Both core ports must refuse a compressed request body with 415 and cap every body by its path before the API behind it reads it: 2 MiB for the core''s own APIs and 65 MiB for a statement upload; every Connect service must read a message of at most 1 MiB; and both servers must bound how long a request may take to arrive and a connection may idle.'
logic_claims:
  - CLAIM-2413CF2990A49BB8
semantic_clauses:
  - 'Both core ports must refuse a compressed request body with 415 and cap every body by its path before the API behind it reads it: 2 MiB for the core''s own APIs and 65 MiB for a statement upload; every Connect service must read a message of at most 1 MiB; and both servers must bound how long a request may take to arrive and a connection may idle.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a195576c394410a1b3083f6b06a6ae1a6d9de1bf37ce8d9b36279731a9d700ac
semantic_inventory:
  - claim_key: CLAIM-2413CF2990A49BB8
    claim_text: 'Both core ports must refuse a compressed request body with 415 and cap every body by its path before the API behind it reads it: 2 MiB for the core''s own APIs and 65 MiB for a statement upload; every Connect service must read a message of at most 1 MiB; and both servers must bound how long a request may take to arrive and a connection may idle'
    role: normative
    status: modeled
    span:
      start: 0
      end: 345
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:56.767Z'
id: REQ-core-ports-bound-bodies
type: req
---
Both core ports must refuse a compressed request body with 415 and cap every body by its path before the API behind it reads it: 2 MiB for the core's own APIs and 65 MiB for a statement upload; every Connect service must read a message of at most 1 MiB; and both servers must bound how long a request may take to arrive and a connection may idle.

## Context

A statement upload is 64 MiB at most (DefaultMaxUploadBytes), sent by the browser as a form the front door proxies to the view plugin and by the view plugin as a stream of 64 KiB chunks; the built-in UI gives it five minutes, which is the read timeout. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p0-dos), from the security review's finding P0-2: a ~400 KB gzip body to ReadService/ListAccounts allocated about 713 MiB in the core before its 401, because no Connect handler capped what it read and authentication runs only after the message is decoded. The author made the fix a decision, D-184. No firm requirement for D-184 existed in kibi on this branch when it was built.

## Source

> The core bounds every request body and uploaded file before authenticating or parsing it, and every parser bounds what it allocates, so no unauthenticated request or small file can exhaust its memory.

docs/DESIGN.md#D-184
