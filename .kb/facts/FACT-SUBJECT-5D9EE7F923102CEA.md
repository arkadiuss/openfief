---
title: Plugins Network Policy
status: active
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: plugins.network_policy
canonical_key: plugins.network_policy
claim_key: CLAIM-A8DCAEEA129FE717
claim_text: A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, and never DNS.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:23.128Z'
id: FACT-SUBJECT-5D9EE7F923102CEA
type: fact
---

Definition not stated: the subject plugins.network_policy is named by requirement REQ-plugins-network-policy-core-and-proxy, which links to the subject fact FACT-SUBJECT-5D9EE7F923102CEA through constrains. The people who own that source can say what it is in the product and what it is not.
