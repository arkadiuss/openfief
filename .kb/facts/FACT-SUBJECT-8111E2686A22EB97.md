---
title: Audit Change Transaction Isolation
status: active
fact_kind: subject
subject_key: audit.change_transaction_isolation
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-8111E2686A22EB97
type: fact
---
The rule that a change kept together with its audit entry makes no network call inside its transaction: connector, model provider, SpiceDB and cluster calls happen before it opens.