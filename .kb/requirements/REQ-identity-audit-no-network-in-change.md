---
title: 'No change that is kept together with its audit entry may make a network call inside its transaction: a call to a connector, a model provider, SpiceDB or the cluster must be made before the transaction opens, with its answer handed in, or after the transaction has committed.'
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:09.353Z'
semantic_text: 'No change that is kept together with its audit entry may make a network call inside its transaction: a call to a connector, a model provider, SpiceDB or the cluster must be made before the transaction opens, with its answer handed in, or after the transaction has committed.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 04baddf2dd62a815a768b5bbac189bbdbefde7d9b38bd06d6745e4d07d21c5a8
semantic_inventory:
  - claim_key: CLAIM-AADDAABB9ACED103
    claim_text: 'No change that is kept together with its audit entry may make a network call inside its transaction: a call to a connector, a model provider, SpiceDB or the cluster must be made before the transaction opens, with its answer handed in, or after the transaction has committed'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 273
logic_claims:
  - CLAIM-AADDAABB9ACED103
id: REQ-identity-audit-no-network-in-change
type: req
---
No change that is kept together with its audit entry may make a network call inside its transaction: a call to a connector, a model provider, SpiceDB or the cluster must be made before the transaction opens, with its answer handed in, or after the transaction has committed.

## Context

The change's stores share one connection and hold their locks for as long as the change runs, so a slow call there would hold the database up. The owner's statement upload and bond lot entry and deletion ask the connector first (PrepareFile, PrepareBondLots) and only then run the import (Commit); what a change tells SpiceDB afterwards, such as an account's placement, is deferred until the outermost change commits. Every client the core builds over the network refuses a call made with a change's context, and an architecture test keeps the PostgreSQL adapters from importing a network client.

## Source

> NO NETWORK I/O INSIDE A CHANGE'S TRANSACTION (connector, model provider, SpiceDB, any Connect/HTTP client). Audit every outbox.Within / audited change path; move any such call before Begin and pass results in.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18