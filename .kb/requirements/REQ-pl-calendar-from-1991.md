---
title: The Polish business-day calendar must refuse dates before 1991.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-103-3
  - lane:strict
text_ref: design-decisions:D-103#3
origin:
  kind: import
  ref: design-decisions:D-103#3
  recorded_at: '2026-10-29T10:00:48.432Z'
semantic_text: The Polish business-day calendar must refuse dates before 1991.
logic_claims:
  - CLAIM-476B133AFD59FCEE
semantic_clauses:
  - The Polish business-day calendar must refuse dates before 1991.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7a53c02537fe8cf40a32eee2f9feeafe46d76e3fa37baf844b9a576284de08fe
semantic_inventory:
  - claim_key: CLAIM-476B133AFD59FCEE
    claim_text: The Polish business-day calendar must refuse dates before 1991
    role: normative
    status: modeled
    span:
      start: 0
      end: 62
id: REQ-pl-calendar-from-1991
type: req
---
The Polish business-day calendar must refuse dates before 1991.

## Context

D-103 defines the Polish business-day calendar (REQ-pl-business-day-calendar) and limits it to dates from 1991 on. Reason not stated. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> The Polish business-day calendar counts Saturdays, Sundays and the statutory public holidays year by year, including 6 January from 2011 and 24 December from 2025, with the Easter-based days computed. Epidemic days off are not modelled, and dates before 1991 are refused.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-103#3
