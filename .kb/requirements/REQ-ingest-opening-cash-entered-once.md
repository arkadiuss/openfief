---
title: The cash balance the owner enters with the opening upload of a statement that holds none must be booked once as the account's opening cash on its go-live date, and refused with any other upload, for an account already seeded, in another currency than the account's, or by a connector whose statement holds the cash.
status: open
text_ref: docs/DESIGN.md#D-136
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-136
  - lane:strict
semantic_text: The cash balance the owner enters with the opening upload of a statement that holds none must be booked once as the account's opening cash on its go-live date, and refused with any other upload, for an account already seeded, in another currency than the account's, or by a connector whose statement holds the cash.
logic_claims:
  - CLAIM-EF69FA5FF51D9578
semantic_clauses:
  - The cash balance the owner enters with the opening upload of a statement that holds none must be booked once as the account's opening cash on its go-live date, and refused with any other upload, for an account already seeded, in another currency than the account's, or by a connector whose statement holds the cash.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6fec4834eb6ffd7a5693dc60c3345a4ff785df067ad8b45b190e8eb64984287e
semantic_inventory:
  - claim_key: CLAIM-EF69FA5FF51D9578
    claim_text: The cash balance the owner enters with the opening upload of a statement that holds none must be booked once as the account's opening cash on its go-live date, and refused with any other upload, for an account already seeded, in another currency than the account's, or by a connector whose statement holds the cash
    role: normative
    status: modeled
    span:
      start: 0
      end: 314
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:52.302Z'
id: REQ-ingest-opening-cash-entered-once
type: req
---
The cash balance the owner enters with the opening upload of a statement that holds none must be booked once as the account's opening cash on its go-live date, and refused with any other upload, for an account already seeded, in another currency than the account's, or by a connector whose statement holds the cash.

## Context

D-136 has the owner enter an XTB account's cash balance once, since the export holds none (R-24); D-23 books it as opening cash at go-live, and D-7 keeps data entering through the connector. Recorded by the XTB start implementation agent (impl/xtb-start) for D-136; default firmness, not reviewed by the author yet; firm requirements for D-136 are to come from the design session.

## Source

> An XTB account starts from an export covering its whole history, or from the open positions of a current export, which seed one opening lot per position at its open time and price, with the cash balance the owner enters once.

docs/DESIGN.md#D-136
