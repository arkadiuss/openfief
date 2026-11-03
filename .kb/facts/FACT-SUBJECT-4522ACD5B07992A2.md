---
title: Deploy Network Policies
status: active
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: deploy.network_policies
canonical_key: deploy.network_policies
claim_key: CLAIM-600CFF1FB23ABA9F
claim_text: 'The database, the egress proxy and the operator must each have a NetworkPolicy admitting only their callers: the database the core, its own instances and CloudNativePG''s operator; the egress proxy plugin pods, with its own egress leaving the private, shared, link-local and loopback ranges out; the operator nothing.'
origin:
  kind: agent
  recorded_at: '2026-11-03T00:10:42.926Z'
id: FACT-SUBJECT-4522ACD5B07992A2
type: fact
---
