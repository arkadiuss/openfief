---
title: Valuation River Migrations
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: valuation.river_migrations
canonical_key: valuation.river_migrations
claim_key: CLAIM-BA71BFF910DC244F
claim_text: River's migrations must run under an advisory lock shared by every process that migrates the database, so that concurrent migrations, such as the core's and the demo loader's, all succeed and apply each version once.
origin:
  kind: agent
  recorded_at: '2026-11-03T17:15:02.265Z'
id: FACT-SUBJECT-4E48ABDC27143143
type: fact
---
