---
title: 'An account that holds transactions but isn''t seeded started from a statement of its whole history: the core must choose transactions mode for its uploads and refuse an opening upload for it.'
status: open
text_ref: docs/DESIGN.md#D-136
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-136
  - lane:strict
semantic_text: 'An account that holds transactions but isn''t seeded started from a statement of its whole history: the core must choose transactions mode for its uploads and refuse an opening upload for it.'
logic_claims:
  - CLAIM-5698E9B01846E50B
semantic_clauses:
  - 'An account that holds transactions but isn''t seeded started from a statement of its whole history: the core must choose transactions mode for its uploads and refuse an opening upload for it.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f625a5f62e70ce72470d76753baf751adcc2627780548425e97ebcf36b1d447f
semantic_inventory:
  - claim_key: CLAIM-5698E9B01846E50B
    claim_text: 'An account that holds transactions but isn''t seeded started from a statement of its whole history: the core must choose transactions mode for its uploads and refuse an opening upload for it'
    role: normative
    status: modeled
    span:
      start: 0
      end: 189
origin:
  kind: agent
  recorded_at: '2026-11-01T07:36:02.218Z'
id: REQ-ingest-start-from-whole-history
type: req
---
An account that holds transactions but isn't seeded started from a statement of its whole history: the core must choose transactions mode for its uploads and refuse an opening upload for it.

## Context

D-136 lets an XTB account start from an export of its whole history; D-109 has the core choose opening for an account not yet seeded, which for such an account would count its holdings twice. Recorded by the XTB start implementation agent (impl/xtb-start) for D-136; default firmness, not reviewed by the author yet; firm requirements for D-136 are to come from the design session.

## Source

> An XTB account starts from an export covering its whole history, or from the open positions of a current export, which seed one opening lot per position at its open time and price, with the cash balance the owner enters once.

docs/DESIGN.md#D-136
