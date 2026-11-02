---
title: The read API must refuse a compressed request body with 415 and cap every other body and message at 1 MiB, answering 413 or resource exhausted, before it checks the caller's token or decodes the message.
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: The read API must refuse a compressed request body with 415 and cap every other body and message at 1 MiB, answering 413 or resource exhausted, before it checks the caller's token or decodes the message.
logic_claims:
  - CLAIM-5F9952885EB49774
semantic_clauses:
  - The read API must refuse a compressed request body with 415 and cap every other body and message at 1 MiB, answering 413 or resource exhausted, before it checks the caller's token or decodes the message.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ba12a52f7059991f7bc5a37064e416a8f598c5d9d851de308478c0ec07235e8d
semantic_inventory:
  - claim_key: CLAIM-5F9952885EB49774
    claim_text: The read API must refuse a compressed request body with 415 and cap every other body and message at 1 MiB, answering 413 or resource exhausted, before it checks the caller's token or decodes the message
    role: normative
    status: modeled
    span:
      start: 0
      end: 202
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:48.588Z'
id: REQ-views-read-api-bounded-bodies
type: req
---
The read API must refuse a compressed request body with 415 and cap every other body and message at 1 MiB, answering 413 or resource exhausted, before it checks the caller's token or decodes the message.

## Context

Vanguard inflates a compressed REST body without a limit, and Connect decodes the message before the authentication interceptor runs, so the bounds must come first. The read API's requests carry only a few IDs, dates and a page token. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p0-dos), from the security review's finding P0-2: a ~400 KB gzip body to ReadService/ListAccounts allocated about 713 MiB in the core before its 401, because no Connect handler capped what it read and authentication runs only after the message is decoded. The author made the fix a decision, D-184. No firm requirement for D-184 existed in kibi on this branch when it was built.

## Source

> The core bounds every request body and uploaded file before authenticating or parsing it, and every parser bounds what it allocates, so no unauthenticated request or small file can exhaust its memory.

docs/DESIGN.md#D-184
