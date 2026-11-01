---
title: The built-in UI must ask, for an account that hasn't started, how it starts, from the holdings of a current statement with the cash balance entered once or from a statement of its whole history, and send the cash with the upload.
status: open
text_ref: docs/DESIGN.md#D-136
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-136
  - lane:strict
semantic_text: The built-in UI must ask, for an account that hasn't started, how it starts, from the holdings of a current statement with the cash balance entered once or from a statement of its whole history, and send the cash with the upload.
logic_claims:
  - CLAIM-91E97C77A5F378FE
semantic_clauses:
  - The built-in UI must ask, for an account that hasn't started, how it starts, from the holdings of a current statement with the cash balance entered once or from a statement of its whole history, and send the cash with the upload.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e41fcf4bbe67ca6efd69e3506ddac11708076451946057d82476ea2f7232a140
semantic_inventory:
  - claim_key: CLAIM-91E97C77A5F378FE
    claim_text: The built-in UI must ask, for an account that hasn't started, how it starts, from the holdings of a current statement with the cash balance entered once or from a statement of its whole history, and send the cash with the upload
    role: normative
    status: modeled
    span:
      start: 0
      end: 228
origin:
  kind: agent
  recorded_at: '2026-11-01T07:36:15.303Z'
id: REQ-ui-statement-upload-start
type: req
---
The built-in UI must ask, for an account that hasn't started, how it starts, from the holdings of a current statement with the cash balance entered once or from a statement of its whole history, and send the cash with the upload.

## Context

D-136 gives an XTB account two starts and has the owner enter its cash once; the work package asked for the built-in UI to choose the start and take the cash. Recorded by the XTB start implementation agent (impl/xtb-start) for D-136; default firmness, not reviewed by the author yet; firm requirements for D-136 are to come from the design session.

## Source

> An XTB account starts from an export covering its whole history, or from the open positions of a current export, which seed one opening lot per position at its open time and price, with the cash balance the owner enters once.

docs/DESIGN.md#D-136
