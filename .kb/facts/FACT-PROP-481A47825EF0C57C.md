---
title: Deploy Until They Are Approved forbidden
status: active
text_ref: design-decisions:D-16#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-16-3
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.plugin_upgrade
property_key: deploy_until_they_are_approved
canonical_key: knowledge-source-design-decisions:plugins.plugin_upgrade:deploy_until_they_are_approved:polarity:forbid
operator: eq
value_type: bool
value_bool: true
polarity: forbid
claim_key: CLAIM-4E50A2E1D2C2CAD3
claim_text: A plugin upgrade that requests more permissions must not deploy until they are approved.
origin:
  kind: agent
  recorded_at: '2026-10-27T20:12:30.524Z'
id: FACT-PROP-481A47825EF0C57C
type: fact
---
