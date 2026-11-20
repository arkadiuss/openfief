---
title: Views Ghostfolio
status: active
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: views.ghostfolio
canonical_key: views.ghostfolio
claim_key: CLAIM-8B086CCEFA5B7584
claim_text: After a bond is re-keyed to its registry ISIN, the Ghostfolio view's next sync must import and delete nothing, since a bond's asset is derived from its series code and its activities name transaction keys.
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:30.098Z'
id: FACT-SUBJECT-30D618C3C5B4C80D
type: fact
---

Definition not stated: the subject views.ghostfolio is named by requirement REQ-views-ghostfolio-rekey-resync, which links to the subject fact FACT-SUBJECT-30D618C3C5B4C80D through constrains. The people who own that source can say what it is in the product and what it is not.
