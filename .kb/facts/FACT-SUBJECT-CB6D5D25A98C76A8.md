---
title: Plugins Signing Keys
status: active
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.signing_keys
canonical_key: plugins.signing_keys
claim_key: CLAIM-BDD7C3912DC1D6A1
claim_text: The operator must publish the cluster's ServiceAccount token issuer and only public RSA or EC signing keys with key IDs, read from the API server's issuer discovery endpoints, in the ConfigMap openfief-sa-jwks at start, periodically and whenever the ConfigMap changes, keeping the published keys when they can't be read.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:38.670Z'
id: FACT-SUBJECT-CB6D5D25A98C76A8
type: fact
---
