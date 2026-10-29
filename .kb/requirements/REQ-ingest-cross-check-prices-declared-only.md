---
title: The core must accept cross-check prices only from a connector whose manifest declares that it sends them.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-112-1
  - lane:strict
text_ref: design-decisions:D-112#1
origin:
  kind: import
  ref: design-decisions:D-112#1
  recorded_at: '2026-10-29T10:54:59.000Z'
semantic_text: The core must accept cross-check prices only from a connector whose manifest declares that it sends them.
logic_claims:
  - CLAIM-6B088C7AD2B37B58
semantic_clauses:
  - The core must accept cross-check prices only from a connector whose manifest declares that it sends them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 717218206a2ddc208b8dc76f35045563e3af11b3ee82b62d9cb84aa50af9e452
semantic_inventory:
  - claim_key: CLAIM-6B088C7AD2B37B58
    claim_text: The core must accept cross-check prices only from a connector whose manifest declares that it sends them
    role: normative
    status: modeled
    span:
      start: 0
      end: 104
id: REQ-ingest-cross-check-prices-declared-only
type: req
---
The core must accept cross-check prices only from a connector whose manifest declares that it sends them.

## Context

D-22 makes IBKR Flex prices a cross-check of the market-data prices, and each plugin's manifest declares what it may do. D-112 lets only a connector whose manifest declares it send cross-check prices. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> Only a connector whose manifest declares it may send cross-check prices. The owner can acknowledge a flagged check, which is recorded append-only, and the UI shows the latest check per instrument with its history.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-112#1
