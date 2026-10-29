---
title: 'What the owner enters by hand must reach the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-79-1
  - lane:strict
text_ref: design-decisions:D-79#1
origin:
  kind: import
  ref: design-decisions:D-79#1
  recorded_at: '2026-10-29T08:08:16.390Z'
semantic_text: 'What the owner enters by hand must reach the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind.'
logic_claims:
  - CLAIM-F26D89C779605379
semantic_clauses:
  - 'What the owner enters by hand must reach the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 1c4814f8c05f71d77962f2fce928b3485d9c10cb7d1a13d9b4c20d0054728459
semantic_inventory:
  - claim_key: CLAIM-F26D89C779605379
    claim_text: 'What the owner enters by hand must reach the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind'
    role: normative
    status: modeled
    span:
      start: 0
      end: 167
id: REQ-core-owner-entry-services
type: req
---
What the owner enters by hand must reach the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind.

## Context

D-79 answers the implementation's question of how the owner's entries reach the core: D-52 has the model portfolio entered in the UI and handed to the core like manual bond lots (D-49), but no decision gave that path a shape. The implementation serves ModelPortfolioService and BondLotService to view plugins only, with an owner-scope session in each call (REQ-portfolio-model-entry, REQ-ingest-bond-lot-service), and builds statement uploads the same way; D-72 adds accounts the owner creates in the UI (REQ-ingest-owner-creates-accounts). REQ-core-owner-entry-service-callers records who may call these services. The read API and script tokens stay read-only, as REQ-AUTO-69B52C44712F8813 (D-6) and REQ-AUTO-2D0C3CC0151C3C5C (D-27) already require. The author recorded D-79 and D-80 after the design session that answered the implementation's questions.

## Source

> What the owner enters by hand reaches the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind. Each is owner-only, versioned on its own, and called only by a view plugin acting for an owner session (D-69). The read API (D-6) and script tokens (D-27) stay read-only.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-79#1
