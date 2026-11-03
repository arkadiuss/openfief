---
title: The egress proxy must refuse, beyond the private and internal ranges, the address of the node it runs on and any ranges the owner names, such as a cluster's public or global IPv6 pod, Service and node ranges, and must not start with a range that doesn't parse.
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: The egress proxy must refuse, beyond the private and internal ranges, the address of the node it runs on and any ranges the owner names, such as a cluster's public or global IPv6 pod, Service and node ranges, and must not start with a range that doesn't parse.
logic_claims:
  - CLAIM-A6920D0DC066FD88
semantic_clauses:
  - The egress proxy must refuse, beyond the private and internal ranges, the address of the node it runs on and any ranges the owner names, such as a cluster's public or global IPv6 pod, Service and node ranges, and must not start with a range that doesn't parse.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 813e508943771d5bfec8303d224ad133150ff782fe71e19bdde721e2f3f4b911
semantic_inventory:
  - claim_key: CLAIM-A6920D0DC066FD88
    claim_text: The egress proxy must refuse, beyond the private and internal ranges, the address of the node it runs on and any ranges the owner names, such as a cluster's public or global IPv6 pod, Service and node ranges, and must not start with a range that doesn't parse
    role: normative
    status: modeled
    span:
      start: 0
      end: 259
origin:
  kind: agent
  recorded_at: '2026-11-03T00:10:54.686Z'
id: REQ-plugins-egress-denies-node-and-cluster-ranges
type: req
---
The egress proxy must refuse, beyond the private and internal ranges, the address of the node it runs on and any ranges the owner names, such as a cluster's public or global IPv6 pod, Service and node ranges, and must not start with a range that doesn't parse.

## Context

A single-node install with a public address, such as the demo, would otherwise let a plugin reach the node's own ports, the API server's and the kubelet's among them, through an allowed name resolving to it. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p1-isolation-net), from the security review the author put before the alpha tag. No firm requirement for this part of D-184 existed in kibi on this branch when it was built.

## Source

> The database, the egress proxy and the operator each have a NetworkPolicy admitting only their callers. Pages the core proxies from a view plugin carry security headers the core sets: a content security policy, framing denied, no content sniffing and no referrer.

docs/DESIGN.md#D-184
