---
title: A connector plugin must write its mapping, FX rates and ingest mode as connector contract messages that keep every amount, rate and quantity exact as a decimal string, every date as YYYY-MM-DD, and must refuse a value with no contract form instead of sending it.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: A connector plugin must write its mapping, FX rates and ingest mode as connector contract messages that keep every amount, rate and quantity exact as a decimal string, every date as YYYY-MM-DD, and must refuse a value with no contract form instead of sending it.
logic_claims:
  - CLAIM-DCDFF191AB58A6CD
semantic_clauses:
  - A connector plugin must write its mapping, FX rates and ingest mode as connector contract messages that keep every amount, rate and quantity exact as a decimal string, every date as YYYY-MM-DD, and must refuse a value with no contract form instead of sending it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d6aa42b8e47b4468426807751eee1dd14db4a6d6572a470176eb9d23e4c3f571
semantic_inventory:
  - claim_key: CLAIM-DCDFF191AB58A6CD
    claim_text: A connector plugin must write its mapping, FX rates and ingest mode as connector contract messages that keep every amount, rate and quantity exact as a decimal string, every date as YYYY-MM-DD, and must refuse a value with no contract form instead of sending it
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:08.409Z'
id: REQ-ingest-delivery-contract-messages
type: req
---
A connector plugin must write its mapping, FX rates and ingest mode as connector contract messages that keep every amount, rate and quantity exact as a decimal string, every date as YYYY-MM-DD, and must refuse a value with no contract form instead of sending it.

## Context

D-49 makes connectors plugins that the core calls; the first-party ones reuse the in-core adapters (ibkrflex, nbp), which produce domain values. The core's connectrpc adapter refuses anything that is not a plain decimal, an ISO date or a known enum, so the plugin side must write exactly that, and money must never pass through floating point (AGENTS.md hard rules, D-20). Refines REQ-ingest-core-calls-connectors and REQ-connector-contract-messages. Modeled by Claude Code for the M0 plugin binaries task (agent A: SDK verifier, operator key publishing, first-party plugins); the author has not reviewed it yet.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md D-49
