---
title: Plugins Secret Roles
status: active
text_ref: docs/DESIGN.md#D-89
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-89
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: plugins.secret_roles
canonical_key: plugins.secret_roles
claim_key: CLAIM-CB4583049B0C1862
claim_text: A plugin instance must request, and its workload must mount, the Secret <instance>-<role> for each Secret role its manifest names, and a role must be a DNS label.
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:26.223Z'
id: FACT-SUBJECT-010DD798EAB0E3D4
type: fact
---

Definition not stated: the subject plugins.secret_roles is named by requirements REQ-plugins-no-double-hyphen-in-instance-or-role, REQ-plugins-requested-secret-names, REQ-plugins-requested-secret-names-double-hyphen, which link to the subject fact FACT-SUBJECT-010DD798EAB0E3D4 through constrains. The people who own that source can say what it is in the product and what it is not.
