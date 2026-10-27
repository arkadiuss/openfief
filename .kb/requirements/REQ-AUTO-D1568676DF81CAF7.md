---
title: The FX rate provider must be swappable, with NBP as the default FX data plugin.
status: open
text_ref: design-decisions:D-20#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-20-2
  - lane:strict
semantic_text: The FX rate provider must be swappable, with NBP as the default FX data plugin.
logic_claims:
  - CLAIM-B3419798C3470A5B
semantic_clauses:
  - The FX rate provider must be swappable, with NBP as the default FX data plugin.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f87b6ae4aa4a30254bb1142069f9e06e964df8a4859b763d428caee86f247687
semantic_inventory:
  - claim_key: CLAIM-B3419798C3470A5B
    claim_text: The FX rate provider must be swappable, with NBP as the default FX data plugin
    role: normative
    status: modeled
    span:
      start: 0
      end: 78
origin:
  kind: import
  ref: design-decisions:D-20#2
  recorded_at: '2026-10-27T20:12:35.806Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-D1568676DF81CAF7
type: req
---
The FX rate provider must be swappable, with NBP as the default FX data plugin.

## Source

> NBP table A rates convert every amount to PLN in M0, behind a swappable rate-provider interface; NBP is the default FX data plugin. The Polish jurisdiction pack pins NBP for tax whatever the UI uses. Original currencies are always kept.

Source: OpenFief author decisions (DESIGN.md section 10) - D-20#2
