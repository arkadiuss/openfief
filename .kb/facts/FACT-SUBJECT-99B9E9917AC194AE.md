---
title: Operator Ca Bundle Refresh
status: active
text_ref: design-session:connector-endpoint-ca-refinements
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-refinements
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: operator.ca_bundle_refresh
canonical_key: operator.ca_bundle_refresh
claim_key: CLAIM-83640C80FC38E6B5
claim_text: While a running plugin instance expects an extra CA bundle, the operator must stamp a checksum of its ca-bundle Secret data on the pod template and read the Secret again at least every minute, so creating, changing or deleting the Secret rolls the pod, reading only that Secret by name through a Role the owner install creates.
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:12.653Z'
id: FACT-SUBJECT-99B9E9917AC194AE
type: fact
---
