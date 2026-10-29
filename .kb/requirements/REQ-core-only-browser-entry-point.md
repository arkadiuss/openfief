---
title: The core must be the only entry point for browsers, so that whatever route the owner adds needs to reach the core alone.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-81-1
  - lane:strict
text_ref: design-decisions:D-81#1
origin:
  kind: import
  ref: design-decisions:D-81#1
  recorded_at: '2026-10-29T15:05:03.288Z'
semantic_text: The core must be the only entry point for browsers, so that whatever route the owner adds needs to reach the core alone.
logic_claims:
  - CLAIM-2E9629BDE3222F6C
semantic_clauses:
  - The core must be the only entry point for browsers, so that whatever route the owner adds needs to reach the core alone.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c2dd5dce8ba1d98e0c4d175656862a46db27256091994f428ed2ee21578e4a7a
semantic_inventory:
  - claim_key: CLAIM-2E9629BDE3222F6C
    claim_text: The core must be the only entry point for browsers, so that whatever route the owner adds needs to reach the core alone
    role: normative
    status: modeled
    span:
      start: 0
      end: 119
id: REQ-core-only-browser-entry-point
type: req
---
The core must be the only entry point for browsers, so that whatever route the owner adds needs to reach the core alone.

## Why this changed

REQ-core-browser-entry-point named `kubectl port-forward` or the owner's own Ingress as the way that needs to reach the core alone, following D-81's first wording. In the docs session of 2026-10-29 the author changed D-28: `kubectl port-forward` is no longer the default way in, OpenFief ships the core's ClusterIP Service plus an optional Gateway API HTTPRoute component (REQ-platform-httproute-component), and an Ingress or a port-forward is the owner's choice. D-81 was reworded to match: whatever route the owner adds needs to reach the core alone, and this requirement states that.

## Context

D-81 answers the implementation's question about view plugins that browsers reach: D-59 says only the core may call a plugin (REQ-plugins-core-only-caller), but the built-in UI is a view plugin that browsers reached through the demo's Ingress, which the demo allowed with a demo-only NetworkPolicy from Traefik. D-81 keeps D-59 literally true behind any route. OpenFief exposes nothing by default (REQ-platform-exposes-nothing-by-default). The author recorded D-81 and D-82 after the design session that answered the implementation's questions, and reworded D-81 when reviewing the docs, which over-emphasized `kubectl port-forward`.

## Source

> The core is the only entry point for browsers. It reverse-proxies the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so only the core ever calls a plugin (D-59), and whatever route the owner adds (D-28) needs to reach the core alone. A view plugin's manifest declares the path prefix it serves to browsers, and the grant approves it.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-81#1
