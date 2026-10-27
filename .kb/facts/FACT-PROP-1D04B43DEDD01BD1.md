---
title: Create A New Transaction Version That Supersedes The Old One And The Old Version Must Never Be Deleted required
status: active
text_ref: design-decisions:D-18#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-18-4
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.changed_source_row
property_key: create_a_new_transaction_version_that_supersedes_the_old_one_and_the_old_version_must_never_be_deleted
canonical_key: knowledge-source-design-decisions:ledger.changed_source_row:create_a_new_transaction_version_that_supersedes_the_old_one_and_the_old_version_must_never_be_deleted:polarity:require
operator: eq
value_type: bool
value_bool: true
polarity: require
claim_key: CLAIM-084574CAAB590EC4
claim_text: A changed source row must create a new transaction version that supersedes the old one, and the old version must never be deleted.
origin:
  kind: agent
  recorded_at: '2026-10-27T20:12:32.570Z'
id: FACT-PROP-1D04B43DEDD01BD1
type: fact
---
