---
title: The demo loader must record NBP table A rates for the demo's currencies, fetched through the NBP adapter, or offline or when NBP fails, from a committed copy of NBP's published rates.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: The demo loader must record NBP table A rates for the demo's currencies, fetched through the NBP adapter, or offline or when NBP fails, from a committed copy of NBP's published rates.
logic_claims:
  - CLAIM-0E66ED836A08E649
semantic_clauses:
  - The demo loader must record NBP table A rates for the demo's currencies, fetched through the NBP adapter, or offline or when NBP fails, from a committed copy of NBP's published rates.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4b13a3ca170b0e6809ebb8a3f8659092269a98565a657e231a7ce88dc1ac5a00
semantic_inventory:
  - claim_key: CLAIM-0E66ED836A08E649
    claim_text: The demo loader must record NBP table A rates for the demo's currencies, fetched through the NBP adapter, or offline or when NBP fails, from a committed copy of NBP's published rates
    role: normative
    status: modeled
    span:
      start: 0
      end: 182
origin:
  kind: agent
  recorded_at: '2026-10-28T09:31:56.342Z'
id: REQ-demo-fx-rates
type: req
---
The demo loader must record NBP table A rates for the demo's currencies, fetched through the NBP adapter, or offline or when NBP fails, from a committed copy of NBP's published rates.

## Context

D-20 makes NBP table A the FX source; the demo cluster may have no egress, so a committed copy of the public rates keeps the load working, and NBP's rates are public data. Modeled by Claude Code for the demo-data task: the author asked to see OpenFief on the Hetzner demo with real pipelines on fake data; the author has not reviewed this requirement yet.

## Source

> In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface with NBP as the default FX plugin.

docs/DESIGN.md D-20
