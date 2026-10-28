---
title: A plugin with granted egress hosts must get HTTPS_PROXY pointing at the egress proxy's cluster address and an egress allow-list of exactly those hosts, and a plugin without any must get neither a proxy setting nor a proxy route.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: A plugin with granted egress hosts must get HTTPS_PROXY pointing at the egress proxy's cluster address and an egress allow-list of exactly those hosts, and a plugin without any must get neither a proxy setting nor a proxy route.
logic_claims:
  - CLAIM-B8823E04FCA2649D
semantic_clauses:
  - A plugin with granted egress hosts must get HTTPS_PROXY pointing at the egress proxy's cluster address and an egress allow-list of exactly those hosts, and a plugin without any must get neither a proxy setting nor a proxy route.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3217802787c12e7d392aa84cd925b4a58eedea7dc61cb4137b7de6719efd935e
semantic_inventory:
  - claim_key: CLAIM-B8823E04FCA2649D
    claim_text: A plugin with granted egress hosts must get HTTPS_PROXY pointing at the egress proxy's cluster address and an egress allow-list of exactly those hosts, and a plugin without any must get neither a proxy setting nor a proxy route
    role: normative
    status: modeled
    span:
      start: 0
      end: 227
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:41.817Z'
id: REQ-plugins-proxy-env-and-allow-list
type: req
---
A plugin with granted egress hosts must get HTTPS_PROXY pointing at the egress proxy's cluster address and an egress allow-list of exactly those hosts, and a plugin without any must get neither a proxy setting nor a proxy route.

## Context

D-32 has plugins reach external hosts through HTTPS_PROXY and the operator compile each grant's hostnames into the plugin's allow-list. Plugins have no DNS, so HTTPS_PROXY uses the proxy Service's ClusterIP, and NO_PROXY names the core so core calls never go through the proxy. The allow-list is keyed by plugin name and sorted for deterministic output. Refines REQ-AUTO-8CBFBD6A12E0865B and REQ-AUTO-48C62B395F846700. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> Plugins reach external hosts through `HTTPS_PROXY`.

docs/DESIGN.md D-32
