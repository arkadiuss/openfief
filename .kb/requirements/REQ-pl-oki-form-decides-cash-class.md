---
title: An OKI account's form must decide whether its PLN cash is a savings or an investment asset.
status: open
text_ref: design-decisions:D-147#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-147-2
  - lane:strict
semantic_text: An OKI account's form must decide whether its PLN cash is a savings or an investment asset.
logic_claims:
  - CLAIM-C174B5BAD1B00746
semantic_clauses:
  - An OKI account's form must decide whether its PLN cash is a savings or an investment asset.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 856956bfd56c1e8dc6066759637fcc55c32cadb7815b0bed4bf97348e793f224
semantic_inventory:
  - claim_key: CLAIM-C174B5BAD1B00746
    claim_text: An OKI account's form must decide whether its PLN cash is a savings or an investment asset
    role: normative
    status: modeled
    span:
      start: 0
      end: 90
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-147#2
  recorded_at: '2026-11-01T12:27:19.444Z'
id: REQ-pl-oki-form-decides-cash-class
type: req
---
An OKI account's form must decide whether its PLN cash is a savings or an investment asset.

## Context

The author recorded it on 2026-11-01, answering the implementation's OKI questions (#40 to #43, #46). D-127 counts PLN cash as a savings asset, and PLN cash on a brokerage account as an investment asset, for its exemptions (REQ-pl-oki-tax-exemptions). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner sets an OKI account's form when creating it: a bank account, a treasury bond account, a brokerage account or a fund register. The form decides whether its PLN cash is a savings or an investment asset.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-147#2
