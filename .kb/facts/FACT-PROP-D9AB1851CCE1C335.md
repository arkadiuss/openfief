---
title: Cli Prints Ingress Rule For Service Namespace = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-74-5
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-74#5
fact_kind: property_value
subject_key: plugins.in_cluster_service
property_key: cli_prints_ingress_rule_for_service_namespace
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:plugins.in_cluster_service:cli_prints_ingress_rule_for_service_namespace:eq:true
claim_key: CLAIM-C05556EC595A60D7
claim_text: When the granted in-cluster service's namespace has its own NetworkPolicies, the CLI must print the ingress rule to add there.
origin:
  kind: agent
  recorded_at: '2026-10-29T07:38:20.699Z'
id: FACT-PROP-D9AB1851CCE1C335
type: fact
---
