---
title: Demo Loader
status: active
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: demo.loader
canonical_key: demo.loader
claim_key: CLAIM-276A49FAB9D7BECB
claim_text: The demo loader must refuse, before writing anything, a database that holds ledger or valuation data and has no demo marker, with or without --reset-demo; it may load an empty database, which it first marks as a demo database, and may reset only a database that carries the marker.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:58.988Z'
id: FACT-SUBJECT-8445428BF4E0E148
type: fact
---

Definition not stated: the subject demo.loader is named by requirements REQ-demo-fx-rates, REQ-demo-fx-rates-cover-lot-tax-dates, REQ-demo-fx-rates-one-source, REQ-demo-guard-refuses-real-data, REQ-demo-idempotent-per-day, REQ-demo-real-pipelines, REQ-demo-reseed-on-new-loader, REQ-demo-reseeded-before-first-release, REQ-demo-top-up-loaded-inputs, which link to the subject fact FACT-SUBJECT-8445428BF4E0E148 through constrains. The people who own that source can say what it is in the product and what it is not.
