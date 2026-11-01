---
title: An XTB holding assertion for a ticker not confirmed yet must be made once the owner confirms the ticker.
status: open
text_ref: design-decisions:D-155#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-155-2
  - lane:strict
semantic_text: An XTB holding assertion for a ticker not confirmed yet must be made once the owner confirms the ticker.
logic_claims:
  - CLAIM-AC84462EB4BFEDBC
semantic_clauses:
  - An XTB holding assertion for a ticker not confirmed yet must be made once the owner confirms the ticker.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3145c64ca742f77abe4d2eccc6b53cc329c6dc0f890b6338a73be824603037be
semantic_inventory:
  - claim_key: CLAIM-AC84462EB4BFEDBC
    claim_text: An XTB holding assertion for a ticker not confirmed yet must be made once the owner confirms the ticker
    role: normative
    status: modeled
    span:
      start: 0
      end: 103
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-155#2
  recorded_at: '2026-11-01T12:30:14.800Z'
id: REQ-ingest-xtb-unconfirmed-ticker-asserted-on-confirm
type: req
---
An XTB holding assertion for a ticker not confirmed yet must be made once the owner confirms the ticker.

## Context

The author recorded it on 2026-11-01, settling the last implementation questions (#33, #38, #39). D-116 has rows of an unconfirmed ticker wait as import problems (REQ-ingest-xtb-unconfirmed-ticker-waits) until the owner confirms its ISIN. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An XTB export's `Open Positions` sheet also asserts the account's holdings (D-18): for each ticker with a confirmed ISIN (D-116), the summed volume of its open positions on the report's date. A ticker not confirmed yet is asserted once the owner confirms it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-155#2
