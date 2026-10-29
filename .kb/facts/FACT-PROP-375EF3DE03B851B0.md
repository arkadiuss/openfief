---
title: Core Checks View Plugin Tokens With Token Review = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-82-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-82#1
fact_kind: property_value
subject_key: identity.plugins
property_key: core_checks_view_plugin_tokens_with_token_review
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:identity.plugins:core_checks_view_plugin_tokens_with_token_review:eq:true
claim_key: CLAIM-32EA50C847B50C55
claim_text: The core must check view plugins' tokens with Kubernetes' TokenReview API, which also confirms that the plugin's pod still exists.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:13:25.269Z'
id: FACT-PROP-375EF3DE03B851B0
type: fact
---
