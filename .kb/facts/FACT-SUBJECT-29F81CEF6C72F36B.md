---
title: Operator Ca Bundle Status
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
subject_key: operator.ca_bundle_status
canonical_key: operator.ca_bundle_status
claim_key: CLAIM-7C1D90B7861616A1
claim_text: When a running plugin instance expects an extra CA bundle and its ca-bundle Secret does not exist, the operator must set the Plugin condition CABundleMissing to True with what to do, False when the Secret exists, Unknown when it may not read it, and no condition when none is expected, never holding the Secret data and never keeping the plugin from running.
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:08.224Z'
id: FACT-SUBJECT-29F81CEF6C72F36B
type: fact
---
