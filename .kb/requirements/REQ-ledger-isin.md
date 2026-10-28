---
title: An instrument must be identified by an ISIN with a correct check digit plus a venue symbol, and any other identifier must be rejected.
status: open
text_ref: docs/glossary.md#m0
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-glossary-md-m0
  - lane:strict
semantic_text: An instrument must be identified by an ISIN with a correct check digit plus a venue symbol, and any other identifier must be rejected.
logic_claims:
  - CLAIM-74BCA39AA34B932F
semantic_clauses:
  - An instrument must be identified by an ISIN with a correct check digit plus a venue symbol, and any other identifier must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 18f1644e79d2de8cfbfa52d9caed88ff306eb68fe1453ef3a427fba8fa77ec01
semantic_inventory:
  - claim_key: CLAIM-74BCA39AA34B932F
    claim_text: An instrument must be identified by an ISIN with a correct check digit plus a venue symbol, and any other identifier must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 133
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:18.615Z'
id: REQ-ledger-isin
type: req
---
An instrument must be identified by an ISIN with a correct check digit plus a venue symbol, and any other identifier must be rejected.

## Context

The glossary defines an instrument as something that can be held, identified by ISIN plus a venue symbol. Holdings are keyed by instrument, so a mistyped or corrupted ISIN would silently split one holding into two; checking the ISO 6166 check digit catches that at the boundary. The orchestrating agent's task asked for the check digit to be validated. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Something that can be held, identified by ISIN plus a venue symbol. A bond series is an instrument.

docs/glossary.md (Instrument)
