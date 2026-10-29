---
title: In M0 a jurisdiction pack must be a set of pure functions inside the core, behind one interface, so that another jurisdiction is another implementation of it, and must not be a plugin container or a WebAssembly (WASM) module.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-70-1
  - lane:strict
text_ref: design-decisions:D-70#1
origin:
  kind: import
  ref: design-decisions:D-70#1
  recorded_at: '2026-10-29T05:33:06.644Z'
semantic_text: In M0 a jurisdiction pack must be a set of pure functions inside the core, behind one interface, so that another jurisdiction is another implementation of it, and must not be a plugin container or a WebAssembly (WASM) module.
logic_claims:
  - CLAIM-56C75372137FA3C3
semantic_clauses:
  - In M0 a jurisdiction pack must be a set of pure functions inside the core, behind one interface, so that another jurisdiction is another implementation of it, and must not be a plugin container or a WebAssembly (WASM) module.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 407b5d51927c0a130310c3429df2b8ac85c6f89a3b0654d8b7250c249d3abfb6
semantic_inventory:
  - claim_key: CLAIM-56C75372137FA3C3
    claim_text: In M0 a jurisdiction pack must be a set of pure functions inside the core, behind one interface, so that another jurisdiction is another implementation of it, and must not be a plugin container or a WebAssembly (WASM) module
    role: normative
    status: modeled
    span:
      start: 0
      end: 224
id: REQ-jurisdiction-in-process
type: req
---
In M0 a jurisdiction pack must be a set of pure functions inside the core, behind one interface, so that another jurisdiction is another implementation of it, and must not be a plugin container or a WebAssembly (WASM) module.

## Context

Double-entry keeps cash, withholding tax and currency conversion consistent by construction, and it gives a later lease a natural home as a sub-account. D-70 is the author's reading of GR-13 (regulatory logic is pluggable too: jurisdiction plugins, Poland first) for M0, so the jurisdiction plugins of GR-13 are, in M0, implementations of one interface inside the core rather than plugin containers. The author recorded it with D-65 to D-69, D-71 and D-72 after the design session that answered the implementation's questions.

## Source

> Reading of GR-13 for M0: a jurisdiction pack is a set of pure functions inside the core, behind one interface, so another jurisdiction is another implementation of it. It is not a plugin container and not a WebAssembly (WASM) module.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-70#1
