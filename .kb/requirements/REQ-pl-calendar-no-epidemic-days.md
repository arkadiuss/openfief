---
title: The Polish business-day calendar must not model epidemic days off.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-103-2
  - lane:strict
text_ref: design-decisions:D-103#2
origin:
  kind: import
  ref: design-decisions:D-103#2
  recorded_at: '2026-10-29T10:00:11.427Z'
semantic_text: The Polish business-day calendar must not model epidemic days off.
logic_claims:
  - CLAIM-17A1F4C94C81B2BE
semantic_clauses:
  - The Polish business-day calendar must not model epidemic days off.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 4190bd5b2286e55ffed0b65056e8e91ce6aa095cfbdefc56f4685929acb21f4b
semantic_inventory:
  - claim_key: CLAIM-17A1F4C94C81B2BE
    claim_text: The Polish business-day calendar must not model epidemic days off
    role: normative
    status: modeled
    span:
      start: 0
      end: 65
id: REQ-pl-calendar-no-epidemic-days
type: req
---
The Polish business-day calendar must not model epidemic days off.

## Context

D-103 defines the Polish business-day calendar (REQ-pl-business-day-calendar) and leaves epidemic days off out of it. Reason not stated. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> The Polish business-day calendar counts Saturdays, Sundays and the statutory public holidays year by year, including 6 January from 2011 and 24 December from 2025, with the Easter-based days computed. Epidemic days off are not modelled, and dates before 1991 are refused.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-103#2
