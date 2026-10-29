---
title: Core Is Only Entry Point For Browsers = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-81-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-81#1
fact_kind: property_value
subject_key: core.browser_entry
property_key: core_is_only_entry_point_for_browsers
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:core.browser_entry:core_is_only_entry_point_for_browsers:eq:true
claim_key: CLAIM-A50C05C302FDC229
claim_text: The core must be the only entry point for browsers, so that kubectl port-forward or the owner's own Ingress needs to reach the core alone.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:11:14.680Z'
id: FACT-PROP-A5C42159D9E3E94D
type: fact
---
