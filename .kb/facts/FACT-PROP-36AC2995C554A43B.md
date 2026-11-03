---
title: Never Platform Secret = true
status: active
text_ref: docs/DESIGN.md#D-182
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-182
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.secrets
property_key: never_platform_secret
canonical_key: mcp-kibi-model-requirement:plugins.secrets:never_platform_secret:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-344A4776C65FF854
claim_text: No plugin instance may be named openfief or a name starting with openfief-, no grant may name a Secret starting with openfief-, and uninstalling an instance must delete only the Secrets whose openfief.io/instance label names that instance, read again just before each delete.
origin:
  kind: agent
  recorded_at: '2026-11-02T22:19:51.745Z'
id: FACT-PROP-36AC2995C554A43B
type: fact
---
