---
title: OKI's tax base and tax must be rounded to whole złoty.
status: open
text_ref: design-decisions:D-144#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-144-1
  - lane:strict
semantic_text: OKI's tax base and tax must be rounded to whole złoty.
logic_claims:
  - CLAIM-DDEB9430F354EC6C
semantic_clauses:
  - OKI's tax base and tax must be rounded to whole złoty.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bdcda9e7e38e52847d6f40cb15818369f85ce82cdb7050f2af810533322ae847
semantic_inventory:
  - claim_key: CLAIM-DDEB9430F354EC6C
    claim_text: OKI's tax base and tax must be rounded to whole złoty
    role: normative
    status: modeled
    span:
      start: 0
      end: 54
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-144#1
  recorded_at: '2026-11-01T12:25:57.150Z'
id: REQ-pl-oki-base-and-tax-whole-zloty
type: req
---
OKI's tax base and tax must be rounded to whole złoty.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #40. The decision gives the reason: the tax ordinance rounds bases and taxes. D-127 defines OKI's yearly tax base (REQ-pl-oki-tax-base-average-value). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> OKI's tax base and tax are rounded to whole złoty, as the tax ordinance rounds bases and taxes; each account's average value is kept exact.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-144#1
