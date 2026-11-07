---
title: Cli Prints And Fails On Mismatch = true
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.balance_check
property_key: cli_prints_and_fails_on_mismatch
canonical_key: mcp-kibi-model-requirement:views.balance_check:cli_prints_and_fails_on_mismatch:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5EF6D06383264EBE
claim_text: openfief check balances must print the balance report only to the owner's terminal, as a table or as JSON with --json, and exit with a non-zero status when any balance assertion fails; the core must log only that the check ran, with its counts, never an account, an instrument or a value.
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:27.080Z'
id: FACT-PROP-4AEE3F6370174D26
type: fact
---
