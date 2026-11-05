---
title: The operator must set each plugin container's GOMEMLIMIT from the container's memory limit.
status: open
text_ref: design-decisions:D-189#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-189-2
  - lane:strict
semantic_text: The operator must set each plugin container's GOMEMLIMIT from the container's memory limit.
logic_claims:
  - CLAIM-77E4C9380217DE02
semantic_clauses:
  - The operator must set each plugin container's GOMEMLIMIT from the container's memory limit.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e4869d30102bd6f2c86b7b73a2150ab0a3df70be2eb0258fe541f3d6e47d474e
semantic_inventory:
  - claim_key: CLAIM-77E4C9380217DE02
    claim_text: The operator must set each plugin container's GOMEMLIMIT from the container's memory limit
    role: normative
    status: modeled
    span:
      start: 0
      end: 90
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-189#2
  recorded_at: '2026-11-03T18:24:20.413Z'
id: REQ-operator-plugin-gomemlimit
type: req
---
The operator must set each plugin container's GOMEMLIMIT from the container's memory limit.

## Context

The author decided it on 2026-11-03, before tagging alpha.1, after a review of the core's and the plugins' memory: each sync holds its delivery in memory while it imports it, and a Go plugin that never sees its container's limit grows its heap until the kernel kills it. D-182 gives every plugin container a memory limit; GOMEMLIMIT is Go's soft memory limit, which makes the garbage collector work harder as the heap nears it. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> [D-189] The core runs one sync worker unless the owner raises it together with the core's memory limit, and the operator sets each plugin container's `GOMEMLIMIT` from its memory limit, so a Go plugin collects garbage before the kernel kills it.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-189#2
