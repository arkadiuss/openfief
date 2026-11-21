---
title: Views Browser Proxy
status: active
text_ref: docs/DESIGN.md#D-81
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-81
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: views.browser_proxy
canonical_key: views.browser_proxy
claim_key: CLAIM-14EE12207C3EB2E9
claim_text: When the core proxies a browser request to a view plugin, it must present its own token for that instance and set the Host, X-Forwarded-Proto and X-Forwarded-Host itself, dropping every forwarding header the client sent.
origin:
  kind: agent
  recorded_at: '2026-10-29T12:02:04.732Z'
id: FACT-SUBJECT-058686C92437C05A
type: fact
---

Definition not stated: the subject views.browser_proxy is named by requirement REQ-views-browser-proxy-headers, which links to the subject fact FACT-SUBJECT-058686C92437C05A through constrains. The people who own that source can say what it is in the product and what it is not.
