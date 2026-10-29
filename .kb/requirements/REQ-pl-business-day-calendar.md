---
title: The Polish business-day calendar must count as days off Saturdays, Sundays and the statutory public holidays year by year, including 6 January from 2011 and 24 December from 2025, with the Easter-based days computed.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-103-1
  - lane:strict
text_ref: design-decisions:D-103#1
origin:
  kind: import
  ref: design-decisions:D-103#1
  recorded_at: '2026-10-29T09:59:35.050Z'
semantic_text: The Polish business-day calendar must count as days off Saturdays, Sundays and the statutory public holidays year by year, including 6 January from 2011 and 24 December from 2025, with the Easter-based days computed.
logic_claims:
  - CLAIM-0B4F89FD17BB4C93
semantic_clauses:
  - The Polish business-day calendar must count as days off Saturdays, Sundays and the statutory public holidays year by year, including 6 January from 2011 and 24 December from 2025, with the Easter-based days computed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 23571464dc40afee6db857bec92d3b5612a0c2c1684b9d039e28ad11b5f004b5
semantic_inventory:
  - claim_key: CLAIM-0B4F89FD17BB4C93
    claim_text: The Polish business-day calendar must count as days off Saturdays, Sundays and the statutory public holidays year by year, including 6 January from 2011 and 24 December from 2025, with the Easter-based days computed
    role: normative
    status: modeled
    span:
      start: 0
      end: 215
id: REQ-pl-business-day-calendar
type: req
---
The Polish business-day calendar must count as days off Saturdays, Sundays and the statutory public holidays year by year, including 6 January from 2011 and 24 December from 2025, with the Easter-based days computed.

## Context

D-103 defines the Polish business-day calendar: which days are off in each year, with the holidays that changed over time (6 January from 2011, 24 December from 2025) and the Easter-based days computed rather than listed. REQ-pl-calendar-no-epidemic-days and REQ-pl-calendar-from-1991 hold the rest of D-103. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> The Polish business-day calendar counts Saturdays, Sundays and the statutory public holidays year by year, including 6 January from 2011 and 24 December from 2025, with the Easter-based days computed. Epidemic days off are not modelled, and dates before 1991 are refused.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-103#1
