---
title: One Sleeve Per Account = true
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
subject_key: portfolio.account_sleeves
property_key: one_sleeve_per_account
canonical_key: mcp-kibi-model-requirement:portfolio.account_sleeves:one_sleeve_per_account:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3D9E32EA7A9495C3
claim_text: The portfolio context must keep the one sleeve each account is in, keyed by the ledger's account ID, refuse to put an account the ledger doesn't hold or in a sleeve that isn't stored, and take a deleted account out of its sleeve.
origin:
  kind: agent
  recorded_at: '2026-11-03T11:50:33.860Z'
id: FACT-PROP-3B187D466D26AFFB
type: fact
---
