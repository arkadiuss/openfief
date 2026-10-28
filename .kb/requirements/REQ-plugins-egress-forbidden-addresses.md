---
title: The egress proxy must refuse to dial any address in a private, loopback, link-local, carrier-grade NAT, unique-local, multicast, unspecified, documentation or reserved range, including IPv4 addresses embedded in IPv6 addresses.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: The egress proxy must refuse to dial any address in a private, loopback, link-local, carrier-grade NAT, unique-local, multicast, unspecified, documentation or reserved range, including IPv4 addresses embedded in IPv6 addresses.
logic_claims:
  - CLAIM-97B067D6FA89738F
semantic_clauses:
  - The egress proxy must refuse to dial any address in a private, loopback, link-local, carrier-grade NAT, unique-local, multicast, unspecified, documentation or reserved range, including IPv4 addresses embedded in IPv6 addresses.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a9c7fa38b819deb87fae33cb5488130ffda878db6cf93f25889644cc16db58e0
semantic_inventory:
  - claim_key: CLAIM-97B067D6FA89738F
    claim_text: The egress proxy must refuse to dial any address in a private, loopback, link-local, carrier-grade NAT, unique-local, multicast, unspecified, documentation or reserved range, including IPv4 addresses embedded in IPv6 addresses
    role: normative
    status: modeled
    span:
      start: 0
      end: 226
origin:
  kind: agent
  recorded_at: '2026-10-28T05:32:04.783Z'
id: REQ-plugins-egress-forbidden-addresses
type: req
---
The egress proxy must refuse to dial any address in a private, loopback, link-local, carrier-grade NAT, unique-local, multicast, unspecified, documentation or reserved range, including IPv4 addresses embedded in IPv6 addresses.

## Context

D-32 has the proxy refuse private and internal address ranges, so a granted hostname that resolves inside the cluster, the host network or the cloud metadata service cannot be used to reach it. IPv4 addresses hidden in IPv4-mapped, IPv4-compatible, NAT64 and 6to4 IPv6 addresses are checked as IPv4, and Teredo is refused outright. The check needs net/netip, which domain packages may not import, so it lives in internal/plugins/egress for the proxy adapter. Refines REQ-AUTO-79189DB32149E23F. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> the proxy refuses private and internal address ranges

docs/DESIGN.md D-32
