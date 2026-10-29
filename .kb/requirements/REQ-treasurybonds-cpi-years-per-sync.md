---
title: Each CPI sync must fetch at most four years, within GUS's anonymous limits.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-104-3
  - lane:strict
text_ref: design-decisions:D-104#3
origin:
  kind: import
  ref: design-decisions:D-104#3
  recorded_at: '2026-10-29T10:02:43.651Z'
semantic_text: Each CPI sync must fetch at most four years, within GUS's anonymous limits.
logic_claims:
  - CLAIM-6299769092EC658D
semantic_clauses:
  - Each CPI sync must fetch at most four years, within GUS's anonymous limits.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 72011fb9039a88c029ce050baf1ca3ac11afecdddbfc0eb3873046ed30a5f3e5
semantic_inventory:
  - claim_key: CLAIM-6299769092EC658D
    claim_text: Each CPI sync must fetch at most four years, within GUS's anonymous limits
    role: normative
    status: modeled
    span:
      start: 0
      end: 74
id: REQ-treasurybonds-cpi-years-per-sync
type: req
---
Each CPI sync must fetch at most four years, within GUS's anonymous limits.

## Context

D-104 caps how much CPI history one sync fetches from GUS at four years, so the treasury bond connector stays within the limits GUS sets for anonymous callers; the first sync reaches thirteen years back (REQ-treasurybonds-cpi-first-sync-start). The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> CPI readings use only GUS's final monthly values; preliminary, estimated and flash values are skipped. The first sync starts in January thirteen years back and fetches at most four years per sync, within GUS's anonymous limits. GUS indicator 1832 is used up to 2025 and 2728 from 2026.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-104#3
