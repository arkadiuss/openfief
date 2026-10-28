---
title: An IBKR Flex Withholding Tax row must be booked with the dividend that has the same actionID, or when either lacks one, the same ISIN, currency and date; an unmatched, ambiguous or positive withholding row must become an unsupported ingest problem.
status: open
text_ref: docs/DESIGN.md#D-19
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-19
  - lane:strict
semantic_text: An IBKR Flex Withholding Tax row must be booked with the dividend that has the same actionID, or when either lacks one, the same ISIN, currency and date; an unmatched, ambiguous or positive withholding row must become an unsupported ingest problem.
logic_claims:
  - CLAIM-5BCEDC3E9D48F7B9
semantic_clauses:
  - An IBKR Flex Withholding Tax row must be booked with the dividend that has the same actionID, or when either lacks one, the same ISIN, currency and date; an unmatched, ambiguous or positive withholding row must become an unsupported ingest problem.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2804646740d4afe3257c0592d83ce5e3c9aadbff1da0febc9939e90d5a0fe91c
semantic_inventory:
  - claim_key: CLAIM-5BCEDC3E9D48F7B9
    claim_text: An IBKR Flex Withholding Tax row must be booked with the dividend that has the same actionID, or when either lacks one, the same ISIN, currency and date; an unmatched, ambiguous or positive withholding row must become an unsupported ingest problem
    role: normative
    status: modeled
    span:
      start: 0
      end: 247
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:55.746Z'
id: REQ-ibkrflex-withholding-pairing
type: req
---
An IBKR Flex Withholding Tax row must be booked with the dividend that has the same actionID, or when either lacks one, the same ISIN, currency and date; an unmatched, ambiguous or positive withholding row must become an unsupported ingest problem.

## Context

IBKR reports a dividend and the tax withheld from it as separate cash transaction rows, while the ledger books a dividend as one double-entry transaction across cash, withholding tax and the holding (D-19). Pairing by actionID is exact when IBKR provides it; the ISIN, currency and date fallback covers queries without that column. Withholding refunds and adjustments in later statements are not supported yet. Modeled by Claude Code for the M0 IBKR Flex connector task; the author has not reviewed it yet. The Flex schema is IBKR's public Flex Query reference, not yet verified against a real account (OQ-22).

## Source

> Double-entry postings across cash, position, fee, withholding-tax and FX accounts.

docs/DESIGN.md D-19
