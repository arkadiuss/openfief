---
title: Plugins Plugin Install
status: active
text_ref: design-decisions:D-48#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-48-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.plugin_install
canonical_key: plugins.plugin_install
claim_key: CLAIM-935556E59920D0E7
claim_text: Each plugin install must be a separately named instance with its own config, secrets, grant and egress allow-list.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:18.948Z'
id: FACT-SUBJECT-F3AE047F0BE90664
type: fact
---

Definition not stated: the subject plugins.plugin_install is named by requirements REQ-plugins-install-shows-market-data-list, REQ-plugins-instance-per-install, which link to the subject fact FACT-SUBJECT-F3AE047F0BE90664 through constrains. The people who own that source can say what it is in the product and what it is not.
