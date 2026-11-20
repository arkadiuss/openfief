---
title: Valuation Input Validation
status: active
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: valuation.input_validation
canonical_key: valuation.input_validation
claim_key: CLAIM-73E2AB4610F596C5
claim_text: Recording a price that is not positive or names no instrument, an FX rate that is not positive, names no base or is quoted in its own currency, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing.
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:13.955Z'
id: FACT-SUBJECT-F99282023E49B14B
type: fact
---

Definition not stated: the subject valuation.input_validation is named by requirement REQ-valuation-input-validation-named-base, which links to the subject fact FACT-SUBJECT-F99282023E49B14B through constrains. The people who own that source can say what it is in the product and what it is not.
