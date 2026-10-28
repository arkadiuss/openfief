---
title: Kinds = Plugin, PluginGrant
status: active
text_ref: design-decisions:D-55#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-55-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.custom_resources
property_key: kinds
canonical_key: docs-design-md:plugins.custom_resources:kinds:eq:plugin_plugingrant
operator: eq
value_type: set
value_set:
  - Plugin
  - PluginGrant
claim_key: CLAIM-A469333E6D8269D0
claim_text: OpenFief's only custom resources must be Plugin and PluginGrant; fiefs, intents, approvals and the ledger must live in PostgreSQL behind the core.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:19.090Z'
id: FACT-PROP-CC897825B9C20A83
type: fact
---
