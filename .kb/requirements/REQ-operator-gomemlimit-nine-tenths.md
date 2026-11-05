---
title: A plugin container's GOMEMLIMIT must be nine tenths of its memory limit in bytes, rounded down, and unset when the container has no memory limit.
status: open
text_ref: docs/DESIGN.md#D-189
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-189
  - lane:strict
semantic_text: A plugin container's GOMEMLIMIT must be nine tenths of its memory limit in bytes, rounded down, and unset when the container has no memory limit.
logic_claims:
  - CLAIM-2B47267B9717CAE6
semantic_clauses:
  - A plugin container's GOMEMLIMIT must be nine tenths of its memory limit in bytes, rounded down, and unset when the container has no memory limit.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d17da1709bb84721174dca37e0d522140f56c58b8070280fa693979b3a189aed
semantic_inventory:
  - claim_key: CLAIM-2B47267B9717CAE6
    claim_text: A plugin container's GOMEMLIMIT must be nine tenths of its memory limit in bytes, rounded down, and unset when the container has no memory limit
    role: normative
    status: modeled
    span:
      start: 0
      end: 144
origin:
  kind: agent
  recorded_at: '2026-11-03T18:24:36.104Z'
id: REQ-operator-gomemlimit-nine-tenths
type: req
---
A plugin container's GOMEMLIMIT must be nine tenths of its memory limit in bytes, rounded down, and unset when the container has no memory limit.

## Context

D-189 doesn't say how much below the limit GOMEMLIMIT sits; the implementation chose nine tenths, a common headroom, leaving a tenth for what the kernel counts against the container but the Go runtime doesn't, such as the binary's mapped pages, and for garbage not yet collected. Refines REQ-operator-plugin-gomemlimit. The operator always sets it: a manifest can't set environment variables.

## Source

> [D-189] The core runs one sync worker unless the owner raises it together with the core's memory limit, and the operator sets each plugin container's `GOMEMLIMIT` from its memory limit, so a Go plugin collects garbage before the kernel kills it.

docs/DESIGN.md#D-189
