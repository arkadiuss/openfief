---
title: Kinds Per Holding Account And Total = gross, liquidation, net
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-77-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-77#1
fact_kind: property_value
subject_key: valuation.values
property_key: kinds_per_holding_account_and_total
operator: eq
value_type: set
value_set:
  - gross
  - liquidation
  - net
canonical_key: docs-design-md:valuation.values:kinds_per_holding_account_and_total:eq:gross_liquidation_net
claim_key: CLAIM-94008AF0009FB3B6
claim_text: Every holding, account and total must have three values, gross, net and liquidation, computed by the jurisdiction pack for each kind of product and tax wrapper.
origin:
  kind: agent
  recorded_at: '2026-10-29T07:56:58.712Z'
id: FACT-PROP-3F75FE5E2CCF9446
type: fact
---
