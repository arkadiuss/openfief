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
claim_key: CLAIM-5546B95D1BAD5EC6
claim_text: The docs build must ask kibi for the requirement coverage report in one query, print how long it took, and warn when it took 60 seconds or more, half of kibi's 120-second query limit, as a workflow warning in CI.
origin:
  kind: agent
  recorded_at: '2026-11-07T19:28:01.095Z'
id: FACT-PROP-5EA14B0CBA412425
type: fact
---
