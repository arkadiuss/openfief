---
title: Bond lots and bond series terms must cross the connector contract as exact decimal strings and ISO dates, and the core must tell a connector's refusal, its lack of bond lot support and its unavailability apart.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: Bond lots and bond series terms must cross the connector contract as exact decimal strings and ISO dates, and the core must tell a connector's refusal, its lack of bond lot support and its unavailability apart.
logic_claims:
  - CLAIM-308354281F68B3F2
semantic_clauses:
  - Bond lots and bond series terms must cross the connector contract as exact decimal strings and ISO dates, and the core must tell a connector's refusal, its lack of bond lot support and its unavailability apart.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 213854e3c3d4f0996ab717b4b8838cccbb0177f99b3e599ebf97fdb03ac4c3e1
semantic_inventory:
  - claim_key: CLAIM-308354281F68B3F2
    claim_text: Bond lots and bond series terms must cross the connector contract as exact decimal strings and ISO dates, and the core must tell a connector's refusal, its lack of bond lot support and its unavailability apart
    role: normative
    status: modeled
    span:
      start: 0
      end: 209
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:34.221Z'
id: REQ-ingest-bond-lot-messages
type: req
---
Bond lots and bond series terms must cross the connector contract as exact decimal strings and ISO dates, and the core must tell a connector's refusal, its lack of bond lot support and its unavailability apart.

## Context

D-49 and D-11: the core calls connectors over protobuf contracts. ImportBondLots and Delivery.bond_terms are additions to openfief.connector.v1, so buf breaking passes. Modeled by Claude Code for the manual bond lot task; the author has not reviewed it yet.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md#D-49
