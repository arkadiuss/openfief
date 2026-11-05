---
title: The core must run one connector sync worker unless the owner raises the worker count together with the core's memory limit.
status: open
text_ref: design-decisions:D-189#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-189-1
  - lane:strict
semantic_text: The core must run one connector sync worker unless the owner raises the worker count together with the core's memory limit.
logic_claims:
  - CLAIM-0DBF91F0ABCFA082
semantic_clauses:
  - The core must run one connector sync worker unless the owner raises the worker count together with the core's memory limit.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4c14f04ce5fef206bea322df59debe073b88631a40f2f6c4e6fee2bc27b92664
semantic_inventory:
  - claim_key: CLAIM-0DBF91F0ABCFA082
    claim_text: The core must run one connector sync worker unless the owner raises the worker count together with the core's memory limit
    role: exception
    status: modeled
    span:
      start: 0
      end: 122
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-189#1
  recorded_at: '2026-11-03T18:24:13.355Z'
id: REQ-core-one-sync-worker-by-default
type: req
---
The core must run one connector sync worker unless the owner raises the worker count together with the core's memory limit.

## Context

The author decided it on 2026-11-03, before tagging alpha.1, after a review of the core's and the plugins' memory: each sync holds its delivery in memory while it imports it, and a Go plugin that never sees its container's limit grows its heap until the kernel kills it. D-49 has the core call connectors on a schedule; the core's --sync-workers flag sets the count, 1 by default, and the core's Deployment sets its memory limit and GOMEMLIMIT. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> [D-189] The core runs one sync worker unless the owner raises it together with the core's memory limit, and the operator sets each plugin container's `GOMEMLIMIT` from its memory limit, so a Go plugin collects garbage before the kernel kills it.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-189#1
