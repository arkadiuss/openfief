---
title: Valuation Owner Exemption Class
status: active
text_ref: docs/DESIGN.md#D-132
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-132
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: valuation.owner_exemption_class
canonical_key: valuation.owner_exemption_class
claim_key: CLAIM-7695703E7ACF4BD4
claim_text: The owner must be able to classify an instrument in one of the pack's exemption classes through the SettingsService, stored with the instrument and replacing an earlier choice, without revaluing.
origin:
  kind: agent
  recorded_at: '2026-10-31T18:34:36.034Z'
id: FACT-SUBJECT-078270D8F81FFF3F
type: fact
part_of:
  - valuation.exemption_class
---
The owner's choice of an instrument's exemption class through the SettingsService, stored with the instrument and replacing an earlier choice without revaluing. It is one part of valuation.exemption_class, which also covers how the jurisdiction pack classifies an instrument by itself.