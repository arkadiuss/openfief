---
title: Plugins Manifest Secret
status: active
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.manifest_secret
canonical_key: plugins.manifest_secret
claim_key: CLAIM-E2AAB6F33C41BABF
claim_text: A plugin manifest may mark a Secret role optional; the operator must mount that Secret as an optional volume so the plugin pod starts while the owner has not created it, every other Secret must stay required, the flag must survive the Plugin resource the CLI writes and reads, and the install summary must say the plugin runs without it.
origin:
  kind: agent
  recorded_at: '2026-11-06T21:55:03.419Z'
id: FACT-SUBJECT-B6BDD975F76E8B0B
type: fact
---
