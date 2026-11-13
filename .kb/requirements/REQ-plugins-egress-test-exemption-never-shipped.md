---
title: The release bundle's test fails if the bundle sets one, so a shipped install never has one.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-215-3
  - lane:strict
text_ref: design-decisions:D-215#3
origin:
  kind: import
  ref: design-decisions:D-215#3
  recorded_at: '2026-11-12T23:19:36.473Z'
semantic_text: The release bundle's test fails if the bundle sets one, so a shipped install never has one.
logic_claims:
  - CLAIM-32E74271AE366285
semantic_clauses:
  - The release bundle's test fails if the bundle sets one, so a shipped install never has one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7059c76decf7f3f4967fd5709463efad3181985e34446a96eecc5b56648510d4
semantic_inventory:
  - claim_key: CLAIM-32E74271AE366285
    claim_text: The release bundle's test fails if the bundle sets one, so a shipped install never has one
    role: normative
    status: modeled
    span:
      start: 0
      end: 90
id: REQ-plugins-egress-test-exemption-never-shipped
type: req
---
The release bundle's test fails if the bundle sets one, so a shipped install never has one.

## Context

Part of D-215's test-only exemption from the egress proxy's refused address ranges (REQ-plugins-egress-test-exemption-one-host): "one" is an exemption. The release bundle's script test (scripts/tests/release-bundle.test.sh) and the script itself refuse a bundle whose manifests set the exemption flag or environment variable. The decision states no further reason. The author recorded D-215 in DESIGN.md.

## Source

> The proxy logs a warning at start whenever an exemption is set, and the release bundle's test fails if the bundle sets one, so a shipped install never has one.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-215#3
