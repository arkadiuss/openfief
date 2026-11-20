---
title: Plugins Config Field Host
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-75-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
text_ref: design-decisions:D-75#1
fact_kind: subject
subject_key: plugins.config_field_host
canonical_key: plugins.config_field_host
claim_key: CLAIM-379112A106C0E870
claim_text: Instead of a fixed host, a manifest may request the host of a named config field; the CLI must read that host from the instance's config and show it, and the grant must approve exactly that host.
origin:
  kind: agent
  recorded_at: '2026-10-29T07:39:05.613Z'
id: FACT-SUBJECT-9E1AFC3BC2FA1036
type: fact
---

Definition not stated: the subject plugins.config_field_host is named by requirements REQ-plugins-config-field-host-grant, REQ-plugins-config-field-host-https, REQ-plugins-config-field-host-via-proxy, which link to the subject fact FACT-SUBJECT-9E1AFC3BC2FA1036 through constrains. The people who own that source can say what it is in the product and what it is not.
