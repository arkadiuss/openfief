---
title: Plugin egress must go through the egress proxy.
status: closed
text_ref: design-decisions:D-32#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-32-1
  - lane:strict
semantic_text: Plugin egress must go through the egress proxy.
logic_claims:
  - CLAIM-1FFC126CD441F9F0
semantic_clauses:
  - Plugin egress must go through the egress proxy.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7628befadc2be228ce35a703617adee9b22fb1bb235844cda566bfb1ce1b8f3f
semantic_inventory:
  - claim_key: CLAIM-1FFC126CD441F9F0
    claim_text: Plugin egress must go through the egress proxy
    role: normative
    status: modeled
    span:
      start: 0
      end: 46
origin:
  kind: import
  ref: design-decisions:D-32#1
  recorded_at: '2026-10-27T20:12:51.772Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-8CBFBD6A12E0865B
type: req
---
Plugin egress must go through the egress proxy.

## Source

> Plugin egress goes through an egress proxy built on Stripe's Smokescreen, embedded as a Go library. The operator compiles each `PluginGrant`'s hostnames into that plugin's allow-list. Plugins identify themselves with their projected ServiceAccount token, through a custom role function. NetworkPolicy lets plugins reach only the core and the proxy, with no DNS, and the proxy refuses private and internal address ranges. Plugins reach external hosts through `HTTPS_PROXY`.

Source: OpenFief author decisions (DESIGN.md section 10) - D-32#1
