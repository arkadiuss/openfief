---
title: Warns From Half Of Query Limit = true
status: active
text_ref: docs/dev/kibi.md#setup-notes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-kibi-md-setup-notes
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.kibi_gate
property_key: warns_from_half_of_query_limit
canonical_key: mcp-kibi-model-requirement:tooling.kibi_gate:warns_from_half_of_query_limit:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-453DE1B77AABDEE8
claim_text: The docs build must ask kibi for the requirement coverage report at most 300 requirements at a time, after a query that attaches the store, print how long each page took, and warn when a page took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI.
origin:
  kind: agent
  recorded_at: '2026-11-07T19:28:01.095Z'
id: FACT-PROP-5EA14B0CBA412425
type: fact
---
