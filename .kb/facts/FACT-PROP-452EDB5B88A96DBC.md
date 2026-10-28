---
title: Is Canonical Hostname = true
status: active
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_host
property_key: is_canonical_hostname
canonical_key: mcp-kibi-model-requirement:plugins.egress_host:is_canonical_hostname:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C5B5420AB26564BA
claim_text: An egress host in a manifest or grant must be a canonical DNS hostname with a port, and an IP literal, a wildcard or a single-label name must be rejected.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:21.011Z'
id: FACT-PROP-452EDB5B88A96DBC
type: fact
---
