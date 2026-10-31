---
title: A ticker entry must be booked with the ISIN the owner confirmed for its ticker, and wait as an ingest problem until the owner confirms one.
status: open
text_ref: docs/DESIGN.md#D-116
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-116
  - lane:strict
semantic_text: A ticker entry must be booked with the ISIN the owner confirmed for its ticker, and wait as an ingest problem until the owner confirms one.
logic_claims:
  - CLAIM-CA0492CD5AB6399A
semantic_clauses:
  - A ticker entry must be booked with the ISIN the owner confirmed for its ticker, and wait as an ingest problem until the owner confirms one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4464a077e075dfc49fa7d67d753d73cda478bbb544b614035adb9915889a103c
semantic_inventory:
  - claim_key: CLAIM-CA0492CD5AB6399A
    claim_text: A ticker entry must be booked with the ISIN the owner confirmed for its ticker, and wait as an ingest problem until the owner confirms one
    role: normative
    status: modeled
    span:
      start: 0
      end: 138
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:24.181Z'
id: REQ-ingest-ticker-entries-wait
type: req
---
A ticker entry must be booked with the ISIN the owner confirmed for its ticker, and wait as an ingest problem until the owner confirms one.

## Context

D-116 holds the rows of an unconfirmed ticker as import problems; D-35 needs an ISIN for every instrument. Recorded by the XTB implementation agent (impl/xtb) for the author's XTB work package; default firmness, not reviewed by the author yet.

## Source

> XTB's export names instruments by ticker and market suffix and carries no ISIN (D-35). For each new ticker the market-data connector looks up the ISIN by ticker and market and proposes it, and the owner confirms it once in the UI, or enters it when the lookup finds none. Rows of a ticker not confirmed yet wait as import problems.

docs/DESIGN.md#D-116
