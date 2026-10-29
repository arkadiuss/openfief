---
title: The core must be the only entry point for browsers, so that kubectl port-forward or the owner's own Ingress needs to reach the core alone.
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
  recorded_at: '2026-10-29T08:11:24.875Z'
semantic_text: The core must be the only entry point for browsers, so that kubectl port-forward or the owner's own Ingress needs to reach the core alone.
logic_claims:
  - CLAIM-A50C05C302FDC229
semantic_clauses:
  - The core must be the only entry point for browsers, so that kubectl port-forward or the owner's own Ingress needs to reach the core alone.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: eeaabdb7276b82945be058d58b058b3e1ef8a42a127d6a8a0e4c776450a95d6e
semantic_inventory:
  - claim_key: CLAIM-A50C05C302FDC229
    claim_text: The core must be the only entry point for browsers, so that kubectl port-forward or the owner's own Ingress needs to reach the core alone
    role: normative
    status: modeled
    span:
      start: 0
      end: 137
id: REQ-core-browser-entry-point
type: req
---
The core must be the only entry point for browsers, so that kubectl port-forward or the owner's own Ingress needs to reach the core alone.

## Context

D-81 answers the implementation's question about view plugins that browsers reach: D-59 says only the core may call a plugin (REQ-plugins-core-only-caller), but the built-in UI is a view plugin that browsers reached through the demo's Ingress, which the demo allowed with a demo-only NetworkPolicy from Traefik. D-81 keeps D-59 literally true behind any Ingress. D-28 ships no Ingress (REQ-AUTO-3E7CF23DAD42DA83) and makes kubectl port-forward the default way in (REQ-AUTO-0D5A3B94C46E88DF). The implementation's REQ-ui-plugin-packaging and its scenario SCEN-e2e-ui-login-through-traefik still serve the UI to browsers through the demo's Ingress directly. The author recorded D-81 and D-82 after the design session that answered the implementation's questions.

## Source

> The core is the only entry point for browsers. It reverse-proxies the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so only the core ever calls a plugin (D-59), and `kubectl port-forward` or the owner's own Ingress (D-28) needs to reach the core alone. A view plugin's manifest declares the path prefix it serves to browsers, and the grant approves it.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-81#1
