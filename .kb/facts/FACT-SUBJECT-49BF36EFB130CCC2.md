---
title: Ingest Payout Settings Store
status: active
text_ref: docs/dev/code-organization.md#bounded-contexts
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-bounded-contexts
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.payout_settings_store
canonical_key: ingest.payout_settings_store
claim_key: CLAIM-AB1B1556AB5E5046
claim_text: The ingest context must keep each account's payout setting, with the day it took effect and the settings in force before it, in its own PostgreSQL schema keyed by the ledger's account ID, read the default, as cash, for an account with none, and forget it with a deleted account.
origin:
  kind: agent
  recorded_at: '2026-11-03T11:51:32.097Z'
id: FACT-SUBJECT-49BF36EFB130CCC2
type: fact
---
