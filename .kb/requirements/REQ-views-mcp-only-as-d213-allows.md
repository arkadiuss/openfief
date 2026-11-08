---
title: MCP must be served only as D-213 allows.
status: open
text_ref: design-decisions:D-6#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-6-4
  - lane:strict
semantic_text: MCP must be served only as D-213 allows.
logic_claims:
  - CLAIM-212B79509E0DA1FA
semantic_clauses:
  - MCP must be served only as D-213 allows.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 457ce5ae3306a2b32f85e9c891bbe2ebe815b79374330a5ba3666d1d39946535
semantic_inventory:
  - claim_key: CLAIM-212B79509E0DA1FA
    claim_text: MCP must be served only as D-213 allows
    role: normative
    status: modeled
    span:
      start: 0
      end: 39
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-6#4
  recorded_at: '2026-11-08T06:51:00.766Z'
id: REQ-views-mcp-only-as-d213-allows
type: req
---
MCP must be served only as D-213 allows.

## Context

Author decision D-6, the read-only public API. Its last sentence used to keep MCP out until M2, so that a model reached holdings only through the model router. D-213 has since decided an MCP server for external agents: the core can serve MCP at /mcp on its browser port with the read API's read-only answerers as tools, for a script token the owner allowed to use MCP, limited by its grants (D-209), never changing anything and audited as usage (D-211). Nothing serves MCP yet, so no code implements this requirement.

## Source

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it. MCP is served only as D-213 allows.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-6#4

## Why this changed

REQ-AUTO-B8D1BCB44E147E0C was imported from D-6's last sentence, "MCP access waits for M2, so that model access to holdings goes through the model router". D-213 overrode "MCP access waits for M2" by deciding when and how the core serves MCP, so the author reworded D-6 to defer to it, and the requirement is imported again from the new sentence.
