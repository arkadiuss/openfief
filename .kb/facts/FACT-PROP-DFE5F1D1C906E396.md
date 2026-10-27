---
title: Refuse Private And Internal Address Ranges required
status: active
text_ref: design-decisions:D-32#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-32-5
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_proxy
property_key: refuse_private_and_internal_address_ranges
canonical_key: knowledge-source-design-decisions:plugins.egress_proxy:refuse_private_and_internal_address_ranges:polarity:require
operator: eq
value_type: bool
value_bool: true
polarity: require
claim_key: CLAIM-72206A2FCA516E9B
claim_text: The egress proxy must refuse private and internal address ranges.
origin:
  kind: agent
  recorded_at: '2026-10-27T20:12:54.250Z'
id: FACT-PROP-DFE5F1D1C906E396
type: fact
---
