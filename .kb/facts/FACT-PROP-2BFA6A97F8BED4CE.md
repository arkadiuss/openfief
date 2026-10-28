---
title: Refuses Internal Ranges = true
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
subject_key: plugins.egress_address
property_key: refuses_internal_ranges
canonical_key: mcp-kibi-model-requirement:plugins.egress_address:refuses_internal_ranges:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-97B067D6FA89738F
claim_text: The egress proxy must refuse to dial any address in a private, loopback, link-local, carrier-grade NAT, unique-local, multicast, unspecified, documentation or reserved range, including IPv4 addresses embedded in IPv6 addresses.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:32:02.472Z'
id: FACT-PROP-2BFA6A97F8BED4CE
type: fact
---
