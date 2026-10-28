---
title: OpenFief's only custom resources must be Plugin and PluginGrant; fiefs, intents, approvals and the ledger must live in PostgreSQL behind the core.
status: open
text_ref: design-decisions:D-55#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-55-1
  - lane:strict
semantic_text: OpenFief's only custom resources must be Plugin and PluginGrant; fiefs, intents, approvals and the ledger must live in PostgreSQL behind the core.
logic_claims:
  - CLAIM-A469333E6D8269D0
semantic_clauses:
  - OpenFief's only custom resources must be Plugin and PluginGrant; fiefs, intents, approvals and the ledger must live in PostgreSQL behind the core.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ffabbeaa1e5516f2a519a0a59bc70dd5d70f3d46c6db1d59072728a93beb8f6d
semantic_inventory:
  - claim_key: CLAIM-A469333E6D8269D0
    claim_text: OpenFief's only custom resources must be Plugin and PluginGrant; fiefs, intents, approvals and the ledger must live in PostgreSQL behind the core
    role: normative
    status: modeled
    span:
      start: 0
      end: 145
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-55#1
  recorded_at: '2026-10-28T06:09:20.356Z'
id: REQ-plugins-only-plugin-crds
type: req
---
OpenFief's only custom resources must be Plugin and PluginGrant; fiefs, intents, approvals and the ledger must live in PostgreSQL behind the core.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. Standard NetworkPolicy matches IP addresses, not hostnames, so outbound traffic goes through a proxy that checks hostnames.

## Source

> The operator manages plugin workloads only and never touches money. Fiefs, intents, approvals and the ledger live in PostgreSQL behind the core, never in custom resources or etcd: etcd is not a transactional ledger, and GR-6's guarantee must not depend on who has Kubernetes API access. The only custom resources are `Plugin` and `PluginGrant`.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-55#1
