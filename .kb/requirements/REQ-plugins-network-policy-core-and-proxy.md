---
title: A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, and never DNS.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, and never DNS.
logic_claims:
  - CLAIM-A8DCAEEA129FE717
semantic_clauses:
  - A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, and never DNS.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 391c33743dd036f2bf052b7e88af29c38d6e71f722453e03dc417550eb1211ef
semantic_inventory:
  - claim_key: CLAIM-A8DCAEEA129FE717
    claim_text: A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, and never DNS
    role: normative
    status: modeled
    span:
      start: 0
      end: 184
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:26.996Z'
id: REQ-plugins-network-policy-core-and-proxy
type: req
---
A plugin's compiled NetworkPolicy must allow ingress only from the core and egress only to the core, plus the egress proxy only when the plugin has a granted egress host, and never DNS.

## Context

D-32 lets plugins reach only the core and the proxy, with no DNS. Our reading: a plugin with no granted egress host (the XTB parser, section 3.5) gets no route to the proxy at all; ingress is allowed only from the core on the plugin's Connect port, because the core is the only party D-11 has talking to plugins; the built-in UI behind a user's own Ingress (D-28) needs a policy the user adds. Ports in rules are pod ports, which NetworkPolicy matches. Refines REQ-AUTO-3C8B1ED0D820F0C9. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> NetworkPolicy lets plugins reach only the core and the proxy, with no DNS

docs/DESIGN.md D-32
