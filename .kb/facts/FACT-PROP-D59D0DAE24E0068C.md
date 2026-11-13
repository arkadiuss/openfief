---
title: Exempt Host Count Max <= 1
status: active
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - provenance:mcp-kibi-model-requirement
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: egress_proxy.refused_range_exemption
property_key: exempt_host_count_max
operator: lte
value_type: int
value_int: 1
canonical_key: mcp-kibi-model-requirement:egress_proxy.refused_range_exemption:exempt_host_count_max:lte:1
claim_key: CLAIM-0FB258547E947F44
claim_text: The end-to-end suite's cluster configuration may exempt one host from the egress proxy's refused address ranges, and its NetworkPolicy lets the proxy reach that host.
origin:
  kind: agent
  recorded_at: '2026-11-12T23:18:38.614Z'
id: FACT-PROP-D59D0DAE24E0068C
type: fact
---
