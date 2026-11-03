---
title: Keeps Payout Setting Per Account = true
status: active
text_ref: docs/dev/code-organization.md#bounded-contexts
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-bounded-contexts
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.payout_settings_store
property_key: keeps_payout_setting_per_account
canonical_key: mcp-kibi-model-requirement:ingest.payout_settings_store:keeps_payout_setting_per_account:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-AB1B1556AB5E5046
claim_text: The ingest context must keep each account's payout setting, with the day it took effect and the settings in force before it, in its own PostgreSQL schema keyed by the ledger's account ID, read the default, as cash, for an account with none, and forget it with a deleted account.
origin:
  kind: agent
  recorded_at: '2026-11-03T11:51:36.579Z'
id: FACT-PROP-ED0287A60C74CDDC
type: fact
---
