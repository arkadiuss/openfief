---
title: Plugins Egress Route
status: active
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.egress_route
canonical_key: plugins.egress_route
claim_key: CLAIM-B8823E04FCA2649D
claim_text: A plugin with granted egress hosts must get HTTPS_PROXY pointing at the egress proxy's cluster address and an egress allow-list of exactly those hosts, and a plugin without any must get neither a proxy setting nor a proxy route.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:37.094Z'
id: FACT-SUBJECT-927927C04D44139D
type: fact
---
