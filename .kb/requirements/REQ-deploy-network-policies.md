---
title: 'The database, the egress proxy and the operator must each have a NetworkPolicy admitting only their callers: the database the core, its own instances and CloudNativePG''s operator; the egress proxy plugin pods, with its own egress leaving the private, shared, link-local and loopback ranges out; the operator nothing.'
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: 'The database, the egress proxy and the operator must each have a NetworkPolicy admitting only their callers: the database the core, its own instances and CloudNativePG''s operator; the egress proxy plugin pods, with its own egress leaving the private, shared, link-local and loopback ranges out; the operator nothing.'
logic_claims:
  - CLAIM-600CFF1FB23ABA9F
semantic_clauses:
  - 'The database, the egress proxy and the operator must each have a NetworkPolicy admitting only their callers: the database the core, its own instances and CloudNativePG''s operator; the egress proxy plugin pods, with its own egress leaving the private, shared, link-local and loopback ranges out; the operator nothing.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3a82d8568abee297c82c3dd6e281bc2663139a44536c49fcaf749faa2c36f52b
semantic_inventory:
  - claim_key: CLAIM-600CFF1FB23ABA9F
    claim_text: 'The database, the egress proxy and the operator must each have a NetworkPolicy admitting only their callers: the database the core, its own instances and CloudNativePG''s operator; the egress proxy plugin pods, with its own egress leaving the private, shared, link-local and loopback ranges out; the operator nothing'
    role: normative
    status: modeled
    span:
      start: 0
      end: 315
origin:
  kind: agent
  recorded_at: '2026-11-03T00:10:47.950Z'
id: REQ-deploy-network-policies
type: req
---
The database, the egress proxy and the operator must each have a NetworkPolicy admitting only their callers: the database the core, its own instances and CloudNativePG's operator; the egress proxy plugin pods, with its own egress leaving the private, shared, link-local and loopback ranges out; the operator nothing.

## Context

Without them any pod in the cluster, a compromised plugin included, could open a connection to the database, use the proxy, or reach the operator's endpoints. The database's policy is in deploy/core, which every install applies, because deploy/database must hold only the Cluster. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p1-isolation-net), from the security review the author put before the alpha tag. No firm requirement for this part of D-184 existed in kibi on this branch when it was built.

## Source

> The database, the egress proxy and the operator each have a NetworkPolicy admitting only their callers. Pages the core proxies from a view plugin carry security headers the core sets: a content security policy, framing denied, no content sniffing and no referrer.

docs/DESIGN.md#D-184
