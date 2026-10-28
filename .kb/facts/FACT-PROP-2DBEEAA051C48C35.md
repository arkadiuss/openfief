---
title: Maps Flex Statements In Requested Mode = true
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.ibkr_flex
property_key: maps_flex_statements_in_requested_mode
canonical_key: mcp-kibi-model-requirement:plugins.ibkr_flex:maps_flex_statements_in_requested_mode:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-8BF9471BDBE8A3CE
claim_text: The IBKR Flex plugin must answer Sync by pulling its Flex Query from the Flex Web Service with the token and query ID from its mounted Secret, read on every sync and never logged or put in an error, and ImportFile by mapping an uploaded Flex statement, both with the ibkrflex mapper in the mode the call names, reaching only ndcdyn.interactivebrokers.com.
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:29.329Z'
id: FACT-PROP-2DBEEAA051C48C35
type: fact
---
