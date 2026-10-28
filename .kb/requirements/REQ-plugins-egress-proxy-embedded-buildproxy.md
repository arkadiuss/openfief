---
title: The egress proxy must serve Smokescreen's BuildProxy handler from OpenFief's own HTTP server, never let Smokescreen exit the process or install signal handlers, and on shutdown stop accepting, let open tunnels drain for a bounded time and close the rest.
status: open
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
semantic_text: The egress proxy must serve Smokescreen's BuildProxy handler from OpenFief's own HTTP server, never let Smokescreen exit the process or install signal handlers, and on shutdown stop accepting, let open tunnels drain for a bounded time and close the rest.
logic_claims:
  - CLAIM-43B2694FD83DD205
semantic_clauses:
  - The egress proxy must serve Smokescreen's BuildProxy handler from OpenFief's own HTTP server, never let Smokescreen exit the process or install signal handlers, and on shutdown stop accepting, let open tunnels drain for a bounded time and close the rest.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4aa40d02914d76131c9e6de3090342b43a15c5f353500fc5da6409c3a6702cf8
semantic_inventory:
  - claim_key: CLAIM-43B2694FD83DD205
    claim_text: The egress proxy must serve Smokescreen's BuildProxy handler from OpenFief's own HTTP server, never let Smokescreen exit the process or install signal handlers, and on shutdown stop accepting, let open tunnels drain for a bounded time and close the rest
    role: normative
    status: modeled
    span:
      start: 0
      end: 253
origin:
  kind: agent
  recorded_at: '2026-10-28T08:46:58.110Z'
id: REQ-plugins-egress-proxy-embedded-buildproxy
type: req
---
The egress proxy must serve Smokescreen's BuildProxy handler from OpenFief's own HTTP server, never let Smokescreen exit the process or install signal handlers, and on shutdown stop accepting, let open tunnels drain for a bounded time and close the rest.

## Context

D-63 picks BuildProxy over StartWithConfig because StartWithConfig exits the process on errors and takes over SIGTERM, SIGHUP and SIGUSR2. Embedding the handler means the proxy repeats the setup StartWithConfig did: config validation, the self-connection guard on every port, connection tracking, server timeouts and an idle timeout so long tunnels live while active. The 25 s drain fits Kubernetes' default 30 s grace. Modeled by Claude Code for the M0 egress proxy v2 task (D-63 conformance); the author has not reviewed it yet.

## Source

> The egress proxy embeds Smokescreen through `BuildProxy` inside OpenFief's own HTTP server; `StartWithConfig` would exit the process on errors and take over signals. Smokescreen allows every destination when no allow-list is set, so the proxy never runs without one and refuses all destinations for a plugin that has none. Compiled allow-lists are swapped atomically. A plugin sends its projected token as the password in the proxy URL, which Go's and Node's HTTP clients send as a `Proxy-Authorization` header; the role function checks it with Kubernetes' TokenReview API (audience `openfief-egress`) and caches the result.

docs/DESIGN.md D-63
