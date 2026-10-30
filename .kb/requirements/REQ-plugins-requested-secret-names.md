---
title: A plugin instance must request, and its workload must mount, the Secret <instance>-<role> for each Secret role its manifest names, and a role must be a DNS label.
status: open
text_ref: docs/DESIGN.md#D-89
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-89
  - lane:strict
semantic_text: A plugin instance must request, and its workload must mount, the Secret <instance>-<role> for each Secret role its manifest names, and a role must be a DNS label.
logic_claims:
  - CLAIM-CB4583049B0C1862
semantic_clauses:
  - A plugin instance must request, and its workload must mount, the Secret <instance>-<role> for each Secret role its manifest names, and a role must be a DNS label.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 36690b706c5bb7c6557cda7877e1428d77bbe5abbd6303442461a17a3f0b9b27
semantic_inventory:
  - claim_key: CLAIM-CB4583049B0C1862
    claim_text: A plugin instance must request, and its workload must mount, the Secret <instance>-<role> for each Secret role its manifest names, and a role must be a DNS label
    role: normative
    status: modeled
    span:
      start: 0
      end: 161
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:32.438Z'
id: REQ-plugins-requested-secret-names
type: req
---
A plugin instance must request, and its workload must mount, the Secret <instance>-<role> for each Secret role its manifest names, and a role must be a DNS label.

## Context

D-89 names Secrets by role in the manifest and has the operator mount <instance>-<role>. The grant names the Secret in full so the owner approves exactly what is mounted; a DNS-label role keeps the name a valid Secret name. Refines REQ-plugins-manifest-secret-roles and REQ-operator-instance-secret-names.

## Source

> [D-89] A manifest names each Secret by its role, such as `flex-token`, and the operator mounts the Secret named `<instance>-<role>`, so two instances of one image never share a Secret.

docs/DESIGN.md#D-89
