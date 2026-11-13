---
title: The proxy logs a warning at start whenever an exemption is set.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-215-2
  - lane:strict
text_ref: design-decisions:D-215#2
origin:
  kind: import
  ref: design-decisions:D-215#2
  recorded_at: '2026-11-12T23:19:17.166Z'
semantic_text: The proxy logs a warning at start whenever an exemption is set.
logic_claims:
  - CLAIM-05804CF18A9A09D5
semantic_clauses:
  - The proxy logs a warning at start whenever an exemption is set.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 0b52f2aaacd7ac2874a20c063258fdf2dfad2d0b52910b0f9d3d6fd5a98693fb
semantic_inventory:
  - claim_key: CLAIM-05804CF18A9A09D5
    claim_text: The proxy logs a warning at start whenever an exemption is set
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 62
id: REQ-plugins-egress-test-exemption-warns-at-start
type: req
---
The proxy logs a warning at start whenever an exemption is set.

## Context

Part of D-215's test-only exemption from the egress proxy's refused address ranges (REQ-plugins-egress-test-exemption-one-host): an exemption weakens D-32's rule that the proxy refuses private ranges, so the proxy says so in its log each time it starts with one, naming the exempt host. The decision states no further reason. The author recorded D-215 in DESIGN.md.

## Source

> The proxy logs a warning at start whenever an exemption is set, and the release bundle's test fails if the bundle sets one, so a shipped install never has one.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-215#2
