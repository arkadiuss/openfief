---
title: Plugins Image Signature
status: active
text_ref: design-decisions:D-45#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-45-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.image_signature
canonical_key: plugins.image_signature
claim_key: CLAIM-3CA881B937966C4D
claim_text: The operator must accept only images signed by OpenFief's CI, a publisher pinned in the catalog, or a signer the owner added.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:38:51.648Z'
id: FACT-SUBJECT-DD45415226C332C2
type: fact
---

Definition not stated: the subject plugins.image_signature is named by requirement REQ-plugins-trusted-signers-only, which links to the subject fact FACT-SUBJECT-DD45415226C332C2 through constrains. The people who own that source can say what it is in the product and what it is not.
