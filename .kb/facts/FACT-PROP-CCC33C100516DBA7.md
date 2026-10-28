---
title: Graceful Shutdown = true
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: core.service
property_key: graceful_shutdown
canonical_key: mcp-kibi-model-requirement:core.service:graceful_shutdown:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C8298DEF301B39E0
claim_text: On SIGTERM the core must stop reporting ready, let in-flight requests and running revaluations finish within its shutdown timeout, then cancel what is left so River retries it.
origin:
  kind: agent
  recorded_at: '2026-10-28T14:33:11.751Z'
id: FACT-PROP-CCC33C100516DBA7
type: fact
---
