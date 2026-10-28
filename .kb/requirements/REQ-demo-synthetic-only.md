---
title: 'The demo dataset must hold no real account data: every instrument must have a synthetic ISIN with a valid check digit and the TESTD marker after its country code, IBKR statements must name only the placeholder account U0000000, account IDs must start with demo-, and the only real figures may be public ones: bond series terms, GUS CPI readings and NBP table A rates.'
status: open
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
semantic_text: 'The demo dataset must hold no real account data: every instrument must have a synthetic ISIN with a valid check digit and the TESTD marker after its country code, IBKR statements must name only the placeholder account U0000000, account IDs must start with demo-, and the only real figures may be public ones: bond series terms, GUS CPI readings and NBP table A rates.'
logic_claims:
  - CLAIM-803FF4F690854ED9
semantic_clauses:
  - 'The demo dataset must hold no real account data: every instrument must have a synthetic ISIN with a valid check digit and the TESTD marker after its country code, IBKR statements must name only the placeholder account U0000000, account IDs must start with demo-, and the only real figures may be public ones: bond series terms, GUS CPI readings and NBP table A rates.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 68cc27f7ec2f78dec8bac13513e4bfb6335438a0b09159ce2bad86f9f8541d72
semantic_inventory:
  - claim_key: CLAIM-803FF4F690854ED9
    claim_text: 'The demo dataset must hold no real account data: every instrument must have a synthetic ISIN with a valid check digit and the TESTD marker after its country code, IBKR statements must name only the placeholder account U0000000, account IDs must start with demo-, and the only real figures may be public ones: bond series terms, GUS CPI readings and NBP table A rates'
    role: normative
    status: modeled
    span:
      start: 0
      end: 366
origin:
  kind: agent
  recorded_at: '2026-10-28T09:31:24.070Z'
id: REQ-demo-synthetic-only
type: req
---
The demo dataset must hold no real account data: every instrument must have a synthetic ISIN with a valid check digit and the TESTD marker after its country code, IBKR statements must name only the placeholder account U0000000, account IDs must start with demo-, and the only real figures may be public ones: bond series terms, GUS CPI readings and NBP table A rates.

## Context

GR-12 puts privacy first and AGENTS.md bans real financial data anywhere; a demo is shown to other people, so it must not resemble anyone's portfolio. Public statistics and bond terms are not personal data and make the bond valuations real. Modeled by Claude Code for the demo-data task: the author asked to see OpenFief on the Hetzner demo with real pipelines on fake data; the author has not reviewed this requirement yet.

## Source

> The privacy and security model comes first, designed per use case.

docs/DESIGN.md GR-12; AGENTS.md: "No real financial data, account numbers or credentials anywhere: code, tests, recordings, logs, commits."
