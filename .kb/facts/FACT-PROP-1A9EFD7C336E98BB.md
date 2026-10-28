---
title: Reports In Configured Currency Pln Only = true
status: active
text_ref: docs/DESIGN.md#D-62
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-62
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.reporting_currency
property_key: reports_in_configured_currency_pln_only
canonical_key: mcp-kibi-model-requirement:views.reporting_currency:reports_in_configured_currency_pln_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-65B29AB56E609541
claim_text: The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than PLN with a typed unsupported reporting currency error until cross rates are decided.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:06:23.253Z'
id: FACT-PROP-1A9EFD7C336E98BB
type: fact
---
