---
title: Every source's market-data values must be stored, so the lower-ranked ones serve as cross-checks.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-94-3
  - lane:strict
text_ref: design-decisions:D-94#3
origin:
  kind: import
  ref: design-decisions:D-94#3
  recorded_at: '2026-10-29T08:50:08.924Z'
semantic_text: Every source's market-data values must be stored, so the lower-ranked ones serve as cross-checks.
logic_claims:
  - CLAIM-0CDDB543BEB3F85C
semantic_clauses:
  - Every source's market-data values must be stored, so the lower-ranked ones serve as cross-checks.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: ac68cf1dce99178872d0be3965d50f1da73fefd9e74d4e9ae6470c2e2ba0d255
semantic_inventory:
  - claim_key: CLAIM-0CDDB543BEB3F85C
    claim_text: Every source's market-data values must be stored, so the lower-ranked ones serve as cross-checks
    role: normative
    status: modeled
    span:
      start: 0
      end: 96
id: REQ-valuation-market-data-all-sources-stored
type: req
---
Every source's market-data values must be stored, so the lower-ranked ones serve as cross-checks.

## Context

D-22 uses a second source to cross-check market data, such as IBKR Flex prices against market-data prices (REQ-AUTO-7374F67D85269872). D-94 lets only the highest-ranked source with a value win for each date (REQ-valuation-market-data-highest-rank-wins), so it keeps every source's values rather than overwriting them, and the lower-ranked values stay available as those cross-checks. REQ-valuation-index-levels, which the author has not reviewed, still replaces an earlier level of the same index and date whatever its source. The author recorded D-93 to D-95 after the design session that answered the implementation's questions.

## Source

> The owner ranks the sources of each kind of market data (prices, exchange rates and index readings) in priority order, and for each date the highest-ranked source with a value wins. Every source's values are stored, so the lower-ranked ones serve as cross-checks (D-22). By default the first-party connector for each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-94#3
