---
title: Verifies And Shows Permissions Before Apply = true
status: active
text_ref: design-decisions:D-46#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-46-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.install_cli
property_key: verifies_and_shows_permissions_before_apply
canonical_key: docs-design-md:plugins.install_cli:verifies_and_shows_permissions_before_apply:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E9882267F0645C11
claim_text: The openfief CLI must verify a plugin's signature and show its requested permissions before applying the Plugin and PluginGrant with the owner's kubeconfig.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:00.967Z'
id: FACT-PROP-D60934BF06444D9C
type: fact
---
