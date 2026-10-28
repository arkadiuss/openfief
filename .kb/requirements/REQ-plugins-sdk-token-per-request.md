---
title: The SDK must build the proxy URL from the plugin's projected token file on each proxied request, using the proxy's ClusterIP.
status: open
text_ref: design-decisions:D-63#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-63-2
  - lane:strict
semantic_text: The SDK must build the proxy URL from the plugin's projected token file on each proxied request, using the proxy's ClusterIP.
logic_claims:
  - CLAIM-606B986E53F9C992
semantic_clauses:
  - The SDK must build the proxy URL from the plugin's projected token file on each proxied request, using the proxy's ClusterIP.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 82d21c7790f5c80200e3ef7dc76e7830b652d63e536134c22a42a46cd21f949d
semantic_inventory:
  - claim_key: CLAIM-606B986E53F9C992
    claim_text: The SDK must build the proxy URL from the plugin's projected token file on each proxied request, using the proxy's ClusterIP
    role: normative
    status: modeled
    span:
      start: 0
      end: 124
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-63#2
  recorded_at: '2026-10-28T06:18:49.608Z'
id: REQ-plugins-sdk-token-per-request
type: req
---
The SDK must build the proxy URL from the plugin's projected token file on each proxied request, using the proxy's ClusterIP.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. Standard NetworkPolicy matches IP addresses, not hostnames, so outbound traffic goes through a proxy that checks hostnames.

## Source

> The egress proxy embeds Smokescreen through `BuildProxy` inside OpenFief's own HTTP server; `StartWithConfig` would exit the process on errors and take over signals. Smokescreen allows every destination when no allow-list is set, so the proxy never runs without one and refuses all destinations for a plugin that has none. Compiled allow-lists are swapped atomically. A plugin sends its projected token as the password in the proxy URL, which Go's and Node's HTTP clients send as a `Proxy-Authorization` header; the role function checks it with Kubernetes' TokenReview API (audience `openfief-egress`) and caches the result. Projected tokens rotate, so the SDK builds the proxy URL (D-32's `HTTPS_PROXY`) from the token file on each request, using the proxy's ClusterIP because plugins have no DNS. Smokescreen is pinned to a commit after v0.1.0 until a release carries its later hardening fixes.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-63#2
