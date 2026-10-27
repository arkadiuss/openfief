---
title: Allow Each Plugin Only The Hostnames In Its Grant required
status: active
text_ref: design-decisions:D-32#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-32-2
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_proxy
property_key: allow_each_plugin_only_the_hostnames_in_its_grant
canonical_key: knowledge-source-design-decisions:plugins.egress_proxy:allow_each_plugin_only_the_hostnames_in_its_grant:polarity:require
operator: eq
value_type: bool
value_bool: true
polarity: require
claim_key: CLAIM-E70A85926858A66F
claim_text: The egress proxy must allow each plugin only the hostnames in its grant.
origin:
  kind: agent
  recorded_at: '2026-10-27T20:12:52.380Z'
id: FACT-PROP-0C641A2167D22CAB
type: fact
---
