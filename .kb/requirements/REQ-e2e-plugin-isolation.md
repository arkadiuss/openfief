---
title: On both kind and k3d, each plugin must run in its own pod, the XTB parser must have no egress at all, a granted plugin's request through the egress proxy for a host outside its grant must be refused, and a direct connection or a DNS lookup from a plugin pod must fail.
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: On both kind and k3d, each plugin must run in its own pod, the XTB parser must have no egress at all, a granted plugin's request through the egress proxy for a host outside its grant must be refused, and a direct connection or a DNS lookup from a plugin pod must fail.
logic_claims:
  - CLAIM-FD88A29F704445E2
semantic_clauses:
  - On both kind and k3d, each plugin must run in its own pod, the XTB parser must have no egress at all, a granted plugin's request through the egress proxy for a host outside its grant must be refused, and a direct connection or a DNS lookup from a plugin pod must fail.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5321f5866158f6d36e6262dfa43ccdaeef92071ddb55fd7ded6e6b1039827a55
semantic_inventory:
  - claim_key: CLAIM-FD88A29F704445E2
    claim_text: On both kind and k3d, each plugin must run in its own pod, the XTB parser must have no egress at all, a granted plugin's request through the egress proxy for a host outside its grant must be refused, and a direct connection or a DNS lookup from a plugin pod must fail
    role: normative
    status: modeled
    span:
      start: 0
      end: 267
origin:
  kind: agent
  recorded_at: '2026-11-06T20:49:59.546Z'
id: REQ-e2e-plugin-isolation
type: req
---
On both kind and k3d, each plugin must run in its own pod, the XTB parser must have no egress at all, a granted plugin's request through the egress proxy for a host outside its grant must be refused, and a direct connection or a DNS lookup from a plugin pod must fail.

## Context

D-200 makes criterion 10 of M0's exit an end-to-end scenario on the firm requirements it checks, passing on both kind and k3d. Those are D-32's network limits (core, egress proxy, one approved in-cluster service, no DNS), D-32's allow-list of granted hosts and D-32's egress through the proxy. Firm requirements already implemented and covered by unit tests through direct links get no scenario here, since a scenario would stop those links from counting; this requirement depends on them instead.

## Source

> **Plugin isolation.** Each plugin runs in its own pod. The XTB parser has no egress at all. A test plugin's request through the proxy for a host outside its grant is refused, and attempts to bypass the proxy (a direct connection, a DNS lookup) fail.

design-decisions:D-200#10 (docs/DESIGN.md, D-200, criterion 10)
