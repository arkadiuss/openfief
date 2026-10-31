---
title: For a plugin's granted in-cluster service, the owner's install must create, in that service's namespace, a Role and RoleBinding that let the operator read that one Service.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-122-1
  - lane:strict
text_ref: design-decisions:D-122#1
origin:
  kind: import
  ref: design-decisions:D-122#1
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: For a plugin's granted in-cluster service, the owner's install must create, in that service's namespace, a Role and RoleBinding that let the operator read that one Service.
logic_claims:
  - CLAIM-B45D5FE7058EF0EA
semantic_clauses:
  - For a plugin's granted in-cluster service, the owner's install must create, in that service's namespace, a Role and RoleBinding that let the operator read that one Service.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c87b82e3e56ff9aa9fc02ba0381cb5c9cf29fc5aeaf4e0153fce3dc6dcb9b7f6
semantic_inventory:
  - claim_key: CLAIM-B45D5FE7058EF0EA
    claim_text: For a plugin's granted in-cluster service, the owner's install must create, in that service's namespace, a Role and RoleBinding that let the operator read that one Service
    role: normative
    status: modeled
    span:
      start: 0
      end: 171
id: REQ-plugins-in-cluster-service-read-role
type: req
---
For a plugin's granted in-cluster service, the owner's install must create, in that service's namespace, a Role and RoleBinding that let the operator read that one Service.

## Context

D-74 has the operator add a NetworkPolicy egress rule to an in-cluster service's pods and hand the plugin the service's ClusterIP (REQ-operator-in-cluster-service-route), which means reading that Service in another namespace, while D-15 confines the operator to a Role in `openfief`. The implementation asked how the operator reads it (question #2 on docs/implementation-questions.md); the author kept the install-created Role and RoleBinding, scoped to that one Service, over recording the Service in the grant or a ClusterRole, and recorded it as D-122 on 2026-10-31. The install runs with the owner's kubeconfig (D-46).

## Source

> For D-74's in-cluster service, the owner's install creates, in that service's namespace, a Role and RoleBinding that let the operator read that one Service, and uninstall deletes them. The operator's permissions otherwise stay inside `openfief` (D-15).

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-122#1
