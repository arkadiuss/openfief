---
title: The core must serve BondLotService next to the read API to view plugins, reach the bound connector instance through the client its connector discovery builds, answering unavailable until then, and record delivered bond series terms through valuation so imported lots are valued by the bond rules.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The core must serve BondLotService next to the read API to view plugins, reach the bound connector instance through the client its connector discovery builds, answering unavailable until then, and record delivered bond series terms through valuation so imported lots are valued by the bond rules.
logic_claims:
  - CLAIM-14322D32E2325A58
semantic_clauses:
  - The core must serve BondLotService next to the read API to view plugins, reach the bound connector instance through the client its connector discovery builds, answering unavailable until then, and record delivered bond series terms through valuation so imported lots are valued by the bond rules.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 98c46ae76297b8ee75381db32e42e544c4214d2cf00fae1c7f9960380be6fb08
semantic_inventory:
  - claim_key: CLAIM-14322D32E2325A58
    claim_text: The core must serve BondLotService next to the read API to view plugins, reach the bound connector instance through the client its connector discovery builds, answering unavailable until then, and record delivered bond series terms through valuation so imported lots are valued by the bond rules
    role: normative
    status: modeled
    span:
      start: 0
      end: 295
origin:
  kind: agent
  recorded_at: '2026-10-29T01:54:50.264Z'
id: REQ-core-bond-lot-service
type: req
---
The core must serve BondLotService next to the read API to view plugins, reach the bound connector instance through the client its connector discovery builds, answering unavailable until then, and record delivered bond series terms through valuation so imported lots are valued by the bond rules.

## Context

D-49 routes manual bond-lot entries through the core to the right connector; D-9 says their valuation is computed. Entry needs connector discovery, so it works only in a cluster with connectors on. Modeled by Claude Code for the manual bond lot task; the author has not reviewed it yet.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md#D-49
