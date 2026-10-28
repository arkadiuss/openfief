---
title: 'The demo loader must load the dataset through OpenFief''s real paths: the IBKR history as Flex statements through the ibkrflex connector and the ingest service, seeded at go-live and then one transactions batch per month; the XTB and bond accounts through the ledger''s seeding and import use cases; prices, FX rates, CPI, bond terms and the IKE wrapper through valuation''s input recorder; and the stored valuations through the River revaluation workers, exiting once every demo account is valued through today.'
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: 'The demo loader must load the dataset through OpenFief''s real paths: the IBKR history as Flex statements through the ibkrflex connector and the ingest service, seeded at go-live and then one transactions batch per month; the XTB and bond accounts through the ledger''s seeding and import use cases; prices, FX rates, CPI, bond terms and the IKE wrapper through valuation''s input recorder; and the stored valuations through the River revaluation workers, exiting once every demo account is valued through today.'
logic_claims:
  - CLAIM-39974F0AD4840B58
semantic_clauses:
  - 'The demo loader must load the dataset through OpenFief''s real paths: the IBKR history as Flex statements through the ibkrflex connector and the ingest service, seeded at go-live and then one transactions batch per month; the XTB and bond accounts through the ledger''s seeding and import use cases; prices, FX rates, CPI, bond terms and the IKE wrapper through valuation''s input recorder; and the stored valuations through the River revaluation workers, exiting once every demo account is valued through today.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 933b7b8fab8808f71982712b051e6dd49cc27bf5581738ffd91e3eeef4c4ff3b
semantic_inventory:
  - claim_key: CLAIM-39974F0AD4840B58
    claim_text: 'The demo loader must load the dataset through OpenFief''s real paths: the IBKR history as Flex statements through the ibkrflex connector and the ingest service, seeded at go-live and then one transactions batch per month; the XTB and bond accounts through the ledger''s seeding and import use cases; prices, FX rates, CPI, bond terms and the IKE wrapper through valuation''s input recorder; and the stored valuations through the River revaluation workers, exiting once every demo account is valued through today'
    role: normative
    status: modeled
    span:
      start: 0
      end: 508
origin:
  kind: agent
  recorded_at: '2026-10-28T09:31:31.933Z'
id: REQ-demo-real-pipelines
type: req
---
The demo loader must load the dataset through OpenFief's real paths: the IBKR history as Flex statements through the ibkrflex connector and the ingest service, seeded at go-live and then one transactions batch per month; the XTB and bond accounts through the ledger's seeding and import use cases; prices, FX rates, CPI, bond terms and the IKE wrapper through valuation's input recorder; and the stored valuations through the River revaluation workers, exiting once every demo account is valued through today.

## Context

The author wants to see the real pipelines working on fake data, not rows written straight into tables; each step uses the same use case real data takes (D-23 seeding, D-18 imports, D-36 revaluations). XTB and the bond registry have no connector yet, so their data goes through the ledger's API. Modeled by Claude Code for the demo-data task: the author asked to see OpenFief on the Hetzner demo with real pipelines on fake data; the author has not reviewed this requirement yet.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time.

docs/DESIGN.md D-36
