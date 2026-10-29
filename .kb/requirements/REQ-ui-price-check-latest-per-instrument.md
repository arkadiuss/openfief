---
title: The UI must show the latest price check per instrument with its history.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-112-3
  - lane:strict
text_ref: design-decisions:D-112#3
origin:
  kind: import
  ref: design-decisions:D-112#3
  recorded_at: '2026-10-29T10:56:25.689Z'
semantic_text: The UI must show the latest price check per instrument with its history.
logic_claims:
  - CLAIM-32F951F536D28AA9
semantic_clauses:
  - The UI must show the latest price check per instrument with its history.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 9ef283838ef2579081e674ca6d8e4fb8a57f739d182b2bef461f59b250560692
semantic_inventory:
  - claim_key: CLAIM-32F951F536D28AA9
    claim_text: The UI must show the latest price check per instrument with its history
    role: normative
    status: modeled
    span:
      start: 0
      end: 71
id: REQ-ui-price-check-latest-per-instrument
type: req
---
The UI must show the latest price check per instrument with its history.

## Why this changed

The author's decision D-112 has the UI show the latest price check per instrument with its history, and lets the owner acknowledge a flagged check (REQ-valuation-price-check-acknowledge). REQ-ui-price-checks listed every flagged price check on its own page, so this requirement replaces it.

## Context

D-22 makes IBKR Flex prices a cross-check of the market-data prices, and D-110 and D-111 decide when a check is flagged. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> Only a connector whose manifest declares it may send cross-check prices. The owner can acknowledge a flagged check, which is recorded append-only, and the UI shows the latest check per instrument with its history.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-112#3
