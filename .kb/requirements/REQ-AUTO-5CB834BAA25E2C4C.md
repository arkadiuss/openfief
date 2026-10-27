---
title: Amounts must always keep their original currency.
status: open
text_ref: design-decisions:D-20#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-20-4
  - lane:strict
semantic_text: Amounts must always keep their original currency.
logic_claims:
  - CLAIM-45DEEDE1E49D6BCB
semantic_clauses:
  - Amounts must always keep their original currency.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2b79eee7bf55e4b152b2cb4b189d25ab9ed140a7677e47dd2f9fe7a1372c7d4d
semantic_inventory:
  - claim_key: CLAIM-45DEEDE1E49D6BCB
    claim_text: Amounts must always keep their original currency
    role: normative
    status: modeled
    span:
      start: 0
      end: 48
origin:
  kind: import
  ref: design-decisions:D-20#4
  recorded_at: '2026-10-27T20:12:36.901Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-5CB834BAA25E2C4C
type: req
---
Amounts must always keep their original currency.

## Source

> NBP table A rates convert every amount to PLN in M0, behind a swappable rate-provider interface; NBP is the default FX data plugin. The Polish jurisdiction pack pins NBP for tax whatever the UI uses. Original currencies are always kept.

Source: OpenFief author decisions (DESIGN.md section 10) - D-20#4
