---
title: Host From Https Url = true
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
subject_key: plugins.config_hosts
property_key: host_from_https_url
canonical_key: mcp-kibi-model-requirement:plugins.config_hosts:host_from_https_url:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C713554AE3C1C86B
claim_text: A config field a manifest names in egressFromConfig must give the host and port of the https URL it holds, lower-cased, nothing when it is absent or empty, and make the Plugin invalid when it is not an https URL of one hostname without a user part.
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:51.926Z'
id: FACT-PROP-97A9374A1D3C67C2
type: fact
---
