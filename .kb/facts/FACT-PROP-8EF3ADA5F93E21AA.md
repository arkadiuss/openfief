---
title: Inbound Only From Allowed Cidrs = true
status: active
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.server
property_key: inbound_only_from_allowed_cidrs
canonical_key: mcp-kibi-model-requirement:demo.server:inbound_only_from_allowed_cidrs:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-895A6F8BE6B71F16
claim_text: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, and HTTP and HTTPS only from DEMO_HTTP_CIDRS, which default to everyone, through a Hetzner firewall with no other inbound rule; the automation must refuse to create a server when no key has that label.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:34:44.836Z'
id: FACT-PROP-8EF3ADA5F93E21AA
type: fact
---
