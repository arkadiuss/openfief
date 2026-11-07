---
title: The owner must be able to override any series' terms.
status: open
text_ref: design-decisions:D-97#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-97-3
  - lane:strict
semantic_text: The owner must be able to override any series' terms.
logic_claims:
  - CLAIM-B2B96551D715DD68
semantic_clauses:
  - The owner must be able to override any series' terms.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a92f1b5c3a86f65f8a6b0a08d5d73d387e2e86e23d0ba01286648612d4479675
semantic_inventory:
  - claim_key: CLAIM-B2B96551D715DD68
    claim_text: The owner must be able to override any series' terms
    role: normative
    status: modeled
    span:
      start: 0
      end: 52
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-97#3
  recorded_at: '2026-11-07T12:37:43.959Z'
id: REQ-ingest-bond-terms-owner-overrides
type: req
---
The owner must be able to override any series' terms.

## Context

D-9 has the owner enter treasury bond lots by hand, and the bond rules need each series' first-period rate, margin and early-redemption fee to value them. Author decision D-97, as reworded when the design session answered OQ-30, says where those terms come from: the rates from the Ministry of Finance's published retail bond data, whose licence (CC BY 4.0) lets anyone re-use it when the ministry is named as the source (R-27), the fees from a reviewed table of the issue letters (R-28), and the owner's own entry over both, the latest entry winning (REQ-ingest-bond-terms-latest-entry-wins).

## Source

> The owner can override any series' terms, and the latest entry wins.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-97#3

## Why this changed

REQ-ingest-bond-terms-entered-with-lot was imported from D-97's first wording, under which the owner had to enter a series' terms with a lot until the connector could read them from obligacjeskarbowe.pl. The reworded D-97 has the connector fill in the terms from the ministry's data and the reviewed fee table, so entering them is no longer required; what stays is that the owner can override any series' terms, which this requirement states. That the latest entry wins stays as the firm REQ-ingest-bond-terms-latest-entry-wins.
