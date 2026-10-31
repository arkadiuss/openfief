---
title: Permissions Outside Openfief Namespace = only reading each plugin's granted in-cluster Service
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-122-3
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-122#3
fact_kind: property_value
subject_key: plugins.operator
property_key: permissions_outside_openfief_namespace
operator: eq
value_type: string
value_string: only reading each plugin's granted in-cluster Service
canonical_key: kb-requirements-req-operator-namespace-scoped-except-service-read-md:plugins.operator:permissions_outside_openfief_namespace:eq:only_reading_each_plugin_s_granted_in-cluster_service
claim_key: CLAIM-22CEE3F79BCA2434
claim_text: Apart from reading each plugin's granted in-cluster Service, the operator must hold permissions only within the openfief namespace.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:09:27.066Z'
id: FACT-PROP-4CEDB7310DB200A1
type: fact
---
