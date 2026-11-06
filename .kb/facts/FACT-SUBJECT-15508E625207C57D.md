---
title: E2e Egress Proxy Log
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: e2e.egress_proxy_log
canonical_key: e2e.egress_proxy_log
claim_key: CLAIM-C5C852B6F860B961
claim_text: On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance and nothing of its path, query, headers, body or token, and every request it allowed must be to a host on that instance's egress allow-list.
origin:
  kind: agent
  recorded_at: '2026-11-06T20:50:01.325Z'
id: FACT-SUBJECT-15508E625207C57D
type: fact
---
