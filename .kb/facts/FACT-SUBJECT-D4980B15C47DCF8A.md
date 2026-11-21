---
title: Operator Plugin Status
status: active
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: operator.plugin_status
canonical_key: operator.plugin_status
claim_key: CLAIM-5D427B563C4C0A91
claim_text: When a Plugin's spec requests permissions its PluginGrant does not approve, the operator must keep running the previously deployed version and report exactly the unapproved permissions in status.pendingApproval and an ApprovalRequired condition.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:10.696Z'
id: FACT-SUBJECT-D4980B15C47DCF8A
type: fact
---

Definition not stated: the subject operator.plugin_status is named by requirements REQ-operator-plugin-trust-status, REQ-operator-signature-signer-status, REQ-operator-upgrade-status, which link to the subject fact FACT-SUBJECT-D4980B15C47DCF8A through constrains. The people who own that source can say what it is in the product and what it is not.
