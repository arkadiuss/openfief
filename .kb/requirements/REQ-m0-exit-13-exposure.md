---
title: The default installation must create no route, no LoadBalancer or NodePort Service and no tunnel, and with the HTTPRoute component the core, and through it the UI, must be reachable over HTTPS on the owner's hostname through their Gateway.
status: open
text_ref: design-decisions:D-200#13
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-13
  - lane:strict
semantic_text: The default installation must create no route, no LoadBalancer or NodePort Service and no tunnel, and with the HTTPRoute component the core, and through it the UI, must be reachable over HTTPS on the owner's hostname through their Gateway.
logic_claims:
  - CLAIM-EAB79D0283387F2F
semantic_clauses:
  - The default installation must create no route, no LoadBalancer or NodePort Service and no tunnel, and with the HTTPRoute component the core, and through it the UI, must be reachable over HTTPS on the owner's hostname through their Gateway.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1be2382f4f6fbe77652c0ae319ffe6a144c34ef96f2b55ee52ca025ef1262f79
semantic_inventory:
  - claim_key: CLAIM-EAB79D0283387F2F
    claim_text: The default installation must create no route, no LoadBalancer or NodePort Service and no tunnel, and with the HTTPRoute component the core, and through it the UI, must be reachable over HTTPS on the owner's hostname through their Gateway
    role: normative
    status: modeled
    span:
      start: 0
      end: 238
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#13
  recorded_at: '2026-11-07T03:32:24.461Z'
id: REQ-m0-exit-13-exposure
type: req
---
The default installation must create no route, no LoadBalancer or NodePort Service and no tunnel, and with the HTTPRoute component the core, and through it the UI, must be reachable over HTTPS on the owner's hostname through their Gateway.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-13.

## Source

> **Exposure.** The default installation creates no route, no LoadBalancer or NodePort Service and no tunnel. With the HTTPRoute component, the core, and through it the UI, is reachable over HTTPS on the owner's hostname through their Gateway.

design-decisions:D-200#13 (docs/DESIGN.md, D-200, criterion 13)
