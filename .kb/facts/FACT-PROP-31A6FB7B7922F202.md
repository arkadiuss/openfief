---
title: Uninstall Deletes Operator Read Role = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-122-2
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-122#2
fact_kind: property_value
subject_key: plugins.in_cluster_service
property_key: uninstall_deletes_operator_read_role
operator: eq
value_type: bool
value_bool: true
canonical_key: kb-requirements-req-plugins-in-cluster-service-role-removed-md:plugins.in_cluster_service:uninstall_deletes_operator_read_role:eq:true
claim_key: CLAIM-861DA3614DCAF06E
claim_text: Uninstalling a plugin with a granted in-cluster service must delete the Role and RoleBinding its install created in that service's namespace.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:09:06.243Z'
id: FACT-PROP-31A6FB7B7922F202
type: fact
---
