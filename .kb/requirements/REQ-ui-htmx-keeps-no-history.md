---
title: Every page must turn htmx's history cache off and refresh a page missing from the history, so htmx keeps no snapshot of a page in sessionStorage.
status: open
text_ref: docs/DESIGN.md#D-151
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-151
  - lane:strict
semantic_text: Every page must turn htmx's history cache off and refresh a page missing from the history, so htmx keeps no snapshot of a page in sessionStorage.
logic_claims:
  - CLAIM-4D0DE30346870BCF
semantic_clauses:
  - Every page must turn htmx's history cache off and refresh a page missing from the history, so htmx keeps no snapshot of a page in sessionStorage.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c4f4245368d3fa160ef9e53f0622c8dda659e214073b28f9675800442cee5873
semantic_inventory:
  - claim_key: CLAIM-4D0DE30346870BCF
    claim_text: Every page must turn htmx's history cache off and refresh a page missing from the history, so htmx keeps no snapshot of a page in sessionStorage
    role: normative
    status: modeled
    span:
      start: 0
      end: 144
origin:
  kind: agent
  recorded_at: '2026-11-03T16:31:05.962Z'
id: REQ-ui-htmx-keeps-no-history
type: req
---
Every page must turn htmx's history cache off and refresh a page missing from the history, so htmx keeps no snapshot of a page in sessionStorage.

## Context

With hx-push-url, htmx keeps snapshots of the pages it pushed into the history, the owner's record among them, in sessionStorage; D-151 never keeps transactions, lots or account numbers on the device. Ported before the alpha tag (P1g).

## Source

> Transactions, lots and account numbers are never kept.

docs/DESIGN.md#D-151
