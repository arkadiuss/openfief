---
title: Each OKI account's average value must be kept exact.
status: open
text_ref: design-decisions:D-144#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-144-2
  - lane:strict
semantic_text: Each OKI account's average value must be kept exact.
logic_claims:
  - CLAIM-DED050655945D68C
semantic_clauses:
  - Each OKI account's average value must be kept exact.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5ce77eb69df833eaa94195b7734fb7992bc09a0720fb8823fe3c40cbc80fc572
semantic_inventory:
  - claim_key: CLAIM-DED050655945D68C
    claim_text: Each OKI account's average value must be kept exact
    role: normative
    status: modeled
    span:
      start: 0
      end: 51
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-144#2
  recorded_at: '2026-11-01T12:26:04.796Z'
id: REQ-pl-oki-average-value-exact
type: req
---
Each OKI account's average value must be kept exact.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #40. Only the base and the tax are rounded; the average values that sum to the base are not (REQ-pl-oki-tax-base-average-value). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> OKI's tax base and tax are rounded to whole złoty, as the tax ordinance rounds bases and taxes; each account's average value is kept exact.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-144#2
