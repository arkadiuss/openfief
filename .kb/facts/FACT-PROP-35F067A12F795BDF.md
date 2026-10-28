---
title: Entered By The Owner Only = true
status: active
text_ref: docs/DESIGN.md#D-52
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-52
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: portfolio.model_portfolio
property_key: entered_by_the_owner_only
canonical_key: mcp-kibi-model-requirement:portfolio.model_portfolio:entered_by_the_owner_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-AEA4AB0E9E637834
claim_text: The core must let only an owner-scoped session on an allowed origin read or replace the model portfolio, through a service it serves to view plugins only, refusing a read session, another origin, an invalid session and a model that breaks a rule or names a malformed ISIN, and storing nothing on a refusal.
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:29.392Z'
id: FACT-PROP-35F067A12F795BDF
type: fact
---
