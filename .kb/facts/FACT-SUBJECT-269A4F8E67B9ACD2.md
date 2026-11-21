---
title: Core Browser Entry
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-81-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
text_ref: design-decisions:D-81#1
fact_kind: subject
subject_key: core.browser_entry
canonical_key: core.browser_entry
claim_key: CLAIM-A50C05C302FDC229
claim_text: The core must be the only entry point for browsers, so that kubectl port-forward or the owner's own Ingress needs to reach the core alone.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:11:10.833Z'
id: FACT-SUBJECT-269A4F8E67B9ACD2
type: fact
---

Definition not stated: the subject core.browser_entry is named by requirements REQ-core-browser-entry-point, REQ-core-only-browser-entry-point, REQ-core-reverse-proxies-view-plugins, which link to the subject fact FACT-SUBJECT-269A4F8E67B9ACD2 through constrains. The people who own that source can say what it is in the product and what it is not.
