---
title: Plugins Connector Institutions
status: active
text_ref: design-decisions:D-185#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-185-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.connector_institutions
canonical_key: plugins.connector_institutions
claim_key: CLAIM-2C2ABA06B41B195F
claim_text: 'The institutions a connector books for must be a permission its grant approves, as its hostnames are: the signed manifest declares them and the core accepts only what both name.'
origin:
  kind: agent
  recorded_at: '2026-11-03T07:22:34.229Z'
id: FACT-SUBJECT-6C3AD15C63FE7B0C
type: fact
---

Definition not stated: the subject plugins.connector_institutions is named by requirements REQ-plugins-install-warns-shared-institution, REQ-plugins-institutions-grant-approved, REQ-plugins-legacy-grants-approve-declared-institutions, REQ-plugins-upgrade-warns-shared-institution, which link to the subject fact FACT-SUBJECT-6C3AD15C63FE7B0C through constrains. The people who own that source can say what it is in the product and what it is not.
