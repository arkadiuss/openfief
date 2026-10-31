---
title: Cash Balances Per Currency = true
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
property_key: cash_balances_per_currency
canonical_key: mcp-kibi-model-requirement:views.ghostfolio_plugin:cash_balances_per_currency:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-ADE19139887CAFDE
claim_text: 'The Ghostfolio view must give each owned account one cash balance at the end of every date on which the account''s cash in that owned account''s currency changes, counting the transactions it leaves out: the account''s own in the Ghostfolio user''s base currency, and each cash account in its currency, which holds only that balance; and must leave out an owned account''s balances when one is too large for a double.'
origin:
  kind: agent
  recorded_at: '2026-10-31T19:00:27.843Z'
id: FACT-PROP-D308055D705EE97D
type: fact
---
