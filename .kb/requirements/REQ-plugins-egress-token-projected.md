---
title: A plugin with granted egress hosts must get a second projected ServiceAccount token with audience openfief-egress, mounted read-only and named by OPENFIEF_EGRESS_TOKEN_FILE, and a plugin without granted egress must get no such token.
status: open
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
semantic_text: A plugin with granted egress hosts must get a second projected ServiceAccount token with audience openfief-egress, mounted read-only and named by OPENFIEF_EGRESS_TOKEN_FILE, and a plugin without granted egress must get no such token.
logic_claims:
  - CLAIM-B15313923CACC61E
semantic_clauses:
  - A plugin with granted egress hosts must get a second projected ServiceAccount token with audience openfief-egress, mounted read-only and named by OPENFIEF_EGRESS_TOKEN_FILE, and a plugin without granted egress must get no such token.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 986b4346259c54d5569bb968691dd7da78d15981bd8d00c32ab601225bc05a6f
semantic_inventory:
  - claim_key: CLAIM-B15313923CACC61E
    claim_text: A plugin with granted egress hosts must get a second projected ServiceAccount token with audience openfief-egress, mounted read-only and named by OPENFIEF_EGRESS_TOKEN_FILE, and a plugin without granted egress must get no such token
    role: normative
    status: modeled
    span:
      start: 0
      end: 232
origin:
  kind: agent
  recorded_at: '2026-10-28T08:37:45.738Z'
id: REQ-plugins-egress-token-projected
type: req
---
A plugin with granted egress hosts must get a second projected ServiceAccount token with audience openfief-egress, mounted read-only and named by OPENFIEF_EGRESS_TOKEN_FILE, and a plugin without granted egress must get no such token.

## Context

D-63 has the egress proxy check a plugin's projected token with TokenReview for audience openfief-egress, while D-27 gives plugins a token with audience openfief-core for the core. A separate audience means the proxy never receives a token the core would accept, so it cannot replay one. The SDK reads the file named by OPENFIEF_EGRESS_TOKEN_FILE on each proxied request (D-63) and puts the token in the proxy URL as the password. A plugin without egress has no proxy route, so it gets no egress token either. The env name was agreed with the contracts agent through the orchestrator. Modeled by Claude Code for the M0 egress proxy v2 task (D-63 conformance); the author has not reviewed it yet.

## Source

> A plugin sends its projected token as the password in the proxy URL, which Go's and Node's HTTP clients send as a `Proxy-Authorization` header; the role function checks it with Kubernetes' TokenReview API (audience `openfief-egress`) and caches the result.

docs/DESIGN.md D-63
