---
title: On the read API, an attention item about something tied to no account, such as an unmapped broker account, must name only its connector instance.
status: open
text_ref: design-decisions:D-165#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-165-1
  - lane:strict
semantic_text: On the read API, an attention item about something tied to no account, such as an unmapped broker account, must name only its connector instance.
logic_claims:
  - CLAIM-6CCB3AEB86C2B05C
semantic_clauses:
  - On the read API, an attention item about something tied to no account, such as an unmapped broker account, must name only its connector instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8c057edb63b35e477d0b21c2a67a4236dbbf20ed54cdab8c64d8171d58ec15ef
semantic_inventory:
  - claim_key: CLAIM-6CCB3AEB86C2B05C
    claim_text: On the read API, an attention item about something tied to no account, such as an unmapped broker account, must name only its connector instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 144
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-165#1
  recorded_at: '2026-11-02T00:55:35.398Z'
id: REQ-views-unbound-attention-names-instance-only
type: req
---
On the read API, an attention item about something tied to no account, such as an unmapped broker account, must name only its connector instance.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. Read access never includes a broker account number (D-137, D-121). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An attention item about something tied to no account, such as an unmapped broker account, names only its connector instance on the read API. Once viewers exist (D-65), only the owner sees such items.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-165#1
