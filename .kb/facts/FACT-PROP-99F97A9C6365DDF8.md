---
title: Owns Derived Accounts Only = true
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
property_key: owns_derived_accounts_only
canonical_key: mcp-kibi-model-requirement:views.ghostfolio_plugin:owns_derived_accounts_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-10C075DD3C338EF8
claim_text: The Ghostfolio view must own exactly one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, create it with that UUID when it is missing, and never change or delete an account or activity outside its owned accounts.
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:41.095Z'
id: FACT-PROP-99F97A9C6365DDF8
type: fact
---
