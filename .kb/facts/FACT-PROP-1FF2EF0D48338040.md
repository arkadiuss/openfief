---
title: Is Scoped = true
status: active
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: identity.session
property_key: is_scoped
canonical_key: mcp-kibi-model-requirement:identity.session:is_scoped:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-06A52880D769876F
claim_text: 'Every session must carry a scope, read or owner, enforced by the core: a read session may only use the read API, WhoAmI and Logout, while an owner session may also manage passkeys, script tokens and enrollment links.'
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:12.459Z'
id: FACT-PROP-1FF2EF0D48338040
type: fact
---
