---
title: CPI readings must use only GUS's final monthly values, skipping preliminary, estimated and flash values.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-104-1
  - lane:strict
text_ref: design-decisions:D-104#1
origin:
  kind: import
  ref: design-decisions:D-104#1
  recorded_at: '2026-10-29T10:01:26.886Z'
semantic_text: CPI readings must use only GUS's final monthly values, skipping preliminary, estimated and flash values.
logic_claims:
  - CLAIM-4A3EF58FFCDCF83D
semantic_clauses:
  - CPI readings must use only GUS's final monthly values, skipping preliminary, estimated and flash values.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: aab406be6ec3d8b8dadb22b229355b0cb91bd54d9197ed0ca06e54e5b89af4ba
semantic_inventory:
  - claim_key: CLAIM-4A3EF58FFCDCF83D
    claim_text: CPI readings must use only GUS's final monthly values, skipping preliminary, estimated and flash values
    role: normative
    status: modeled
    span:
      start: 0
      end: 103
id: REQ-valuation-cpi-final-only
type: req
---
CPI readings must use only GUS's final monthly values, skipping preliminary, estimated and flash values.

## Context

D-94 ranks the treasury bond connector first for CPI readings. D-104 decides which of GUS's published values count as a reading: only the final monthly ones. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> CPI readings use only GUS's final monthly values; preliminary, estimated and flash values are skipped. The first sync starts in January thirteen years back and fetches at most four years per sync, within GUS's anonymous limits. GUS indicator 1832 is used up to 2025 and 2728 from 2026.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-104#1
