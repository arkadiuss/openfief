---
title: Ssh Only From Allowed Cidrs = true
status: active
text_ref: docs/DESIGN.md#D-28
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-28
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.server
property_key: ssh_only_from_allowed_cidrs
canonical_key: mcp-kibi-model-requirement:demo.server:ssh_only_from_allowed_cidrs:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F44068081F69F05D
claim_text: A demo server must accept SSH only with the project's keys labelled openfief-ssh=true and only from DEMO_SSH_CIDRS, through a Hetzner firewall with no other inbound rule, and the automation must refuse to create a server when no key has that label.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:00:38.147Z'
id: FACT-PROP-B0B3E7F2D9F867CD
type: fact
---
