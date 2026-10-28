---
title: Enforces Monthly Budget = true
status: active
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.automation
property_key: enforces_monthly_budget
canonical_key: mcp-kibi-model-requirement:demo.automation:enforces_monthly_budget:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B4E640AA32052748
claim_text: Before creating a demo resource, the demo automation must sum the monthly gross price of every resource labelled app=openfief in the Hetzner project plus what it would create, show the arithmetic, and refuse when the total exceeds DEMO_BUDGET_EUR, which defaults to 10 EUR.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:00:04.636Z'
id: FACT-PROP-5BD18608E506DB40
type: fact
---
