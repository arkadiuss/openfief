---
title: The NBP adapter must map each table A rate to an FX rate per one unit, dated on the table's effective date, parsed exactly from the JSON number's text and never through a float, and must reject a response that is not a valid table A.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: The NBP adapter must map each table A rate to an FX rate per one unit, dated on the table's effective date, parsed exactly from the JSON number's text and never through a float, and must reject a response that is not a valid table A.
logic_claims:
  - CLAIM-E02CF23AFD4E27A1
semantic_clauses:
  - The NBP adapter must map each table A rate to an FX rate per one unit, dated on the table's effective date, parsed exactly from the JSON number's text and never through a float, and must reject a response that is not a valid table A.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3b92c5d99560d9d3182b43d3065912329e96a0e262a06e0b89e3febdd9610e95
semantic_inventory:
  - claim_key: CLAIM-E02CF23AFD4E27A1
    claim_text: The NBP adapter must map each table A rate to an FX rate per one unit, dated on the table's effective date, parsed exactly from the JSON number's text and never through a float, and must reject a response that is not a valid table A
    role: normative
    status: modeled
    span:
      start: 0
      end: 232
origin:
  kind: agent
  recorded_at: '2026-10-28T05:57:19.368Z'
id: REQ-valuation-nbp-decimal-exact
type: req
---
The NBP adapter must map each table A rate to an FX rate per one unit, dated on the table's effective date, parsed exactly from the JSON number's text and never through a float, and must reject a response that is not a valid table A.

## Context

Money, prices and rates are never float64 (docs/dev/code-organization.md, AGENTS.md); a float would silently change rates such as 0.00025276. NBP's JSON shape stays in the adapter (anti-corruption), and a malformed table is an error rather than a partial or guessed result. Refines REQ-AUTO-5769411DB9663682. Modeled by Claude Code for the M0 NBP FX source task; the author has not reviewed it yet.

## Source

> In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface with NBP as the default FX plugin.

docs/DESIGN.md D-20
