---
title: Reconciles Activities And Balances = true
status: active
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.ghostfolio_plugin
property_key: reconciles_activities_and_balances
canonical_key: mcp-kibi-model-requirement:views.ghostfolio_plugin:reconciles_activities_and_balances:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C3D542B224EBA9F9
claim_text: Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates and, when it imports onto a MANUAL asset, those on MANUAL assets it does not desire; must empty with one account-filtered bulk delete an owned account that should hold nothing; must set every desired cash balance an owned account lacks or holds with another value and delete its balances on other dates; and must leave nothing to change for a second sync.
origin:
  kind: agent
  recorded_at: '2026-10-30T01:21:18.214Z'
id: FACT-PROP-5CD9AB023118EEAF
type: fact
---
