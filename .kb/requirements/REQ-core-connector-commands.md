---
title: In a Kubernetes cluster core serve must run the connector scheduler in the background when connectors are on, the default in a pod, without affecting readiness, and wait for it on shutdown; it must import deliveries through the ledger and record their prices, FX rates and index readings through valuation's input use cases; core connector bind must bind an instance to a stored account and core connector sync must sync one instance at once.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: In a Kubernetes cluster core serve must run the connector scheduler in the background when connectors are on, the default in a pod, without affecting readiness, and wait for it on shutdown; it must import deliveries through the ledger and record their prices, FX rates and index readings through valuation's input use cases; core connector bind must bind an instance to a stored account and core connector sync must sync one instance at once.
logic_claims:
  - CLAIM-D9766EA1CD6588CE
semantic_clauses:
  - In a Kubernetes cluster core serve must run the connector scheduler in the background when connectors are on, the default in a pod, without affecting readiness, and wait for it on shutdown; it must import deliveries through the ledger and record their prices, FX rates and index readings through valuation's input use cases; core connector bind must bind an instance to a stored account and core connector sync must sync one instance at once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d4482a4e149ec3d3a24bf16e48aa29d669e8e0d6b05c070c0e0bf0dcd82bdb1e
semantic_inventory:
  - claim_key: CLAIM-D9766EA1CD6588CE
    claim_text: In a Kubernetes cluster core serve must run the connector scheduler in the background when connectors are on, the default in a pod, without affecting readiness, and wait for it on shutdown; it must import deliveries through the ledger and record their prices, FX rates and index readings through valuation's input use cases; core connector bind must bind an instance to a stored account and core connector sync must sync one instance at once
    role: normative
    status: modeled
    span:
      start: 0
      end: 441
origin:
  kind: agent
  recorded_at: '2026-10-28T17:23:04.898Z'
id: REQ-core-connector-commands
type: req
---
In a Kubernetes cluster core serve must run the connector scheduler in the background when connectors are on, the default in a pod, without affecting readiness, and wait for it on shutdown; it must import deliveries through the ledger and record their prices, FX rates and index readings through valuation's input use cases; core connector bind must bind an instance to a stored account and core connector sync must sync one instance at once.

## Context

D-49 puts schedules, retries and batches in the core. The orchestrator asked for a --connectors flag on by default in a cluster, a connector sync command for the demo, a connector bind command for connector bindings, and /readyz unaffected by connector failures. Index readings become valuation's monthly readings for the month of their date. Modeled by Claude Code for the core-connectors task (orchestrator brief: CoreTokens, discovery, cmd/core wiring); the author has not reviewed this requirement yet.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md D-49
