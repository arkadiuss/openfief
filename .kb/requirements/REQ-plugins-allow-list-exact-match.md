---
title: The egress proxy must allow a destination only when its host and port exactly match an entry in the plugin's egress allow-list, ignoring only ASCII case and one trailing dot.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: The egress proxy must allow a destination only when its host and port exactly match an entry in the plugin's egress allow-list, ignoring only ASCII case and one trailing dot.
logic_claims:
  - CLAIM-CF486E213DCBB024
semantic_clauses:
  - The egress proxy must allow a destination only when its host and port exactly match an entry in the plugin's egress allow-list, ignoring only ASCII case and one trailing dot.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2abce7197635db0cd292af1accee55e022da44964779270c7002323605420e7b
semantic_inventory:
  - claim_key: CLAIM-CF486E213DCBB024
    claim_text: The egress proxy must allow a destination only when its host and port exactly match an entry in the plugin's egress allow-list, ignoring only ASCII case and one trailing dot
    role: normative
    status: modeled
    span:
      start: 0
      end: 173
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:48.594Z'
id: REQ-plugins-allow-list-exact-match
type: req
---
The egress proxy must allow a destination only when its host and port exactly match an entry in the plugin's egress allow-list, ignoring only ASCII case and one trailing dot.

## Context

D-32 allows each plugin only the hostnames in its grant. Exact matching means a subdomain, a parent domain or a longer name ending in a granted one is refused. DNS names are case-insensitive and may carry a final dot, so those two differences are ignored; Unicode case folding is not, because characters such as the Kelvin sign fold to ASCII letters and would let a different name match. Refines REQ-AUTO-48C62B395F846700. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> The operator compiles each `PluginGrant`'s hostnames into that plugin's allow-list.

docs/DESIGN.md D-32
