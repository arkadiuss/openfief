---
title: GUS indicator 1832 must be used for CPI up to 2025 and indicator 2728 from 2026.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-104-4
  - lane:strict
text_ref: design-decisions:D-104#4
origin:
  kind: import
  ref: design-decisions:D-104#4
  recorded_at: '2026-10-29T10:03:22.680Z'
semantic_text: GUS indicator 1832 must be used for CPI up to 2025 and indicator 2728 from 2026.
logic_claims:
  - CLAIM-6F04299D0E6E7F01
semantic_clauses:
  - GUS indicator 1832 must be used for CPI up to 2025 and indicator 2728 from 2026.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 095479d9bc3115568c6b33e0422fd73355c66a5772735a17bed41ce0e6ba379f
semantic_inventory:
  - claim_key: CLAIM-6F04299D0E6E7F01
    claim_text: GUS indicator 1832 must be used for CPI up to 2025 and indicator 2728 from 2026
    role: normative
    status: modeled
    span:
      start: 0
      end: 79
id: REQ-valuation-gus-cpi-indicators
type: req
---
GUS indicator 1832 must be used for CPI up to 2025 and indicator 2728 from 2026.

## Context

D-104 names the GUS indicator that CPI readings come from for each span of years: 1832 up to 2025 and 2728 from 2026. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> CPI readings use only GUS's final monthly values; preliminary, estimated and flash values are skipped. The first sync starts in January thirteen years back and fetches at most four years per sync, within GUS's anonymous limits. GUS indicator 1832 is used up to 2025 and 2728 from 2026.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-104#4
