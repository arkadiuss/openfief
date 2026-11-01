---
title: An opening upload must record its account's go-live at once, even when nothing can be booked yet, so later uploads are transactions and waiting positions join the opening on that date.
status: open
text_ref: docs/DESIGN.md#D-160
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-160
  - lane:strict
semantic_text: An opening upload must record its account's go-live at once, even when nothing can be booked yet, so later uploads are transactions and waiting positions join the opening on that date.
logic_claims:
  - CLAIM-D24A33B5E8134AAC
semantic_clauses:
  - An opening upload must record its account's go-live at once, even when nothing can be booked yet, so later uploads are transactions and waiting positions join the opening on that date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3a5516c017fdf22cfc4976128714e2974137022ae4a559b92b3e9fd633e30f49
semantic_inventory:
  - claim_key: CLAIM-D24A33B5E8134AAC
    claim_text: An opening upload must record its account's go-live at once, even when nothing can be booked yet, so later uploads are transactions and waiting positions join the opening on that date
    role: normative
    status: modeled
    span:
      start: 0
      end: 183
origin:
  kind: agent
  recorded_at: '2026-11-01T16:50:56.599Z'
id: REQ-ingest-opening-upload-starts-account
type: req
---
An opening upload must record its account's go-live at once, even when nothing can be booked yet, so later uploads are transactions and waiting positions join the opening on that date.

## Context

D-160 starts an account with its opening upload; before it, an account whose positions all waited had no go-live and could reach a dead end when later rows were booked before its opening lots. Recorded by the implementation agent of D-159 to D-162 (impl/d159-162) with default firmness: no firm requirement for these decisions existed in kibi when they were built, so this traces the code to the decision directly until the design session imports the firm ones.

## Source

> An opening upload starts its account at once: the account's go-live is recorded with the upload, even when nothing can be booked yet. Positions waiting for their ticker's ISIN (D-116) join the opening when the owner confirms it, and later uploads are imported as transactions.

docs/DESIGN.md#D-160
