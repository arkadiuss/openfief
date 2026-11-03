---
title: Every stored index reading must record the connector instance that delivered it.
status: open
text_ref: design-decisions:D-186#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-186-2
  - lane:strict
semantic_text: Every stored index reading must record the connector instance that delivered it.
logic_claims:
  - CLAIM-2F972296FE3947F1
semantic_clauses:
  - Every stored index reading must record the connector instance that delivered it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 550178b1a8dc8e14b7f469e46591127ca21e1c214f282297241a70380cf47f74
semantic_inventory:
  - claim_key: CLAIM-2F972296FE3947F1
    claim_text: Every stored index reading must record the connector instance that delivered it
    role: normative
    status: modeled
    span:
      start: 0
      end: 79
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-186#2
  recorded_at: '2026-11-03T10:07:56.232Z'
id: REQ-valuation-index-readings-record-instance
type: req
---
Every stored index reading must record the connector instance that delivered it.

## Context

The author decided it on 2026-11-03, in the review's second pass before the alpha tag, after finding that any connector's index readings were stored and ranked as a source. D-124 makes a source one connector instance, and D-179 has every stored bond term and index level record its instance; D-186 says the same of index readings. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> D-179 also covers the index readings bond terms depend on, such as CPI and the NBP reference rate: only a connector whose manifest declares them and whose grant approves that may deliver them, and each stored reading records the instance that delivered it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-186#2
