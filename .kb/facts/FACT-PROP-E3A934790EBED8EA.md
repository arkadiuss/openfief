---
title: Bond Values And Liquidation Match Reference = true
status: active
text_ref: design-decisions:D-200#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-5
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: m0.exit
property_key: bond_values_and_liquidation_match_reference
canonical_key: mcp-kibi-model-requirement:m0.exit:bond_values_and_liquidation_match_reference:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B5FBEA947A426D12
claim_text: For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; the liquidation value must match a hand calculation for at least one series of each type held; and totals must carry all three values.
origin:
  kind: agent
  recorded_at: '2026-11-08T01:31:57.063Z'
id: FACT-PROP-E3A934790EBED8EA
type: fact
---
