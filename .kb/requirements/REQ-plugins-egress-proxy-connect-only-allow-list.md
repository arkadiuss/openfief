---
title: The egress proxy must open only HTTP CONNECT tunnels, directly and to a host and port on the identified plugin's egress allow-list, and refuse plain HTTP proxying and requests to chain through another proxy.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: The egress proxy must open only HTTP CONNECT tunnels, directly and to a host and port on the identified plugin's egress allow-list, and refuse plain HTTP proxying and requests to chain through another proxy.
logic_claims:
  - CLAIM-1649B811359E478A
semantic_clauses:
  - The egress proxy must open only HTTP CONNECT tunnels, directly and to a host and port on the identified plugin's egress allow-list, and refuse plain HTTP proxying and requests to chain through another proxy.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a32c2ae7373063c771a2055ebd205ba3584e9ca416f87052843ee98ec94c6add
semantic_inventory:
  - claim_key: CLAIM-1649B811359E478A
    claim_text: The egress proxy must open only HTTP CONNECT tunnels, directly and to a host and port on the identified plugin's egress allow-list, and refuse plain HTTP proxying and requests to chain through another proxy
    role: normative
    status: modeled
    span:
      start: 0
      end: 206
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:24.597Z'
id: REQ-plugins-egress-proxy-connect-only-allow-list
type: req
---
The egress proxy must open only HTTP CONNECT tunnels, directly and to a host and port on the identified plugin's egress allow-list, and refuse plain HTTP proxying and requests to chain through another proxy.

## Context

D-32 sends plugin egress through the proxy with HTTPS_PROXY, so plugins only need CONNECT and TLS stays end to end. Smokescreen also supports plain HTTP proxying and an X-Upstream-Https-Proxy header that chains to another proxy; either would let a plugin reach or send data around its allow-list, so both are refused. For the same reason the proxy refuses to start when its own HTTPS_PROXY is set, which Smokescreen's goproxy would otherwise chain through. Refines REQ-AUTO-48C62B395F846700 and REQ-AUTO-8CBFBD6A12E0865B. Modeled by Claude Code for the M0 egress proxy task (internal/plugins/app, smokescreen and kube adapters, cmd/egress-proxy); the author has not reviewed it yet.

## Source

> Plugin egress goes through an egress proxy built on Stripe's Smokescreen, embedded as a Go library.

docs/DESIGN.md D-32
