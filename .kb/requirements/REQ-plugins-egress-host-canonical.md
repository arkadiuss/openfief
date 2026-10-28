---
title: An egress host in a manifest or grant must be a canonical DNS hostname with a port, and an IP literal, a wildcard or a single-label name must be rejected.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: An egress host in a manifest or grant must be a canonical DNS hostname with a port, and an IP literal, a wildcard or a single-label name must be rejected.
logic_claims:
  - CLAIM-C5B5420AB26564BA
semantic_clauses:
  - An egress host in a manifest or grant must be a canonical DNS hostname with a port, and an IP literal, a wildcard or a single-label name must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a5a28557bca2f2adb3ab713a5c2c23f8661af3b99cbe699d470a7300bef29631
semantic_inventory:
  - claim_key: CLAIM-C5B5420AB26564BA
    claim_text: An egress host in a manifest or grant must be a canonical DNS hostname with a port, and an IP literal, a wildcard or a single-label name must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 153
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:23.165Z'
id: REQ-plugins-egress-host-canonical
type: req
---
An egress host in a manifest or grant must be a canonical DNS hostname with a port, and an IP literal, a wildcard or a single-label name must be rejected.

## Context

D-32 has the egress proxy allow each plugin only the hostnames in its grant, and the proxy also refuses private ranges. An IP literal would skip both the hostname check and the proxy's own resolution; a wildcard would grant hosts the owner never saw; a single-label name would resolve through cluster search domains. Hosts are lower-case ASCII with IDNA A-labels so that one name has one spelling in grants, allow-lists and logs. The port defaults to 443 because plugins reach external hosts through HTTPS_PROXY. Refines REQ-AUTO-48C62B395F846700. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> The operator compiles each `PluginGrant`'s hostnames into that plugin's allow-list.

docs/DESIGN.md D-32
