---
title: Reconciles Owned Accounts Idempotently = true
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
property_key: reconciles_owned_accounts_idempotently
canonical_key: mcp-kibi-model-requirement:views.ghostfolio_plugin:reconciles_owned_accounts_idempotently:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B85BFCF2D59CA400
claim_text: Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates, empty with one account-filtered bulk delete an owned account that should hold nothing, and leave nothing to change for a second sync.
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:56.689Z'
id: FACT-PROP-04F58C874E3F4F64
type: fact
---
