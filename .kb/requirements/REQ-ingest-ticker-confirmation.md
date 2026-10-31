---
title: The owner must confirm a ticker's ISIN once, proposed or typed, through an owner-only service on the core's plugin port, which stores it and books every entry waiting for it.
status: open
text_ref: docs/DESIGN.md#D-116
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-116
  - lane:strict
semantic_text: The owner must confirm a ticker's ISIN once, proposed or typed, through an owner-only service on the core's plugin port, which stores it and books every entry waiting for it.
logic_claims:
  - CLAIM-55C9BD1974A13EC9
semantic_clauses:
  - The owner must confirm a ticker's ISIN once, proposed or typed, through an owner-only service on the core's plugin port, which stores it and books every entry waiting for it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 31c5b4d2a6d63ad5b06ac67c1a52e56eab68d71a4217d599d86bf42febcb7fe5
semantic_inventory:
  - claim_key: CLAIM-55C9BD1974A13EC9
    claim_text: The owner must confirm a ticker's ISIN once, proposed or typed, through an owner-only service on the core's plugin port, which stores it and books every entry waiting for it
    role: normative
    status: modeled
    span:
      start: 0
      end: 173
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:47.122Z'
id: REQ-ingest-ticker-confirmation
type: req
---
The owner must confirm a ticker's ISIN once, proposed or typed, through an owner-only service on the core's plugin port, which stores it and books every entry waiting for it.

## Context

D-116 has the owner confirm once; D-79 and D-115 put the owner's entries on owner-only services on the plugin port. Recorded by the XTB implementation agent (impl/xtb) for the author's XTB work package; default firmness, not reviewed by the author yet.

## Source

> XTB's export names instruments by ticker and market suffix and carries no ISIN (D-35). For each new ticker the market-data connector looks up the ISIN by ticker and market and proposes it, and the owner confirms it once in the UI, or enters it when the lookup finds none. Rows of a ticker not confirmed yet wait as import problems.

docs/DESIGN.md#D-116
