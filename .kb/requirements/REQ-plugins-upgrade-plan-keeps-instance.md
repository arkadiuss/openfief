---
title: An upgrade must keep the instance's name, its config unless the owner gives a new one, and its in-cluster service while the new manifest asks for one, grant exactly what the new version requests, and report what that adds and removes.
status: open
text_ref: docs/DESIGN.md#D-90
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-90
  - lane:strict
semantic_text: An upgrade must keep the instance's name, its config unless the owner gives a new one, and its in-cluster service while the new manifest asks for one, grant exactly what the new version requests, and report what that adds and removes.
logic_claims:
  - CLAIM-BF35988B00FCFA12
semantic_clauses:
  - An upgrade must keep the instance's name, its config unless the owner gives a new one, and its in-cluster service while the new manifest asks for one, grant exactly what the new version requests, and report what that adds and removes.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c3aa41f4741defbe2334e931549515f3804c7b3ae17c20f96a7fa2d4d3547460
semantic_inventory:
  - claim_key: CLAIM-BF35988B00FCFA12
    claim_text: An upgrade must keep the instance's name, its config unless the owner gives a new one, and its in-cluster service while the new manifest asks for one, grant exactly what the new version requests, and report what that adds and removes
    role: exception
    status: modeled
    span:
      start: 0
      end: 233
origin:
  kind: agent
  recorded_at: '2026-10-29T17:04:15.661Z'
id: REQ-plugins-upgrade-plan-keeps-instance
type: req
---
An upgrade must keep the instance's name, its config unless the owner gives a new one, and its in-cluster service while the new manifest asks for one, grant exactly what the new version requests, and report what that adds and removes.

## Context

D-90 has the upgrade show how requested permissions differ; the grant approving exactly what was shown follows the install (D-46). Keeping config and in-cluster service is the narrowest reading. Refines REQ-plugins-cli-upgrade.

## Source

> [D-90] The CLI refuses a catalog name that matches several versions and lists them instead of picking one, and refuses to install over an existing instance. `openfief plugin upgrade <instance> <version>` verifies the new image, shows how its requested permissions differ, and applies the upgrade only when the owner confirms.

docs/DESIGN.md#D-90
