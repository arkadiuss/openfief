---
title: Twr Cash Flow Exclusions within confirmed_transfer
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-105-4
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-105#4
fact_kind: property_value
subject_key: views.performance
property_key: twr_cash_flow_exclusions
operator: subset_of
value_type: set
value_set:
  - confirmed_transfer
canonical_key: docs-design-md:views.performance:twr_cash_flow_exclusions:subset_of:confirmed_transfer
claim_key: CLAIM-7EA84DB6754B430C
claim_text: Only confirmed transfers must be left out of the time-weighted return's cash flows.
origin:
  kind: agent
  recorded_at: '2026-10-29T10:11:15.269Z'
id: FACT-PROP-CC6363403AD6890F
type: fact
---
