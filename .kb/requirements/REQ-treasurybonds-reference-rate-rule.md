---
title: 'The treasury bond connector must apply the reference-rate rule itself: the NBP reference rate in force on the tenth business day before a month is that month''s reading, dated on the month''s first day.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-102-1
  - lane:strict
text_ref: design-decisions:D-102#1
origin:
  kind: import
  ref: design-decisions:D-102#1
  recorded_at: '2026-10-29T09:58:42.898Z'
semantic_text: 'The treasury bond connector must apply the reference-rate rule itself: the NBP reference rate in force on the tenth business day before a month is that month''s reading, dated on the month''s first day.'
logic_claims:
  - CLAIM-10A8EA9AF652DFB6
semantic_clauses:
  - 'The treasury bond connector must apply the reference-rate rule itself: the NBP reference rate in force on the tenth business day before a month is that month''s reading, dated on the month''s first day.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 076d90c45f26a4786ea9a2347144f6bf59a273aa84fc1790fbc81764adad73c5
semantic_inventory:
  - claim_key: CLAIM-10A8EA9AF652DFB6
    claim_text: 'The treasury bond connector must apply the reference-rate rule itself: the NBP reference rate in force on the tenth business day before a month is that month''s reading, dated on the month''s first day'
    role: normative
    status: modeled
    span:
      start: 0
      end: 199
id: REQ-treasurybonds-reference-rate-rule
type: req
---
The treasury bond connector must apply the reference-rate rule itself: the NBP reference rate in force on the tenth business day before a month is that month's reading, dated on the month's first day.

## Context

D-94 ranks the treasury bond connector first for the NBP reference rate. D-102 has the connector apply the reference-rate rule itself, so each month gets one reading dated on its first day. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> The treasury bond connector applies the reference-rate rule itself: the NBP reference rate in force on the tenth business day before a month is that month's reading, dated on the month's first day.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-102#1
