---
title: Uninstalling a plugin with a granted in-cluster service must delete the Role and RoleBinding its install created in that service's namespace.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-122-2
  - lane:strict
text_ref: design-decisions:D-122#2
origin:
  kind: import
  ref: design-decisions:D-122#2
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: Uninstalling a plugin with a granted in-cluster service must delete the Role and RoleBinding its install created in that service's namespace.
logic_claims:
  - CLAIM-861DA3614DCAF06E
semantic_clauses:
  - Uninstalling a plugin with a granted in-cluster service must delete the Role and RoleBinding its install created in that service's namespace.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2e9ce22f33eaaf402ad94cdaeeb7e0656bd4562e4ddb31bffa6e178f3838f254
semantic_inventory:
  - claim_key: CLAIM-861DA3614DCAF06E
    claim_text: Uninstalling a plugin with a granted in-cluster service must delete the Role and RoleBinding its install created in that service's namespace
    role: normative
    status: modeled
    span:
      start: 0
      end: 140
id: REQ-plugins-in-cluster-service-role-removed
type: req
---
Uninstalling a plugin with a granted in-cluster service must delete the Role and RoleBinding its install created in that service's namespace.

## Context

The install creates that Role and RoleBinding so the operator can read the one Service D-74 grants (REQ-plugins-in-cluster-service-read-role); removing them on uninstall leaves the operator no permission outside `openfief` once no plugin needs one, as D-15 intends. Uninstall already deletes an instance's workload, grant, secrets and egress allow-list (REQ-plugins-uninstall-keeps-data). The author recorded this as D-122 on 2026-10-31, answering question #2 on docs/implementation-questions.md.

## Source

> For D-74's in-cluster service, the owner's install creates, in that service's namespace, a Role and RoleBinding that let the operator read that one Service, and uninstall deletes them. The operator's permissions otherwise stay inside `openfief` (D-15).

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-122#2
