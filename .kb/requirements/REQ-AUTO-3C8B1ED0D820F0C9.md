---
title: Plugin network access must be limited to the core and the egress proxy, with no DNS.
status: closed
text_ref: design-decisions:D-32#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-32-4
  - lane:strict
semantic_text: Plugin network access must be limited to the core and the egress proxy, with no DNS.
logic_claims:
  - CLAIM-195BD9964561B4FA
semantic_clauses:
  - Plugin network access must be limited to the core and the egress proxy, with no DNS.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3ba35871346382bac5ec915def37e3968ee01ad8b6b2e10676a394ae22a4c4a8
semantic_inventory:
  - claim_key: CLAIM-195BD9964561B4FA
    claim_text: Plugin network access must be limited to the core and the egress proxy, with no DNS
    role: normative
    status: modeled
    span:
      start: 0
      end: 83
origin:
  kind: import
  ref: design-decisions:D-32#4
  recorded_at: '2026-10-27T20:12:54.004Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-3C8B1ED0D820F0C9
type: req
---
Plugin network access must be limited to the core and the egress proxy, with no DNS.

## Source

> Plugin egress goes through an egress proxy built on Stripe's Smokescreen, embedded as a Go library. The operator compiles each `PluginGrant`'s hostnames into that plugin's allow-list. Plugins identify themselves with their projected ServiceAccount token, through a custom role function. NetworkPolicy lets plugins reach only the core and the proxy, with no DNS, and the proxy refuses private and internal address ranges. Plugins reach external hosts through `HTTPS_PROXY`.

Source: OpenFief author decisions (DESIGN.md section 10) - D-32#4
