---
title: The core must read a connector's answer up to 16 MiB, failing the call with RESOURCE_EXHAUSTED past it; must count the answer's messages and repeated elements on the wire before decoding it and refuse one of more than 640,000 with ErrDeliveryTooLarge; and must refuse a decoded delivery of more than 200,000 items, postings counted, before translating any of it.
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: The core must read a connector's answer up to 16 MiB, failing the call with RESOURCE_EXHAUSTED past it; must count the answer's messages and repeated elements on the wire before decoding it and refuse one of more than 640,000 with ErrDeliveryTooLarge; and must refuse a decoded delivery of more than 200,000 items, postings counted, before translating any of it.
logic_claims:
  - CLAIM-4425DC9262ED8418
semantic_clauses:
  - The core must read a connector's answer up to 16 MiB, failing the call with RESOURCE_EXHAUSTED past it; must count the answer's messages and repeated elements on the wire before decoding it and refuse one of more than 640,000 with ErrDeliveryTooLarge; and must refuse a decoded delivery of more than 200,000 items, postings counted, before translating any of it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f242b7bcc346deee5d185666c89ef53846f25f8fc9be0ed4fef6e10e56e3e3a7
semantic_inventory:
  - claim_key: CLAIM-4425DC9262ED8418
    claim_text: The core must read a connector's answer up to 16 MiB, failing the call with RESOURCE_EXHAUSTED past it; must count the answer's messages and repeated elements on the wire before decoding it and refuse one of more than 640,000 with ErrDeliveryTooLarge; and must refuse a decoded delivery of more than 200,000 items, postings counted, before translating any of it
    role: normative
    status: modeled
    span:
      start: 0
      end: 361
origin:
  kind: agent
  recorded_at: '2026-11-03T05:55:17.974Z'
id: REQ-ingest-connector-answer-counted-on-wire
type: req
---
The core must read a connector's answer up to 16 MiB, failing the call with RESOURCE_EXHAUSTED past it; must count the answer's messages and repeated elements on the wire before decoding it and refuse one of more than 640,000 with ErrDeliveryTooLarge; and must refuse a decoded delivery of more than 200,000 items, postings counted, before translating any of it.

## Context

An empty protobuf message costs two bytes on the wire and about 90 bytes decoded, so a byte cap alone does not bound what decoding allocates. Recorded with default firmness by the implementation agent of the second security review's fixes (fix/review2-r1-r4), from finding R1: checkDelivery counted items only after proto.Unmarshal, so 31 MiB of empty postings, within the old 32 MiB byte cap, decoded to 16.7 million messages and about 1.75 GB of heap in a core whose memory limit is 256 MiB. Measured on this branch: the byte cap filled with empty postings now grows the heap by about 31 MiB, the most empty messages the count lets through by about 63 MiB, and 200,000 valid prices decoded and translated by about 140 MiB. A larger history needs paged deliveries, which are not built.

## Why this changed

It supersedes REQ-ingest-connector-answer-bounded: that requirement counted a delivery's items only after decoding it, and its 32 MiB byte cap let a connector make the core decode millions of empty messages, far past the core's 256 MiB memory limit, so the caps were lowered to what the memory limit holds and the count moved before decoding.

## Source

> The core bounds every request body and uploaded file before authenticating or parsing it, and every parser bounds what it allocates, so no unauthenticated request or small file can exhaust its memory.

docs/DESIGN.md#D-184
