---
title: The egress proxy must never run without an allow-list and must refuse every destination for a plugin that has none.
status: open
text_ref: design-decisions:D-63#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-63-1
  - lane:strict
semantic_text: The egress proxy must never run without an allow-list and must refuse every destination for a plugin that has none.
logic_claims:
  - CLAIM-84EE04E0241DDA78
semantic_clauses:
  - The egress proxy must never run without an allow-list and must refuse every destination for a plugin that has none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1a2f177c183aa9b17befb68aa2bdd3616cd9db3199eb3e6845961d315d92d6de
semantic_inventory:
  - claim_key: CLAIM-84EE04E0241DDA78
    claim_text: The egress proxy must never run without an allow-list and must refuse every destination for a plugin that has none
    role: normative
    status: modeled
    span:
      start: 0
      end: 114
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-63#1
  recorded_at: '2026-10-28T06:18:45.409Z'
id: REQ-plugins-egress-fail-closed
type: req
---
The egress proxy must never run without an allow-list and must refuse every destination for a plugin that has none.

## Context

Each plugin runs out of the core's process, with its own network policy, secrets and identity, so a broken or hostile plugin can reach only what it was granted. Standard NetworkPolicy matches IP addresses, not hostnames, so outbound traffic goes through a proxy that checks hostnames.

## Source

> The egress proxy embeds Smokescreen through `BuildProxy` inside OpenFief's own HTTP server; `StartWithConfig` would exit the process on errors and take over signals. Smokescreen allows every destination when no allow-list is set, so the proxy never runs without one and refuses all destinations for a plugin that has none. Compiled allow-lists are swapped atomically. A plugin sends its projected token as the password in the proxy URL, which Go's and Node's HTTP clients send as a `Proxy-Authorization` header; the role function checks it with Kubernetes' TokenReview API (audience `openfief-egress`) and caches the result. Projected tokens rotate, so the SDK builds the proxy URL (D-32's `HTTPS_PROXY`) from the token file on each request, using the proxy's ClusterIP because plugins have no DNS. Smokescreen is pinned to a commit after v0.1.0 until a release carries its later hardening fixes.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-63#1
