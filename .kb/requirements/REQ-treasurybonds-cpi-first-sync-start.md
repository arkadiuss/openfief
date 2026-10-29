---
title: The first CPI sync must start in January thirteen years back.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-104-2
  - lane:strict
text_ref: design-decisions:D-104#2
origin:
  kind: import
  ref: design-decisions:D-104#2
  recorded_at: '2026-10-29T10:02:05.569Z'
semantic_text: The first CPI sync must start in January thirteen years back.
logic_claims:
  - CLAIM-20340300486E4477
semantic_clauses:
  - The first CPI sync must start in January thirteen years back.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 5d04185afdf992c0afc5edbe5d73f2ba0a57730388353849e3db2eaff12faf44
semantic_inventory:
  - claim_key: CLAIM-20340300486E4477
    claim_text: The first CPI sync must start in January thirteen years back
    role: normative
    status: modeled
    span:
      start: 0
      end: 60
id: REQ-treasurybonds-cpi-first-sync-start
type: req
---
The first CPI sync must start in January thirteen years back.

## Context

D-94 ranks the treasury bond connector first for CPI readings. D-104 sets how far back its first CPI sync reaches: January, thirteen years back. Reason not stated. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> CPI readings use only GUS's final monthly values; preliminary, estimated and flash values are skipped. The first sync starts in January thirteen years back and fetches at most four years per sync, within GUS's anonymous limits. GUS indicator 1832 is used up to 2025 and 2728 from 2026.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-104#2
