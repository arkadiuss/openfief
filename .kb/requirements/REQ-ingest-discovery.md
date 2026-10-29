---
title: The core must sync only Plugin instances that are Ready, not being deleted, allowed the connector capability in their effective permissions, and served by a Service with a ClusterIP and a connect port, must call each at that ClusterIP and port, and must follow changes to Plugins and Services through watches.
status: open
text_ref: docs/DESIGN.md#D-48
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-48
  - lane:strict
semantic_text: The core must sync only Plugin instances that are Ready, not being deleted, allowed the connector capability in their effective permissions, and served by a Service with a ClusterIP and a connect port, must call each at that ClusterIP and port, and must follow changes to Plugins and Services through watches.
logic_claims:
  - CLAIM-7AB1AE162A3ECEB1
semantic_clauses:
  - The core must sync only Plugin instances that are Ready, not being deleted, allowed the connector capability in their effective permissions, and served by a Service with a ClusterIP and a connect port, must call each at that ClusterIP and port, and must follow changes to Plugins and Services through watches.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 90839e2419b7308a652df58a5a4ce62d6a0bea39fa65ccf8e76ab9140d494562
semantic_inventory:
  - claim_key: CLAIM-7AB1AE162A3ECEB1
    claim_text: The core must sync only Plugin instances that are Ready, not being deleted, allowed the connector capability in their effective permissions, and served by a Service with a ClusterIP and a connect port, must call each at that ClusterIP and port, and must follow changes to Plugins and Services through watches
    role: normative
    status: modeled
    span:
      start: 0
      end: 308
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:56.347Z'
id: REQ-ingest-discovery
type: req
---
The core must sync only Plugin instances that are Ready, not being deleted, allowed the connector capability in their effective permissions, and served by a Service with a ClusterIP and a connect port, must call each at that ClusterIP and port, and must follow changes to Plugins and Services through watches.

## Context

Instances are named installs (D-48) that come and go (D-58, D-59); the operator records readiness and effective permissions in the Plugin's status and creates a ClusterIP Service per plugin, and plugins have no DNS, so the core calls the ClusterIP. Reading the Plugin as unstructured data keeps ingest independent of the operator's Go types. Modeled by Claude Code for the core-connectors task (orchestrator brief: CoreTokens, discovery, cmd/core wiring); the author has not reviewed this requirement yet.

## Source

> Each install is a separately named instance with its own config, secrets, grant and egress allow-list.

docs/DESIGN.md D-48

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md D-49
