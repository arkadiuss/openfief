---
title: Identify Plugins By Their Projected Serviceaccount Token required
status: active
text_ref: design-decisions:D-32#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-32-3
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_proxy
property_key: identify_plugins_by_their_projected_serviceaccount_token
canonical_key: knowledge-source-design-decisions:plugins.egress_proxy:identify_plugins_by_their_projected_serviceaccount_token:polarity:require
operator: eq
value_type: bool
value_bool: true
polarity: require
claim_key: CLAIM-1065939859669F4D
claim_text: The egress proxy must identify plugins by their projected ServiceAccount token.
origin:
  kind: agent
  recorded_at: '2026-10-27T20:12:52.878Z'
id: FACT-PROP-B7976F5DAC647953
type: fact
---
