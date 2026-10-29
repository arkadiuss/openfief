---
title: Published For Plugins = true
status: active
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.signing_keys
property_key: published_for_plugins
canonical_key: mcp-kibi-model-requirement:plugins.signing_keys:published_for_plugins:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-BDD7C3912DC1D6A1
claim_text: The operator must publish the cluster's ServiceAccount token issuer and only public RSA or EC signing keys with key IDs, read from the API server's issuer discovery endpoints, in the ConfigMap openfief-sa-jwks at start, periodically and whenever the ConfigMap changes, keeping the published keys when they can't be read.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:40.685Z'
id: FACT-PROP-22A0B2B7EE62CC89
type: fact
---
