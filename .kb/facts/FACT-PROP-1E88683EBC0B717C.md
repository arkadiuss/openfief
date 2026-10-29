---
title: Config Host Or Cluster Service = true
status: active
text_ref: docs/DESIGN.md#D-75
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-75
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.ghostfolio_plugin
property_key: config_host_or_cluster_service
canonical_key: mcp-kibi-model-requirement:plugins.ghostfolio_plugin:config_host_or_cluster_service:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-65B6293BE303885D
claim_text: The Ghostfolio plugin's manifest must ask only for the view capability, its security-token Secret, the host of its url config field and an in-cluster service; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach an external Ghostfolio only through the egress proxy and an in-cluster one only at the address the operator passes, and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.
origin:
  kind: agent
  recorded_at: '2026-10-29T15:13:07.007Z'
id: FACT-PROP-1E88683EBC0B717C
type: fact
---
