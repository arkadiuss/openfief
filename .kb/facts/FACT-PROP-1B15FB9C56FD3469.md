---
title: Tunnels Only Listed Destinations = true
status: active
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_proxy
property_key: tunnels_only_listed_destinations
canonical_key: mcp-kibi-model-requirement:plugins.egress_proxy:tunnels_only_listed_destinations:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1649B811359E478A
claim_text: The egress proxy must open only HTTP CONNECT tunnels, directly and to a host and port on the identified plugin's egress allow-list, and refuse plain HTTP proxying and requests to chain through another proxy.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:20.798Z'
id: FACT-PROP-1B15FB9C56FD3469
type: fact
---
