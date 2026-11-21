---
title: Core Network Policy
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-92-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
text_ref: design-decisions:D-92#1
fact_kind: subject
subject_key: core.network_policy
canonical_key: core.network_policy
claim_key: CLAIM-80DB53D5E461D669
claim_text: The core's NetworkPolicy must allow the Kubernetes API on port 6443 at any address.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:37:38.567Z'
id: FACT-SUBJECT-882842AE33F484A5
type: fact
---

Definition not stated: the subject core.network_policy is named by requirement REQ-core-network-policy-kubernetes-api, which links to the subject fact FACT-SUBJECT-882842AE33F484A5 through constrains. The people who own that source can say what it is in the product and what it is not.
