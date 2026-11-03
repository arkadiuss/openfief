---
title: Only a connector whose signed manifest declares index readings, such as CPI and the NBP reference rate, and whose grant approves that may deliver them; any other connector's index readings must be dropped with an attention item, as D-179 drops bond terms.
status: open
text_ref: design-decisions:D-186#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-186-1
  - lane:strict
semantic_text: Only a connector whose signed manifest declares index readings, such as CPI and the NBP reference rate, and whose grant approves that may deliver them; any other connector's index readings must be dropped with an attention item, as D-179 drops bond terms.
logic_claims:
  - CLAIM-0A97212A82AF36BD
semantic_clauses:
  - Only a connector whose signed manifest declares index readings, such as CPI and the NBP reference rate, and whose grant approves that may deliver them; any other connector's index readings must be dropped with an attention item, as D-179 drops bond terms.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ff4a9bfe51a130930bc94179e2db9888dcbefe92811d9319045c91285ec18904
semantic_inventory:
  - claim_key: CLAIM-0A97212A82AF36BD
    claim_text: Only a connector whose signed manifest declares index readings, such as CPI and the NBP reference rate, and whose grant approves that may deliver them; any other connector's index readings must be dropped with an attention item, as D-179 drops bond terms
    role: normative
    status: modeled
    span:
      start: 0
      end: 254
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-186#1
  recorded_at: '2026-11-03T10:07:43.584Z'
id: REQ-ingest-index-readings-declared-approved
type: req
---
Only a connector whose signed manifest declares index readings, such as CPI and the NBP reference rate, and whose grant approves that may deliver them; any other connector's index readings must be dropped with an attention item, as D-179 drops bond terms.

## Context

The author decided it on 2026-11-03, in the review's second pass before the alpha tag, after finding that any connector's index readings were stored and ranked as a source. D-179 made bond series terms and market index levels permissions a connector's signed manifest declares and its grant approves, dropping undeclared ones with an attention item; bond terms value the owner's bonds from these readings, so D-186 extends the same rule to them. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> D-179 also covers the index readings bond terms depend on, such as CPI and the NBP reference rate: only a connector whose manifest declares them and whose grant approves that may deliver them, and each stored reading records the instance that delivered it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-186#1
