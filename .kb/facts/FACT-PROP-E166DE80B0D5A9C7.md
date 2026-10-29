---
title: Serves Only Front Door Read Api And Health Endpoints = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-115-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-115#1
fact_kind: property_value
subject_key: core.browser_port
property_key: serves_only_front_door_read_api_and_health_endpoints
operator: eq
value_type: bool
value_bool: true
canonical_key: kb-requirements-req-core-browser-port-scope-md:core.browser_port:serves_only_front_door_read_api_and_health_endpoints:eq:true
claim_key: CLAIM-56B90E86E4B73A4D
claim_text: The core's browser port, which the Service's http port targets, must serve only the front door, the read API and the health endpoints.
origin:
  kind: agent
  recorded_at: '2026-10-29T16:12:20.059Z'
id: FACT-PROP-E166DE80B0D5A9C7
type: fact
---
