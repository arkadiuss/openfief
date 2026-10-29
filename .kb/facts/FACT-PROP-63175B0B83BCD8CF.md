---
title: Issuer Export Uses Series Isin Derivation = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-99-2
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-99#2
fact_kind: property_value
subject_key: ingest.treasury_bonds
property_key: issuer_export_uses_series_isin_derivation
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:ingest.treasury_bonds:issuer_export_uses_series_isin_derivation:eq:true
claim_key: CLAIM-B03098C123D58F53
claim_text: An import of the issuer's own export must derive each bond's ISIN from its series code the same way, so one series is never held twice.
origin:
  kind: agent
  recorded_at: '2026-10-29T09:49:57.371Z'
id: FACT-PROP-63175B0B83BCD8CF
type: fact
---
