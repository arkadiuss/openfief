---
title: Plugin Isolation Holds = true
status: active
text_ref: design-decisions:D-200#10
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-10
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: m0.exit
property_key: plugin_isolation_holds
canonical_key: mcp-kibi-model-requirement:m0.exit:plugin_isolation_holds:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-872F0D49928AA93E
claim_text: Each plugin must run in its own pod, the XTB parser must have no egress at all, a test plugin's request through the egress proxy for a host outside its grant must be refused, and attempts to bypass the proxy, a direct connection or a DNS lookup, must fail.
origin:
  kind: agent
  recorded_at: '2026-11-07T03:32:07.222Z'
id: FACT-PROP-D9EB1627F96B6496
type: fact
---
