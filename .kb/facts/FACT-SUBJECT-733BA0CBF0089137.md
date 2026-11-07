---
title: Plugins Config Host Change
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
subject_key: plugins.config_host_change
canonical_key: plugins.config_host_change
claim_key: CLAIM-0924ABCEA06E85BE
claim_text: When a plugin base URL moves to a host its grant does not approve, the running version must keep the old approved egress allow-list with the new host pending approval, and the upgrade diff must show the new host added and the old one removed.
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:17.137Z'
id: FACT-SUBJECT-733BA0CBF0089137
type: fact
---
