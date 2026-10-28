---
title: The connector contract, openfief.connector.v1.ConnectorService, must carry amounts, prices, rates and quantities as exact decimal strings, dates as YYYY-MM-DD strings and currencies as ISO 4217 codes, and every Sync or ImportFile call must carry a batch ID the core generates, which the connector's delivery echoes.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The connector contract, openfief.connector.v1.ConnectorService, must carry amounts, prices, rates and quantities as exact decimal strings, dates as YYYY-MM-DD strings and currencies as ISO 4217 codes, and every Sync or ImportFile call must carry a batch ID the core generates, which the connector's delivery echoes.
logic_claims:
  - CLAIM-ED878374C3A9623B
semantic_clauses:
  - The connector contract, openfief.connector.v1.ConnectorService, must carry amounts, prices, rates and quantities as exact decimal strings, dates as YYYY-MM-DD strings and currencies as ISO 4217 codes, and every Sync or ImportFile call must carry a batch ID the core generates, which the connector's delivery echoes.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e779dcf41d764f630edf3a7db38e988fedc1b2f7a85e1d9e6004bc8dd4fa59bf
semantic_inventory:
  - claim_key: CLAIM-ED878374C3A9623B
    claim_text: The connector contract, openfief.connector.v1.ConnectorService, must carry amounts, prices, rates and quantities as exact decimal strings, dates as YYYY-MM-DD strings and currencies as ISO 4217 codes, and every Sync or ImportFile call must carry a batch ID the core generates, which the connector's delivery echoes
    role: normative
    status: modeled
    span:
      start: 0
      end: 314
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:11.379Z'
id: REQ-connector-contract-messages
type: req
---
The connector contract, openfief.connector.v1.ConnectorService, must carry amounts, prices, rates and quantities as exact decimal strings, dates as YYYY-MM-DD strings and currencies as ISO 4217 codes, and every Sync or ImportFile call must carry a batch ID the core generates, which the connector's delivery echoes.

## Context

D-49 has connectors serve a Connect service for syncing and importing files, with the core owning schedules, retries and import batches; a core-generated batch ID lets the core retry a batch safely, since re-imports are idempotent (D-18). Decimal strings keep money exact (no floating point). The contract and its generated Go code are Apache-2.0 (D-5). Refines REQ-ingest-core-calls-connectors and REQ-AUTO-BC327C2200FC820F. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md D-49
