---
title: The core must read a connector's answer up to 32 MiB, failing the call with RESOURCE_EXHAUSTED past it, and must refuse a delivery of more than 500,000 items, postings counted, before translating any of it.
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: The core must read a connector's answer up to 32 MiB, failing the call with RESOURCE_EXHAUSTED past it, and must refuse a delivery of more than 500,000 items, postings counted, before translating any of it.
logic_claims:
  - CLAIM-9E2E735844C765DF
semantic_clauses:
  - The core must read a connector's answer up to 32 MiB, failing the call with RESOURCE_EXHAUSTED past it, and must refuse a delivery of more than 500,000 items, postings counted, before translating any of it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: deb79312caed140749b007d56fa5cc495a3b9c2c81bc331c5628f86d85366d32
semantic_inventory:
  - claim_key: CLAIM-9E2E735844C765DF
    claim_text: The core must read a connector's answer up to 32 MiB, failing the call with RESOURCE_EXHAUSTED past it, and must refuse a delivery of more than 500,000 items, postings counted, before translating any of it
    role: normative
    status: modeled
    span:
      start: 0
      end: 205
origin:
  kind: agent
  recorded_at: '2026-11-02T19:46:05.405Z'
id: REQ-ingest-connector-answer-bounded
type: req
---
The core must read a connector's answer up to 32 MiB, failing the call with RESOURCE_EXHAUSTED past it, and must refuse a delivery of more than 500,000 items, postings counted, before translating any of it.

## Context

A delivery from the largest statement the core takes, 64 MiB, or a market-data backfill of 2,000 instrument-years of daily prices stays within both caps; past them a connector, even a broken one, could make the core decode and translate without end. The refusal names counts, never values (GR-12). Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p0-dos), from the security review's finding P0-2: a ~400 KB gzip body to ReadService/ListAccounts allocated about 713 MiB in the core before its 401, because no Connect handler capped what it read and authentication runs only after the message is decoded. The author made the fix a decision, D-184. No firm requirement for D-184 existed in kibi on this branch when it was built.

## Source

> The core bounds every request body and uploaded file before authenticating or parsing it, and every parser bounds what it allocates, so no unauthenticated request or small file can exhaust its memory.

docs/DESIGN.md#D-184

> GR-12: never log amounts, holdings, account numbers or tokens (docs/DESIGN.md#GR-12, docs/dev/code-organization.md#rules).
