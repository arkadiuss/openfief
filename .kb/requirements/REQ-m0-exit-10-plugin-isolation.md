---
title: Each plugin must run in its own pod, the XTB parser must have no egress at all, a test plugin's request through the egress proxy for a host outside its grant must be refused, and attempts to bypass the proxy, a direct connection or a DNS lookup, must fail.
status: open
text_ref: design-decisions:D-200#10
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-10
  - lane:strict
semantic_text: Each plugin must run in its own pod, the XTB parser must have no egress at all, a test plugin's request through the egress proxy for a host outside its grant must be refused, and attempts to bypass the proxy, a direct connection or a DNS lookup, must fail.
logic_claims:
  - CLAIM-872F0D49928AA93E
semantic_clauses:
  - Each plugin must run in its own pod, the XTB parser must have no egress at all, a test plugin's request through the egress proxy for a host outside its grant must be refused, and attempts to bypass the proxy, a direct connection or a DNS lookup, must fail.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b22c3f95f82634aaeeff6e8d52402b1730458833c71b966525c558c5e0954514
semantic_inventory:
  - claim_key: CLAIM-872F0D49928AA93E
    claim_text: Each plugin must run in its own pod, the XTB parser must have no egress at all, a test plugin's request through the egress proxy for a host outside its grant must be refused, and attempts to bypass the proxy, a direct connection or a DNS lookup, must fail
    role: normative
    status: modeled
    span:
      start: 0
      end: 255
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#10
  recorded_at: '2026-11-07T03:32:09.988Z'
id: REQ-m0-exit-10-plugin-isolation
type: req
---
Each plugin must run in its own pod, the XTB parser must have no egress at all, a test plugin's request through the egress proxy for a host outside its grant must be refused, and attempts to bypass the proxy, a direct connection or a DNS lookup, must fail.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-10.

## Source

> **Plugin isolation.** Each plugin runs in its own pod. The XTB parser has no egress at all. A test plugin's request through the proxy for a host outside its grant is refused, and attempts to bypass the proxy (a direct connection, a DNS lookup) fail.

design-decisions:D-200#10 (docs/DESIGN.md, D-200, criterion 10)

## Why this changed

An earlier agent modeled criterion 10 as the default-firmness requirement REQ-e2e-plugin-isolation before D-200's criteria were imported as firm requirements. D-200 is the author's decision, so this firm requirement, stated from criterion 10 itself, replaces it, and the end-to-end scenarios that verified REQ-e2e-plugin-isolation now specify this requirement too.
