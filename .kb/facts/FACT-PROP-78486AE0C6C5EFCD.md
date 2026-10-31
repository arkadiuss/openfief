---
title: Owns Account And Cash Accounts = true
status: active
text_ref: docs/DESIGN.md#D-133
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-133
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.ghostfolio_plugin
property_key: owns_account_and_cash_accounts
canonical_key: mcp-kibi-model-requirement:views.ghostfolio_plugin:owns_account_and_cash_accounts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3B70B911B00513F2
claim_text: The Ghostfolio view must own one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, and one per other currency of that account's cash, identified by a UUID derived from the instance's name, the account's ID and the currency, create each with its UUID when it is missing, and never change or delete an account or activity outside its owned accounts.
origin:
  kind: agent
  recorded_at: '2026-10-31T19:00:18.595Z'
id: FACT-PROP-78486AE0C6C5EFCD
type: fact
---
