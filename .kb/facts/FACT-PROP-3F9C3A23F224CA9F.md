---
title: Embeds Smokescreen Via Buildproxy = true
status: active
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_proxy
property_key: embeds_smokescreen_via_buildproxy
canonical_key: mcp-kibi-model-requirement:plugins.egress_proxy:embeds_smokescreen_via_buildproxy:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-43B2694FD83DD205
claim_text: The egress proxy must serve Smokescreen's BuildProxy handler from OpenFief's own HTTP server, never let Smokescreen exit the process or install signal handlers, and on shutdown stop accepting, let open tunnels drain for a bounded time and close the rest.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:46:53.252Z'
id: FACT-PROP-3F9C3A23F224CA9F
type: fact
---
