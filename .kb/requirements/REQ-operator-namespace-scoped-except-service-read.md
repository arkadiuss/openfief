---
title: Apart from reading each plugin's granted in-cluster Service, the operator must hold permissions only within the openfief namespace.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-122-3
  - lane:strict
text_ref: design-decisions:D-122#3
origin:
  kind: import
  ref: design-decisions:D-122#3
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: Apart from reading each plugin's granted in-cluster Service, the operator must hold permissions only within the openfief namespace.
logic_claims:
  - CLAIM-22CEE3F79BCA2434
semantic_clauses:
  - Apart from reading each plugin's granted in-cluster Service, the operator must hold permissions only within the openfief namespace.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 585e0cce24ea326f7ae50f5e4545664f4a03d3487735ded05803c634388619e5
semantic_inventory:
  - claim_key: CLAIM-22CEE3F79BCA2434
    claim_text: Apart from reading each plugin's granted in-cluster Service, the operator must hold permissions only within the openfief namespace
    role: exception
    status: modeled
    span:
      start: 0
      end: 130
id: REQ-operator-namespace-scoped-except-service-read
type: req
---
Apart from reading each plugin's granted in-cluster Service, the operator must hold permissions only within the openfief namespace.

## Why this changed

REQ-AUTO-BBBDDCF207E884A1, from D-15, confined every operator permission to the `openfief` namespace. D-74 has the operator read an in-cluster service in another namespace to route a plugin to its pods, so the implementation asked how (question #2 on docs/implementation-questions.md). On 2026-10-31 the author decided D-122: the owner's install grants the operator read access to that one Service in its own namespace, and the operator's permissions otherwise stay inside `openfief`.

## Context

D-15 makes the operator namespace-scoped, a Role in `openfief` only, so a compromised operator reaches nothing else in the cluster. D-122 is the one narrow exception: a Role and RoleBinding the install creates and uninstall deletes, letting the operator read exactly the Service a plugin's grant approved (REQ-plugins-in-cluster-service-read-role, REQ-plugins-in-cluster-service-role-removed). OpenFief's cluster-wide permissions stay the core's and the egress proxy's ClusterRoles (REQ-core-cluster-wide-permissions).

## Source

> For D-74's in-cluster service, the owner's install creates, in that service's namespace, a Role and RoleBinding that let the operator read that one Service, and uninstall deletes them. The operator's permissions otherwise stay inside `openfief` (D-15).

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-122#3
