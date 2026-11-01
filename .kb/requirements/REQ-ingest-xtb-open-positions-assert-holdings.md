---
title: 'An XTB export''s Open Positions sheet must also assert the account''s holdings: for each ticker with a confirmed ISIN, the summed volume of its open positions on the report''s date.'
status: open
text_ref: design-decisions:D-155#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-155-1
  - lane:strict
semantic_text: 'An XTB export''s Open Positions sheet must also assert the account''s holdings: for each ticker with a confirmed ISIN, the summed volume of its open positions on the report''s date.'
logic_claims:
  - CLAIM-E87C693F776FB8EA
semantic_clauses:
  - 'An XTB export''s Open Positions sheet must also assert the account''s holdings: for each ticker with a confirmed ISIN, the summed volume of its open positions on the report''s date.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 782ce7e01dd8a5bf906bbf359eba9daeb6dd6f38c6b433e8310001a2c80d45aa
semantic_inventory:
  - claim_key: CLAIM-E87C693F776FB8EA
    claim_text: 'An XTB export''s Open Positions sheet must also assert the account''s holdings: for each ticker with a confirmed ISIN, the summed volume of its open positions on the report''s date'
    role: normative
    status: modeled
    span:
      start: 0
      end: 177
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-155#1
  recorded_at: '2026-11-01T12:30:07.351Z'
id: REQ-ingest-xtb-open-positions-assert-holdings
type: req
---
An XTB export's Open Positions sheet must also assert the account's holdings: for each ticker with a confirmed ISIN, the summed volume of its open positions on the report's date.

## Context

The author recorded it on 2026-11-01, settling the last implementation questions (#33, #38, #39). D-18 stores broker position snapshots as balance assertions and flags a mismatch (REQ-AUTO-D9FBE33256704D04); D-136 also seeds opening lots from the same sheet. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An XTB export's `Open Positions` sheet also asserts the account's holdings (D-18): for each ticker with a confirmed ISIN (D-116), the summed volume of its open positions on the report's date. A ticker not confirmed yet is asserted once the owner confirms it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-155#1
