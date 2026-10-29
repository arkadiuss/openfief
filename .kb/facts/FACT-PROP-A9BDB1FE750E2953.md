---
title: Suggestion Rule = deposit into another own account, same amount and currency, at most 5 calendar days apart, nearest unmatched deposit first
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-105-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-105#1
fact_kind: property_value
subject_key: ledger.transfer
property_key: suggestion_rule
operator: eq
value_type: string
value_string: deposit into another own account, same amount and currency, at most 5 calendar days apart, nearest unmatched deposit first
canonical_key: docs-design-md:ledger.transfer:suggestion_rule:eq:deposit_into_another_own_account_same_amount_and_currency_at_most_5_calendar_days_apart_nearest_unmatched_deposit_first
claim_key: CLAIM-B3882A0F98494B51
claim_text: A withdrawal must be suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first.
origin:
  kind: agent
  recorded_at: '2026-10-29T10:09:17.067Z'
id: FACT-PROP-A9BDB1FE750E2953
type: fact
---
